package serve

// Per-session tool-approval-mode persistence (store.SessionApprovalMode
// sidecar): a restart or session switch restores what the user last chose,
// mirroring desktop's per-tab persistence; no record falls back to the default.

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"reasonix/internal/config"
	"reasonix/internal/control"
	"reasonix/internal/store"
)

// approvalModeSidecar is the JSON payload of the per-session sidecar.
type approvalModeSidecar struct {
	Mode string `json:"mode"`
}

// ApplyDesktopDefaultApprovalMode gives a fresh Serve controller the desktop
// default without changing an existing runtime choice.
func ApplyDesktopDefaultApprovalMode(ctrl control.SessionAPI) {
	cfg, err := config.Load()
	if err == nil && cfg.Desktop.DefaultToolApprovalMode != "" {
		ctrl.SetToolApprovalMode(cfg.DesktopDefaultToolApprovalMode())
	}
}

// readSessionApprovalMode returns the persisted mode for a session path, or ""
// when the sidecar is missing or unreadable.
func readSessionApprovalMode(sessionPath string) string {
	raw, err := os.ReadFile(store.SessionApprovalMode(sessionPath))
	if err != nil {
		return ""
	}
	var sc approvalModeSidecar
	if err := json.Unmarshal(raw, &sc); err != nil || sc.Mode == "" {
		return ""
	}
	return config.NormalizeToolApprovalMode(sc.Mode)
}

// writeSessionApprovalMode persists the current mode for a session path via a
// same-directory temp file so a concurrent reader never sees partial JSON.
func writeSessionApprovalMode(sessionPath, mode string) error {
	sidecar := store.SessionApprovalMode(sessionPath)
	if sidecar == "" {
		return nil
	}
	raw, err := json.Marshal(approvalModeSidecar{Mode: mode})
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(sidecar), filepath.Base(sidecar)+".tmp-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	if _, err := f.Write(raw); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, sidecar)
}

// defaultApprovalMode is the posture serve starts with and that sessions
// without a persisted record fall back to. It mirrors
// ApplyDesktopDefaultApprovalMode: config.Load first, built-in default as
// fallback.
func defaultApprovalMode() string {
	cfg, err := config.Load()
	if err != nil {
		return config.Default().DesktopDefaultToolApprovalMode()
	}
	return cfg.DesktopDefaultToolApprovalMode()
}

// applySessionApprovalMode restores a session's persisted mode on the current
// controller, falling back to the default when the session has no record yet.
func (s *Server) applySessionApprovalMode(sessionPath string) {
	applySessionApprovalModeFor(s.ctl(), sessionPath)
}

// applySessionApprovalModeFor restores a session's persisted mode on cur. The
// caller passes one ctl snapshot so the mode lands on the controller that owns
// the path the sidecar was read from.
func applySessionApprovalModeFor(cur control.SessionAPI, sessionPath string) {
	mode := readSessionApprovalMode(sessionPath)
	if mode == "" {
		mode = defaultApprovalMode()
	}
	cur.SetToolApprovalMode(mode)
}

// persistApprovalMode writes the current posture to the active session so a
// restart or a later switch back restores it.
func (s *Server) persistApprovalMode() {
	persistApprovalModeFor(s.ctl())
}

// persistApprovalModeFor writes cur's posture to cur's own session. The caller
// passes one ctl snapshot so mode and path can never split across a rebuild.
func persistApprovalModeFor(cur control.SessionAPI) {
	if p := cur.SessionPath(); p != "" {
		if err := writeSessionApprovalMode(p, cur.ToolApprovalMode()); err != nil {
			slog.Warn("serve: persist approval mode", "err", err)
		}
	}
}

// switchSession runs a /switch <ref> and applies the target session's posture;
// the controller's own Submit path would switch without notifying serve.
func (s *Server) switchSession(trimmed string) error {
	ref := strings.TrimSpace(strings.TrimPrefix(trimmed, "/switch"))
	if ref == "" {
		return fmt.Errorf("unknown command")
	}
	// Session-path-changing critical sequence: serialize with /resume, /new,
	// /fork, and switchModel so the controller and lease move together.
	s.bindMu.Lock()
	defer s.bindMu.Unlock()
	ctrl, ok := s.ctl().(*control.Controller)
	if !ok {
		return fmt.Errorf("unknown command")
	}
	if _, err := ctrl.SwitchBranch(ref); err != nil {
		return err
	}
	applySessionApprovalModeFor(ctrl, ctrl.SessionPath())
	return nil
}
