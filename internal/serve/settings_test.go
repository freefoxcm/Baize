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
vision_models = ["model-b"]
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

func TestApplySafeSettingsPatchValidatesVisionModel(t *testing.T) {
	t.Setenv("BAIZE_SERVE_TEST_MISSING_VISION_KEY", "")
	cfg := config.Default()
	cfg.Providers = []config.ProviderEntry{
		{Name: "local", Kind: "openai", BaseURL: "http://127.0.0.1:1/v1", Models: []string{"text", "vision"}, Default: "text", VisionModels: []string{"vision"}},
		{Name: "locked", Kind: "openai", BaseURL: "https://example.invalid/v1", Models: []string{"vision"}, Default: "vision", APIKeyEnv: "BAIZE_SERVE_TEST_MISSING_VISION_KEY", VisionModels: []string{"vision"}},
	}

	for _, value := range []string{"", "auto", "local/vision"} {
		if err := applySafeSettingsPatch(cfg, settingsPatch{VisionModel: &value}); err != nil {
			t.Fatalf("visionModel %q: %v", value, err)
		}
	}
	if cfg.Agent.VisionModel != "local/vision" {
		t.Fatalf("vision model = %q", cfg.Agent.VisionModel)
	}
	for _, test := range []struct {
		value string
		want  string
	}{
		{value: "local/text", want: "does not support image input"},
		{value: "missing/vision", want: "no such model"},
		{value: "locked/vision", want: "has no key"},
	} {
		err := applySafeSettingsPatch(cfg, settingsPatch{VisionModel: &test.value})
		if err == nil || !strings.Contains(err.Error(), test.want) {
			t.Fatalf("visionModel %q error = %v, want %q", test.value, err, test.want)
		}
	}
}

func TestSafeSettingsListsConfiguredVisionModels(t *testing.T) {
	cfg := config.Default()
	cfg.Providers = []config.ProviderEntry{{
		Name: "local", Kind: "openai", BaseURL: "http://127.0.0.1:1/v1",
		Models: []string{"text", "vision"}, Default: "text", VisionModels: []string{"vision"},
	}}
	cfg.Agent.VisionModel = "auto"
	view := safeSettingsFromConfig(cfg)
	if view.VisionModel != "auto" {
		t.Fatalf("visionModel = %q", view.VisionModel)
	}
	if got := strings.Join(view.VisionModels, ","); got != "local/vision" {
		t.Fatalf("visionModels = %q", got)
	}
}

func TestSettingsRejectsUnavailableVisionModel(t *testing.T) {
	writeSettingsTestConfig(t)
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc, Label: "model-a", ModelRef: "local/model-a", SessionDir: t.TempDir(), WorkspaceRoot: t.TempDir()})
	server := New(ctrl, bc, config.ServeConfig{})
	initial, err := server.safeSettingsView()
	if err != nil {
		t.Fatal(err)
	}
	for _, model := range []string{"local/model-a", "missing/vision"} {
		payload, _ := json.Marshal(settingsPatch{Revision: initial.Revision, VisionModel: &model})
		recorder := httptest.NewRecorder()
		req := localTestRequest(http.MethodPatch, "/settings", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		server.Handler().ServeHTTP(recorder, req)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("visionModel %q status = %d: %s", model, recorder.Code, recorder.Body.String())
		}
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
	server.Handler().ServeHTTP(get, localTestRequest(http.MethodGet, "/settings", nil))
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
	visionModel := "local/model-b"
	approval := "ask"
	ctrl.Submit("keep settings pending")
	waitRunning(t, ctrl)
	payload, _ := json.Marshal(settingsPatch{Revision: initial.Revision, DefaultModel: &model, VisionModel: &visionModel, DefaultApprovalMode: &approval})
	patch := httptest.NewRecorder()
	patchReq := localTestRequest(http.MethodPatch, "/settings", bytes.NewReader(payload))
	patchReq.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(patch, patchReq)
	if patch.Code != http.StatusOK {
		t.Fatalf("PATCH status = %d: %s", patch.Code, patch.Body.String())
	}
	var updated settingsResponse
	if err := json.Unmarshal(patch.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if updated.Revision == initial.Revision || updated.Global.DefaultModel != model || updated.Global.VisionModel != visionModel || updated.Apply != "pending" {
		t.Fatalf("updated settings = %#v", updated)
	}

	conflict := httptest.NewRecorder()
	conflictReq := localTestRequest(http.MethodPatch, "/settings", bytes.NewReader(payload))
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

func TestTaskErrorVisibilitySavesWithoutControllerRebuild(t *testing.T) {
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

	initial, err := server.safeSettingsView()
	if err != nil {
		t.Fatal(err)
	}
	if initial.Global.ShowTaskErrors {
		t.Fatal("showTaskErrors must default to false")
	}
	enabled := true
	payload, _ := json.Marshal(settingsPatch{Revision: initial.Revision, ShowTaskErrors: &enabled})
	recorder := httptest.NewRecorder()
	req := localTestRequest(http.MethodPatch, "/settings", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	server.Handler().ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("PATCH status = %d: %s", recorder.Code, recorder.Body.String())
	}
	var updated settingsResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if !updated.Global.ShowTaskErrors || updated.Apply != "applied" {
		t.Fatalf("updated settings = %#v", updated)
	}
	select {
	case <-rebuilt:
		t.Fatal("display-only setting rebuilt the controller")
	case <-time.After(100 * time.Millisecond):
	}
	persisted, err := config.LoadUserConfigReadOnly()
	if err != nil {
		t.Fatal(err)
	}
	if !persisted.Serve.ShowTaskErrors {
		t.Fatal("show_task_errors was not persisted")
	}
}
