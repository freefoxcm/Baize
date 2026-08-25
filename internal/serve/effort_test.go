package serve

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"reasonix/internal/config"
	"reasonix/internal/control"
)

func TestApplyEffortEditUpsertsMissingProvider(t *testing.T) {
	edit := &config.Config{}
	entry := &config.ProviderEntry{Name: "deepseek", Kind: "openai", BaseURL: "https://api.deepseek.com", Model: "deepseek-v4"}

	if err := applyEffortEdit(edit, entry, "high"); err != nil {
		t.Fatalf("applyEffortEdit: %v", err)
	}
	got, ok := edit.Provider("deepseek")
	if !ok {
		t.Fatal("provider absent from user config should be upserted so the effort edit lands")
	}
	if got.Effort != "high" {
		t.Fatalf("effort = %q, want high", got.Effort)
	}
}

func TestApplyEffortEditEnablesAnthropicThinking(t *testing.T) {
	edit := &config.Config{}
	entry := &config.ProviderEntry{Name: "anthropic", Kind: "anthropic", BaseURL: "https://api.anthropic.com", Model: "claude-opus-4-8"}

	if err := applyEffortEdit(edit, entry, "max"); err != nil {
		t.Fatalf("applyEffortEdit: %v", err)
	}
	got, _ := edit.Provider("anthropic")
	if got.Thinking != "adaptive" {
		t.Fatalf("thinking = %q, want adaptive (effort needs extended thinking to engage)", got.Thinking)
	}
	if got.Effort != "max" {
		t.Fatalf("effort = %q, want max", got.Effort)
	}
}

func TestApplyEffortEditKeepsExistingAnthropicThinking(t *testing.T) {
	edit := &config.Config{Providers: []config.ProviderEntry{
		{Name: "anthropic", Kind: "anthropic", Model: "claude-opus-4-8", Thinking: "always"},
	}}
	entry := &config.ProviderEntry{Name: "anthropic", Kind: "anthropic", Model: "claude-opus-4-8", Thinking: "always"}

	if err := applyEffortEdit(edit, entry, "low"); err != nil {
		t.Fatalf("applyEffortEdit: %v", err)
	}
	got, _ := edit.Provider("anthropic")
	if got.Thinking != "always" {
		t.Fatalf("thinking = %q, want it left untouched", got.Thinking)
	}
}

// TestEffortHandlerCurrentDefaultsToAuto locks in the "auto" display contract:
// an unset stored effort (explicit /effort auto and never-set both store "")
// must surface as "auto" — not the model's default level — so the /effort menu
// and the effort button highlight auto instead of collapsing to e.g. "high".
func TestEffortHandlerCurrentDefaultsToAuto(t *testing.T) {
	isolateServeHome(t, t.TempDir())
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc, ModelRef: "deepseek-flash/deepseek-v4-flash"})
	s := &Server{ctrl: ctrl, bc: bc}

	rec := httptest.NewRecorder()
	s.effort(rec, httptest.NewRequest(http.MethodGet, "/effort", nil))
	var out struct {
		Current string `json:"current"`
		Default string `json:"default"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode effort response: %v", err)
	}
	if out.Current != "auto" {
		t.Fatalf("current = %q, want auto (default %q must not leak into current)", out.Current, out.Default)
	}
}

// TestEffortHandlerReportsStoredEffort ensures an explicitly stored level still
// wins over the auto fallback.
func TestEffortHandlerReportsStoredEffort(t *testing.T) {
	home := t.TempDir()
	isolateServeHome(t, home)
	body := "[[providers]]\nname = \"deepseek-flash\"\nkind = \"openai\"\nbase_url = \"https://api.deepseek.com\"\nmodel = \"deepseek-v4-flash\"\neffort = \"disabled\"\n"
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc, ModelRef: "deepseek-flash/deepseek-v4-flash"})
	s := &Server{ctrl: ctrl, bc: bc}

	rec := httptest.NewRecorder()
	s.effort(rec, httptest.NewRequest(http.MethodGet, "/effort", nil))
	var out struct {
		Current string `json:"current"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode effort response: %v", err)
	}
	if out.Current != "disabled" {
		t.Fatalf("current = %q, want disabled", out.Current)
	}
}

