package serve

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"reasonix/internal/config"
	"reasonix/internal/control"
)

func TestServeStatusSeparatesWorkspaceFromSessionStorage(t *testing.T) {
	bc := NewBroadcaster()
	sessionDir := t.TempDir()
	workspaceRoot := t.TempDir()
	ctrl := control.New(control.Options{Sink: bc, SessionDir: sessionDir, WorkspaceRoot: workspaceRoot})
	t.Cleanup(ctrl.Close)
	recorder := httptest.NewRecorder()
	New(ctrl, bc, config.ServeConfig{}).Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/status", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["cwd"] != sessionDir {
		t.Fatalf("cwd = %v, want compatibility session dir %q", body["cwd"], sessionDir)
	}
	if body["workspaceRoot"] != workspaceRoot {
		t.Fatalf("workspaceRoot = %v, want %q", body["workspaceRoot"], workspaceRoot)
	}
}
