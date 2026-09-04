package serve

import (
	"context"
	"errors"
	"fmt"

	"reasonix/internal/agent"
	"reasonix/internal/control"
)

// newSessionWithModelLocked builds a clean controller for a configured default
// model without first publishing a model switch into the session being left.
// The outgoing controller remains live until the replacement is fully built
// and has secured its fresh session lease.
func (s *Server) newSessionWithModelLocked(ctx context.Context, cur control.SessionAPI, ref string) error {
	if err := cur.Snapshot(); err != nil {
		return fmt.Errorf("snapshot current session: %w", err)
	}
	newCtrl, err := s.build(ctx, ref)
	if err != nil {
		return fmt.Errorf("build default model: %w", err)
	}
	// boot.Build returns a clean controller whose transcript path is allocated
	// lazily. A /new transition must publish a concrete route and acquire its
	// lease before replacing the outgoing controller.
	newCtrl.EnsureSessionPath()
	if tag := s.tagFor(newCtrl); tag != nil {
		tag.PrimePath(newCtrl.SessionPath())
	}
	newCtrl.EnableInteractiveApproval()
	newCtrl.SetOnSessionRecovered(s.sessionRecoveryHandler(newCtrl, s.leases))
	newCtrl.SetToolApprovalMode(defaultApprovalMode())
	_ = newCtrl.SetQualityFloor(cur.QualityFloor())
	if err := s.rebindSessionLeaseFor(newCtrl.SessionPath(), newCtrl); err != nil {
		s.closeTaggedController(newCtrl)
		if errors.Is(err, agent.ErrSessionLeaseHeld) {
			return fmt.Errorf("new session: %s", sessionInUseError(err))
		}
		return fmt.Errorf("new session: unable to secure replacement session")
	}
	if err := persistQualityFloorFor(newCtrl); err != nil {
		oldCtrl, _ := cur.(*control.Controller)
		_ = s.rebindSessionLeaseFor(cur.SessionPath(), oldCtrl)
		s.closeTaggedController(newCtrl)
		return fmt.Errorf("persist quality floor: %w", err)
	}

	if !s.publishControllerSwap(cur, newCtrl, newCtrl.SessionPath()) {
		oldCtrl, _ := cur.(*control.Controller)
		if restoreErr := s.rebindSessionLeaseFor(cur.SessionPath(), oldCtrl); restoreErr != nil {
			s.closeTaggedController(newCtrl)
			return fmt.Errorf("new session: active session changed and ownership could not be restored")
		}
		s.closeTaggedController(newCtrl)
		return fmt.Errorf("new session: active session changed during initialization")
	}
	if tag := s.tagFor(newCtrl); tag != nil {
		tag.Activate()
	}
	s.refreshProviderSetup(currentModelRef(newCtrl))
	cur.Close()
	if oldCtrl, ok := cur.(*control.Controller); ok {
		s.forgetSessionTag(oldCtrl)
	}
	return nil
}
