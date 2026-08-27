package config

import "testing"

func TestOpenCodeGoGLM53FlashPresetCapability(t *testing.T) {
	preset, ok := CuratedProviderPreset("opencode-go")
	if !ok || len(preset.Entries) != 1 {
		t.Fatalf("opencode-go preset = %+v, found %t", preset, ok)
	}
	var cfg Config
	if err := cfg.UpsertProvider(preset.Entries[0]); err != nil {
		t.Fatalf("upsert opencode-go preset: %v", err)
	}
	entry, ok := cfg.ResolveModel("opencode-go/glm-5.3-flash")
	if !ok {
		t.Fatal("opencode-go/glm-5.3-flash did not resolve")
	}
	if protocol := ReasoningProtocolForEntry(entry); protocol != ReasoningProtocolOpenAI {
		t.Fatalf("protocol = %q, want openai", protocol)
	}
	capability := EffortCapabilityForEntry(entry)
	wantLevels := []string{"auto", "low", "high", "max"}
	if !capability.Supported || capability.Default != "auto" || !stringSlicesEqual(capability.Levels, wantLevels) {
		t.Fatalf("effort capability = %+v, want levels %v and auto default", capability, wantLevels)
	}
	if entry.ContextWindow != 1_000_000 || entry.MaxOutputTokens != 32_768 || !EffectiveVision(entry) {
		t.Fatalf("context/output/vision capability mismatch: %+v", entry)
	}
}
