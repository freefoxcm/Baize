package serve

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"reasonix/internal/agent"
	"reasonix/internal/boot"
	"reasonix/internal/config"
	"reasonix/internal/control"
	"reasonix/internal/event"
)

// TestBuildPreservesSessionDirAndWorkspace is the regression guard for the
// "left-side session list disappears after /model or /effort" bug: the serve
// rebuild path (build → boot.Build) must carry the serving workspace's
// per-workspace session store and workspace root, instead of letting
// boot.Build fall back to the global session dir.
func TestBuildPreservesSessionDirAndWorkspace(t *testing.T) {
	t.Setenv("REASONIX_HOME", t.TempDir())
	proj := t.TempDir()
	sessionDir := config.ProjectSessionDir(proj)
	if sessionDir == "" {
		t.Fatal("ProjectSessionDir resolved empty")
	}

	bc := NewBroadcaster()
	old := control.New(control.Options{
		Executor:      agent.New(nil, nil, agent.NewSession("sys"), agent.Options{}, event.Discard),
		Sink:          bc,
		SessionDir:    sessionDir,
		WorkspaceRoot: proj,
		Label:         "old",
	})
	s := &Server{ctrl: old, bc: bc, tokenMode: boot.TokenModeFull}

	newCtrl, err := s.build(context.Background(), "deepseek-flash/deepseek-v4-flash")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	defer newCtrl.Close()

	if got := newCtrl.SessionDir(); got != sessionDir {
		t.Errorf("SessionDir after rebuild = %q, want %q (per-workspace store must survive)", got, sessionDir)
	}
	if got := newCtrl.WorkspaceRoot(); got != proj {
		t.Errorf("WorkspaceRoot after rebuild = %q, want %q", got, proj)
	}
}

// TestSwitchModelKeepsSessionList verifies the user-visible contract: after a
// model switch the GET /sessions list still shows the workspace's sessions.
func TestSwitchModelKeepsSessionList(t *testing.T) {
	t.Setenv("REASONIX_HOME", t.TempDir())
	proj := t.TempDir()
	sessionDir := config.ProjectSessionDir(proj)
	if sessionDir == "" {
		t.Fatal("ProjectSessionDir resolved empty")
	}

	bc := NewBroadcaster()
	sessionFile := agent.NewSessionPath(sessionDir, "old-model")
	if err := os.MkdirAll(filepath.Dir(sessionFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sessionFile, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	old := control.New(control.Options{
		Executor:      agent.New(nil, nil, agent.NewSession("sys"), agent.Options{}, event.Discard),
		Sink:          bc,
		SessionDir:    sessionDir,
		SessionPath:   sessionFile,
		WorkspaceRoot: proj,
		Label:         "old",
	})
	s := &Server{ctrl: old, bc: bc, tokenMode: boot.TokenModeFull}
	s.titles = newTitleCache(sessionDir)

	if err := s.switchModel(context.Background(), "deepseek-flash/deepseek-v4-flash"); err != nil {
		t.Fatalf("switchModel: %v", err)
	}

	if got := s.ctl().SessionDir(); got != sessionDir {
		t.Fatalf("SessionDir after switch = %q, want %q", got, sessionDir)
	}

	rec := httptest.NewRecorder()
	s.sessions(rec, httptest.NewRequest(http.MethodGet, "/sessions", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("sessions status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), filepath.Base(sessionFile)) {
		t.Fatalf("sessions after switch missing %q, body: %s", filepath.Base(sessionFile), rec.Body.String())
	}
}

// TestSwitchProfilePreservesApprovalMode extends the same guard to the
// work-mode rebuild path (switchProfile), which shares build() and the
// authorization carry-over but runs its own publish sequence.
func TestSwitchProfilePreservesApprovalMode(t *testing.T) {
	bc := NewBroadcaster()
	old := control.New(control.Options{
		Executor: agent.New(nil, nil, agent.NewSession("sys"), agent.Options{}, event.Discard),
		Sink:     bc,
		Label:    "old",
	})
	old.SetToolApprovalMode("auto")

	s := &Server{ctrl: old, bc: bc, tokenMode: boot.TokenModeFull}
	var built *control.Controller
	s.buildController = func(_ context.Context, _ string) (*control.Controller, error) {
		built = control.New(control.Options{Sink: bc, Label: "new"})
		return built, nil
	}

	if err := s.switchProfile(context.Background(), boot.TokenModeEconomy); err != nil {
		t.Fatalf("switchProfile: %v", err)
	}
	defer built.Close()

	if got := s.ctl().ToolApprovalMode(); got != "auto" {
		t.Fatalf("approval mode after work-mode switch = %q, want %q", got, "auto")
	}
}

// TestSwitchModelPreservesApprovalMode guards the runtime approval posture
// (ask/auto/yolo) across a rebuild: boot.Build starts from config defaults, so
// switchModel must copy the outgoing controller's mode onto the replacement.
func TestSwitchModelPreservesApprovalMode(t *testing.T) {
	bc := NewBroadcaster()
	old := control.New(control.Options{
		Executor: agent.New(nil, nil, agent.NewSession("sys"), agent.Options{}, event.Discard),
		Sink:     bc,
		Label:    "old",
	})
	old.SetToolApprovalMode("auto")

	s := &Server{ctrl: old, bc: bc, tokenMode: boot.TokenModeFull}
	var built *control.Controller
	s.buildController = func(_ context.Context, _ string) (*control.Controller, error) {
		built = control.New(control.Options{Sink: bc, Label: "new"})
		return built, nil
	}

	if err := s.switchModel(context.Background(), "next-model"); err != nil {
		t.Fatalf("switchModel: %v", err)
	}
	defer built.Close()

	if got := s.ctl().ToolApprovalMode(); got != "auto" {
		t.Fatalf("approval mode after switch = %q, want %q", got, "auto")
	}
}
