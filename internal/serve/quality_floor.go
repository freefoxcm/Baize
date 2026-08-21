package serve

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"reasonix/internal/agent"
	"reasonix/internal/boot"
	"reasonix/internal/control"
)

// qualityFloorFromMeta restores the session-scoped delivery posture. Older
// sidecars used token_mode/agent_preset; keep reading those values while the
// canonical quality_floor field rolls out.
func qualityFloorFromMeta(sessionPath string) string {
	meta, ok, err := agent.LoadBranchMeta(sessionPath)
	if err != nil || !ok {
		return control.QualityFloorStandard
	}
	if floor := strings.TrimSpace(meta.QualityFloor); floor != "" {
		if normalized, err := control.NormalizeQualityFloor(floor); err == nil {
			return normalized
		}
	}
	if boot.NormalizeTokenMode(meta.TokenMode) == boot.TokenModeDelivery || meta.AgentPreset == boot.AgentPresetDelivery {
		return control.QualityFloorDelivery
	}
	return control.QualityFloorStandard
}

func applySessionQualityFloorFor(cur control.SessionAPI, sessionPath string) {
	if cur == nil {
		return
	}
	_ = cur.SetQualityFloor(qualityFloorFromMeta(sessionPath))
}

// persistQualityFloorFor writes the current floor to the same branch metadata
// used by Desktop. Standard stays the compact zero value while the legacy
// compatibility fields are dual-written for older clients.
func persistQualityFloorFor(cur control.SessionAPI) error {
	if cur == nil || strings.TrimSpace(cur.SessionPath()) == "" {
		return nil
	}
	floor := cur.QualityFloor()
	return agent.UpdateBranchMeta(cur.SessionPath(), false, func(meta *agent.BranchMeta) error {
		meta.QualityFloor = ""
		meta.TokenMode = boot.TokenModeFull
		meta.AgentPreset = ""
		if floor == control.QualityFloorDelivery {
			meta.QualityFloor = control.QualityFloorDelivery
			meta.TokenMode = boot.TokenModeDelivery
			meta.AgentPreset = boot.AgentPresetDelivery
		}
		return nil
	})
}

// qualityFloor changes only the active session's delivery posture. It is a
// between-turn setting: changing it while work is active would make receipts
// within one turn carry different policy floors.
func (s *Server) qualityFloor(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Floor string `json:"floor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}
	body.Floor = strings.ToLower(strings.TrimSpace(body.Floor))
	if body.Floor != control.QualityFloorStandard && body.Floor != control.QualityFloorDelivery {
		http.Error(w, "floor must be standard or delivery", http.StatusBadRequest)
		return
	}

	s.bindMu.Lock()
	defer s.bindMu.Unlock()
	cur := s.ctl()
	if controllerHasActiveRuntimeWork(cur) {
		http.Error(w, "cannot change quality floor while active work or background jobs are running", http.StatusConflict)
		return
	}
	if err := cur.SetQualityFloor(body.Floor); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := persistQualityFloorFor(cur); err != nil {
		slog.Warn("serve: persist quality floor", "err", err)
		http.Error(w, "unable to persist quality floor", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