// TestEffortHandlerReportsRuntimeNormalizedEffort covers the degraded-switch
// case: a stored provider-level effort the active model cannot express runs as
// auto, so GET /effort must report auto (not the stale stored value).
func TestEffortHandlerReportsRuntimeNormalizedEffort(t *testing.T) {
	home := t.TempDir()
	isolateServeHome(t, home)
	body := "[[providers]]\nname = \"opencode-go\"\nkind = \"openai\"\nbase_url = \"https://opencode.ai/zen/go/v1\"\nmodels = [\"deepseek-v4-flash\", \"glm-5.2\"]\ndefault = \"deepseek-v4-flash\"\neffort = \"disabled\"\nmodel_overrides = { \"deepseek-v4-flash\" = { reasoning_protocol = \"deepseek\", supported_efforts = [\"disabled\", \"high\", \"max\"], default_effort = \"high\" } }\n"
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc, ModelRef: "opencode-go/glm-5.2"})
	s := &Server{ctrl: ctrl, bc: bc}

	rec := httptest.NewRecorder()
	s.effort(rec, httptest.NewRequest(http.MethodGet, "/effort", nil))
	var out struct {
		Current string `json:"current"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode effort response: %v", err)
	}
	if out.Current != "auto" {
		t.Fatalf("current = %q, want auto (stored disabled degrades at runtime)", out.Current)
	}
}

func TestEffortHandlerHidesNonReasoningCustomModel(t *testing.T) {
	home := t.TempDir()
	isolateServeHome(t, home)
	body := "default_model = \"opencode-proxy/ox-alpha-free\"\n[[providers]]\nname = \"opencode-proxy\"\nkind = \"openai\"\nbase_url = \"https://proxy.example/v1\"\nmodel = \"ox-alpha-free\"\nreasoning_protocol = \"none\"\n"
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc, ModelRef: "opencode-proxy/ox-alpha-free"})
	s := &Server{ctrl: ctrl, bc: bc}

	rec := httptest.NewRecorder()
	s.effort(rec, httptest.NewRequest(http.MethodGet, "/effort", nil))
	var out struct {
		Supported bool     `json:"supported"`
		Levels    []string `json:"levels"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode effort response: %v", err)
	}
	if out.Supported || len(out.Levels) != 0 {
		t.Fatalf("ox-alpha-free effort capability = supported:%t levels:%v, want hidden", out.Supported, out.Levels)
	}
}

func TestEffortHandlerShowsOxAlphaBehindCustomProxy(t *testing.T) {
	home := t.TempDir()
	isolateServeHome(t, home)
	body := "default_model = \"opencode-proxy/ox-alpha-free\"\n[[providers]]\nname = \"opencode-proxy\"\nkind = \"openai\"\nbase_url = \"https://proxy.example/v1\"\nmodel = \"ox-alpha-free\"\n"
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc, ModelRef: "opencode-proxy/ox-alpha-free"})
	s := &Server{ctrl: ctrl, bc: bc}

	rec := httptest.NewRecorder()
	s.effort(rec, httptest.NewRequest(http.MethodGet, "/effort", nil))
	var out struct {
		Supported bool     `json:"supported"`
		Levels    []string `json:"levels"`
		Current   string   `json:"current"`
		Default   string   `json:"default"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode effort response: %v", err)
	}
	want := []string{"auto", "low", "high", "max"}
	if !out.Supported || !slices.Equal(out.Levels, want) || out.Current != "auto" || out.Default != "auto" {
		t.Fatalf("Ox Alpha effort capability = %+v, want levels %v and auto defaults", out, want)
	}
}

// TestThinkingAliasSubmitsEffort covers the /thinking alias: a bare command
// reports the current effort capability, a level argument switches through the
// same path as /effort.
func TestThinkingAliasSubmitsEffort(t *testing.T) {
	isolateServeHome(t, t.TempDir())
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc, ModelRef: "deepseek-flash/deepseek-v4-flash"})
	s := &Server{ctrl: ctrl, bc: bc}

	// Bare /thinking reports capability JSON like /effort.
	rec := httptest.NewRecorder()
	s.submit(rec, httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(`{"input":"/thinking"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("bare /thinking status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	var out struct {
		Current string `json:"current"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode /thinking response: %v", err)
	}
	if out.Current == "" {
		t.Fatalf("bare /thinking current empty: %s", rec.Body.String())
	}

	// /thinking <level> switches like /effort and returns 204.
	rec = httptest.NewRecorder()
	s.submit(rec, httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(`{"input":"/thinking max"}`)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("/thinking max status = %d, want 204 (body %s)", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	s.effort(rec, httptest.NewRequest(http.MethodGet, "/effort", nil))
	var after struct {
		Current string `json:"current"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &after); err != nil {
		t.Fatalf("decode effort response: %v", err)
	}
	if after.Current != "max" {
		t.Fatalf("current after /thinking max = %q, want max", after.Current)
	}
}
