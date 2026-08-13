package boot

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"reasonix/internal/event"
)

// TestBuildDegradesUnsupportedStoredEffort is the regression guard for the
// "cannot switch models after /effort disabled" bug: stored effort is
// provider-scoped, so switching to a model on the same provider whose
// capability vocabulary does not include the stored level must degrade to auto
// instead of hard-failing the build (openai: effort must be low, medium, or
// high). The provider keeps its stored level for models that accept it.
func TestBuildDegradesUnsupportedStoredEffort(t *testing.T) {
	home := t.TempDir()
	fenceBootTestHistoryCatalog(t)
	t.Setenv("REASONIX_HOME", home)
	body := `default_model = "opencode-go/deepseek-v4-flash"
[[providers]]
name = "opencode-go"
kind = "openai"
base_url = "https://opencode.ai/zen/go/v1"
models = ["deepseek-v4-flash", "glm-5.2"]
default = "deepseek-v4-flash"
effort = "disabled"
model_overrides = { "deepseek-v4-flash" = { reasoning_protocol = "deepseek", supported_efforts = ["disabled", "high", "max"], default_effort = "high", context_window = 1048576 } }
`
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	// Switching to a model that cannot express "disabled" must succeed
	// (effort degrades to auto), not fail with a provider validation error.
	ctrl, err := Build(context.Background(), Options{
		Sink:  event.Discard,
		Model: "opencode-go/glm-5.2",
	})
	if err != nil {
		t.Fatalf("Build with unsupported stored effort: %v", err)
	}
	ctrl.Close()

	// Switching to a model that does support "disabled" must keep it.
	ctrl2, err := Build(context.Background(), Options{
		Sink:  event.Discard,
		Model: "opencode-go/deepseek-v4-flash",
	})
	if err != nil {
		t.Fatalf("Build with supported stored effort: %v", err)
	}
	ctrl2.Close()
}
