package cli

import (
	"bufio"
	"reflect"
	"strings"
	"testing"

	"reasonix/internal/config"
	"reasonix/internal/i18n"
)

func TestPromptResponsesProviderManualBuildsResponsesEntry(t *testing.T) {
	result, err := promptResponsesProviderManualWith(
		bufio.NewScanner(strings.NewReader("relay-responses\nmodel-x\n262144\n")),
		"https://relay.example/v1", "RELAY_API_KEY", "secret",
	)
	if err != nil {
		t.Fatalf("promptResponsesProviderManualWith: %v", err)
	}
	if len(result.entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(result.entries))
	}
	entry := result.entries[0]
	if entry.Name != "relay-responses" || entry.Kind != "responses" || entry.BaseURL != "https://relay.example/v1" || entry.DefaultModel() != "model-x" {
		t.Fatalf("entry = %+v", entry)
	}
	if entry.APIKeyEnv != "RELAY_API_KEY" || entry.ContextWindow != 262144 || entry.ResponsesMode != "" {
		t.Fatalf("entry settings = %+v", entry)
	}
	if result.credentials["RELAY_API_KEY"] != "secret" {
		t.Fatal("staged credential missing")
	}
}

func TestNewResponsesSetupEntryAppliesOpenCodeGoContract(t *testing.T) {
	models := []string{"gpt-5.6-luna", "muse-spark-1.2-contributor"}
	entry := newResponsesSetupEntry("opencode-go-responses", "https://opencode.ai/zen/go/v1", models, "OPENCODE_API_KEY", 500000)
	if entry.Kind != "responses" || entry.ResponsesMode != "stateless" || entry.PresetID != "opencode-go-responses" || entry.PresetVersion != config.ProviderPresetVersion {
		t.Fatalf("route contract = %+v", entry)
	}
	if entry.MaxOutputTokens != 32768 || entry.BillingMode != "subscription_equivalent" || !reflect.DeepEqual(entry.VisionModels, models) {
		t.Fatalf("provider defaults = %+v", entry)
	}
	wantEfforts := map[string][]string{
		"gpt-5.6-luna":               {"none", "low", "medium", "high", "xhigh", "max"},
		"muse-spark-1.2-contributor": {"minimal", "low", "medium", "high", "xhigh"},
	}
	for model, want := range wantEfforts {
		override := entry.ModelOverrides[model]
		if !reflect.DeepEqual(override.SupportedEfforts, want) || override.DefaultEffort != "high" {
			t.Fatalf("override %s = %+v, want efforts %v and default high", model, override, want)
		}
	}
}

func TestProviderManagerItemsIncludesResponsesAddAction(t *testing.T) {
	items := providerManagerItems(newProviderSetupSession(&config.Config{}))
	if len(items) != 5 {
		t.Fatalf("items = %d, want 5", len(items))
	}
	if items[2].name != i18n.M.SetupAddResponses || items[2].desc != i18n.M.ResponsesProviderDesc {
		t.Fatalf("responses item = %+v", items[2])
	}
}
