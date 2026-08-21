package serve

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"reasonix/internal/agent"
	"reasonix/internal/config"
	"reasonix/internal/control"
)

func (s *Server) newSession(w http.ResponseWriter, r *http.Request) {
	// Session-path-changing entry point: serialize with /resume, /fork, and
	// switchModel so the controller and the lease keeper move together.
	s.bindMu.Lock()
	defer s.bindMu.Unlock()
	cur := s.ctl()
	if controllerHasActiveRuntimeWork(cur) {
		http.Error(w, "cannot start a new session while active work or background jobs are running", http.StatusConflict)
		return
	}
	if cfg, err := config.LoadUserConfigReadOnly(); err == nil {
		ref := strings.TrimSpace(cfg.DefaultModel)
		if ref != "" && ref != currentModelRef(cur) {
			if err := s.newSessionWithModelLocked(r.Context(), cur, ref); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			s.bc.ResetSession()
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	if err := cur.NewSession(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cur.SetToolApprovalMode(defaultApprovalMode()) // fresh session = configured default
	s.bc.ResetSession()
	if err := s.rebindSessionLease(cur.SessionPath()); err != nil {
		http.Error(w, sessionInUseError(err), http.StatusConflict)
		return
	}
	if err := persistQualityFloorFor(cur); err != nil {
		http.Error(w, "persist quality floor: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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
	newCtrl.EnableInteractiveApproval()
	newCtrl.SetOnSessionRecovered(sessionLeaseRecoveryHandler(s.leases))
	newCtrl.SetToolApprovalMode(defaultApprovalMode())
	_ = newCtrl.SetQualityFloor(cur.QualityFloor())
	if err := s.rebindSessionLeaseFor(newCtrl.SessionPath(), newCtrl); err != nil {
		newCtrl.Close()
		if errors.Is(err, agent.ErrSessionLeaseHeld) {
			return fmt.Errorf("new session: %s", sessionInUseError(err))
		}
		return fmt.Errorf("new session: unable to secure replacement session")
	}
	if err := persistQualityFloorFor(newCtrl); err != nil {
		oldCtrl, _ := cur.(*control.Controller)
		_ = s.rebindSessionLeaseFor(cur.SessionPath(), oldCtrl)
		newCtrl.Close()
		return fmt.Errorf("persist quality floor: %w", err)
	}

	s.mu.Lock()
	if s.ctrl != cur {
		s.mu.Unlock()
		oldCtrl, _ := cur.(*control.Controller)
		if restoreErr := s.rebindSessionLeaseFor(cur.SessionPath(), oldCtrl); restoreErr != nil {
			newCtrl.Close()
			return fmt.Errorf("new session: active session changed and ownership could not be restored")
		}
		newCtrl.Close()
		return fmt.Errorf("new session: active session changed during initialization")
	}
	s.ctrl = newCtrl
	s.mu.Unlock()
	s.refreshProviderSetup(currentModelRef(newCtrl))
	cur.Close()
	return nil
}
