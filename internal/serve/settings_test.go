package serve

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"reasonix/internal/config"
	"reasonix/internal/control"
)

func writeSettingsTestConfig(t *testing.T) {
	t.Helper()
	isolateServeHome(t, t.TempDir())
	path := config.UserConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `default_model = "local/model-a"

[[providers]]
name = "local"
kind = "openai"
base_url = "http://127.0.0.1:1/v1"
models = ["model-a", "model-b"]
default = "model-a"
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestApplySafeSettingsPatchValidatesConcurrency(t *testing.T) {
	cfg := config.Default()
	total, writers := 2, 3
	err := applySafeSettingsPatch(cfg, settingsPatch{MaxSubagentConcurrency: &total, MaxParallelWriters: &writers})
	if err == nil || !strings.Contains(err.Error(), "cannot exceed") {
		t.Fatalf("error = %v", err)
	}
}

func TestSettingsRevisionConflictAndSafeUpdate(t *testing.T) {
	writeSettingsTestConfig(t)
	bc := NewBroadcaster()
	root := t.TempDir()
	ctrl := control.New(control.Options{Sink: bc, Runner: blockingRunner{}, Label: "model-a", ModelRef: "local/model-a", SessionDir: t.TempDir(), WorkspaceRoot: root})
	server := New(ctrl, bc, config.ServeConfig{})
	rebuilt := make(chan struct{}, 1)
	server.rebuildController = func(_ context.Context, _ *control.Controller, ref string) (*control.Controller, error) {
		rebuilt <- struct{}{}
		return control.New(control.Options{Sink: bc, Label: "model-a", ModelRef: ref, SessionDir: ctrl.SessionDir(), WorkspaceRoot: root}), nil
	}

	get := httptest.NewRecorder()
	server.Handler().ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/settings", nil))
	if get.Code != http.StatusOK {
		t.Fatalf("GET status = %d: %s", get.Code, get.Body.String())
	}
	if strings.Contains(get.Body.String(), "base_url") || strings.Contains(get.Body.String(), "api_key") {
		t.Fatalf("safe settings exposed provider configuration: %s", get.Body.String())
	}
	var initial settingsResponse
	if err := json.Unmarshal(get.Body.Bytes(), &initial); err != nil {
		t.Fatal(err)
	}
	model := "local/model-b"
	approval := "ask"
	ctrl.Submit("keep settings pending")
	waitRunning(t, ctrl)
	payload, _ := json.Marshal(settingsPatch{Revision: initial.Revision, DefaultModel: &model, DefaultApprovalMode: &approval})
	patch := httptest.NewRecorder()
	patchReq := httptest.NewRequest(http.MethodPatch, "/settings", bytes.NewReader(payload))
	patchReq.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(patch, patchReq)
	if patch.Code != http.StatusOK {
		t.Fatalf("PATCH status = %d: %s", patch.Code, patch.Body.String())
	}
	var updated settingsResponse
	if err := json.Unmarshal(patch.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if updated.Revision == initial.Revision || updated.Global.DefaultModel != model || updated.Apply != "pending" {
		t.Fatalf("updated settings = %#v", updated)
	}

	conflict := httptest.NewRecorder()
	conflictReq := httptest.NewRequest(http.MethodPatch, "/settings", bytes.NewReader(payload))
	conflictReq.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(conflict, conflictReq)
	if conflict.Code != http.StatusConflict {
		t.Fatalf("stale PATCH status = %d: %s", conflict.Code, conflict.Body.String())
	}
	ctrl.Cancel()
	select {
	case <-rebuilt:
	case <-time.After(2 * time.Second):
		t.Fatal("pending settings were not applied after TurnDone")
	}
}
