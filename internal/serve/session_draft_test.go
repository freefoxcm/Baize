package serve

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"reasonix/internal/agent"
	"reasonix/internal/config"
	"reasonix/internal/control"
)

func TestSessionsSynthesizesCurrentDraftWithoutCreatingTranscript(t *testing.T) {
	dir := t.TempDir()
	draftPath := filepath.Join(dir, "session-draft.jsonl")
	ctrl := control.New(control.Options{SessionDir: dir, SessionPath: draftPath, Label: "test"})
	server := New(ctrl, NewBroadcaster(), config.ServeConfig{})
	req := localTestRequest(http.MethodGet, "/sessions", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var sessions []struct {
		Path    string `json:"path"`
		Title   string `json:"title"`
		Current bool   `json:"current"`
		Draft   bool   `json:"draft"`
		Turns   int    `json:"turns"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &sessions); err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].Path != agent.CanonicalSessionPath(draftPath) || sessions[0].Title != "新会话" || !sessions[0].Current || !sessions[0].Draft || sessions[0].Turns != 0 {
		t.Fatalf("sessions = %#v", sessions)
	}
	if _, err := filepath.Glob(filepath.Join(dir, "*.jsonl")); err != nil {
		t.Fatal(err)
	} else if files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl")); len(files) != 0 {
		t.Fatalf("draft listing created transcripts: %v", files)
	}
}

func TestNewSessionDefaultModelBuildFailureKeepsCurrentController(t *testing.T) {
	writeSettingsTestConfig(t)
	cfg := config.LoadForEdit(config.UserConfigPath())
	if err := cfg.SetDefaultModel("local/model-b"); err != nil {
		t.Fatal(err)
	}
	if err := cfg.SaveTo(config.UserConfigPath()); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "current.jsonl")
	ctrl := control.New(control.Options{SessionDir: dir, SessionPath: path, Label: "model-a", ModelRef: "local/model-a"})
	server := New(ctrl, NewBroadcaster(), config.ServeConfig{})
	server.buildController = func(context.Context, string) (*control.Controller, error) {
		return nil, errors.New("builder unavailable")
	}
	req := localTestRequest(http.MethodPost, "/new", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", rec.Code, rec.Body.String())
	}
	if server.ctl() != ctrl || server.ctl().SessionPath() != path {
		t.Fatal("failed default-model build replaced or rotated the current controller")
	}
}
