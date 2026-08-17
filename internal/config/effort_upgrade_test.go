package config

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestMigrateOfficialDeepSeekProLowDefaultIsNarrow(t *testing.T) {
	pro := Default().Providers[1]
	pro.SupportedEfforts = append([]string(nil), legacyDeepSeekV4Efforts...)

	customEfforts := pro
	customEfforts.SupportedEfforts = []string{"disabled", "high"}

	customEndpoint := pro
	customEndpoint.BaseURL = "https://gateway.example.com/anthropic"

	customIdentity := pro
	customIdentity.Name = "deepseek-pro-copy"

	customDefault := pro
	customDefault.DefaultEffort = "max"

	cfg := &Config{Providers: []ProviderEntry{pro, customEfforts, customEndpoint, customIdentity, customDefault}}
	migrateOfficialDeepSeekProLowDefault(cfg)

	if got := cfg.Providers[0].SupportedEfforts; !reflect.DeepEqual(got, deepSeekV4Efforts) {
		t.Fatalf("official Pro efforts = %v, want %v", got, deepSeekV4Efforts)
	}
	for i, want := range [][]string{
		customEfforts.SupportedEfforts,
		customEndpoint.SupportedEfforts,
		customIdentity.SupportedEfforts,
		customDefault.SupportedEfforts,
	} {
		if got := cfg.Providers[i+1].SupportedEfforts; !reflect.DeepEqual(got, want) {
			t.Fatalf("custom provider %d efforts = %v, want %v", i, got, want)
		}
	}
}

func TestMigrateVerifiedOpenCodeGoDeepSeekLowDefaultsIsNarrow(t *testing.T) {
	preset, ok := CuratedProviderPreset("opencode-go")
	if !ok || len(preset.Entries) != 1 {
		t.Fatal("OpenCode Go preset missing")
	}
	opencode := preset.Entries[0]
	for _, model := range []string{"deepseek-v4-flash", "deepseek-v4-pro"} {
		override := opencode.ModelOverrides[model]
		override.SupportedEfforts = append([]string(nil), legacyDeepSeekV4Efforts...)
		opencode.ModelOverrides[model] = override
	}

	anthropicPreset, ok := CuratedProviderPreset("opencode-go-deepseek-anthropic")
	if !ok || len(anthropicPreset.Entries) != 1 {
		t.Fatal("OpenCode Go DeepSeek Anthropic preset missing")
	}
	anthropic := anthropicPreset.Entries[0]
	anthropic.SupportedEfforts = append([]string(nil), legacyDeepSeekV4Efforts...)

	customEndpoint := cloneProviderEntry(opencode)
	customEndpoint.BaseURL = "https://gateway.example.com/v1"
	customCatalog := cloneProviderEntry(opencode)
	customCatalog.Models = append(customCatalog.Models, "private-model")
	customPrice := cloneProviderEntry(opencode)
	customPrice.Price = deepSeekV4FlashPriceUSD()
	customProtocol := cloneProviderEntry(opencode)
	protocolOverride := customProtocol.ModelOverrides["deepseek-v4-pro"]
	protocolOverride.ReasoningProtocol = ReasoningProtocolOpenAI
	customProtocol.ModelOverrides["deepseek-v4-pro"] = protocolOverride
	customEfforts := cloneProviderEntry(opencode)
	effortOverride := customEfforts.ModelOverrides["deepseek-v4-flash"]
	effortOverride.SupportedEfforts = []string{"disabled", "high"}
	customEfforts.ModelOverrides["deepseek-v4-flash"] = effortOverride
	customIdentity := cloneProviderEntry(opencode)
	customIdentity.PresetID = ""
	productionAnthropic := ProviderEntry{
		Name: "anthropic-opencode-ai", Kind: "anthropic", BaseURL: "https://opencode.ai/zen/go",
		Models: []string{"deepseek-v4-flash", "deepseek-v4-pro"}, Thinking: "adaptive", Effort: "low",
	}

	protected := []ProviderEntry{customEndpoint, customCatalog, customPrice, customProtocol, customEfforts, customIdentity, productionAnthropic}
	wantProtected := cloneProviderEntries(protected)
	cfg := &Config{Providers: append([]ProviderEntry{opencode, anthropic}, protected...)}
	migrateVerifiedOpenCodeGoDeepSeekLowDefaults(cfg)

	for _, model := range []string{"deepseek-v4-flash", "deepseek-v4-pro"} {
		if got := cfg.Providers[0].ModelOverrides[model].SupportedEfforts; !reflect.DeepEqual(got, deepSeekV4Efforts) {
			t.Fatalf("OpenCode Go %s efforts = %v, want %v", model, got, deepSeekV4Efforts)
		}
	}
	if got := cfg.Providers[1].SupportedEfforts; !reflect.DeepEqual(got, deepSeekV4Efforts) {
		t.Fatalf("OpenCode Go Anthropic efforts = %v, want %v", got, deepSeekV4Efforts)
	}
	for i, want := range wantProtected {
		if got := cfg.Providers[i+2]; !reflect.DeepEqual(got, want) {
			t.Fatalf("custom provider %d changed:\n got: %+v\nwant: %+v", i, got, want)
		}
	}
}

func TestOpenCodeGoDeepSeekV4EffortCapabilitiesIncludeLow(t *testing.T) {
	preset, ok := CuratedProviderPreset("opencode-go")
	if !ok || len(preset.Entries) != 1 {
		t.Fatal("OpenCode Go preset missing")
	}
	cfg := &Config{Providers: preset.Entries}
	for _, model := range []string{"deepseek-v4-flash", "deepseek-v4-pro"} {
		entry, ok := cfg.ResolveModel("opencode-go/" + model)
		if !ok {
			t.Fatalf("opencode-go/%s did not resolve", model)
		}
		if protocol := ReasoningProtocolForEntry(entry); protocol != ReasoningProtocolDeepSeek {
			t.Fatalf("opencode %s protocol = %q, want deepseek", model, protocol)
		}
		if cap := EffortCapabilityForEntry(entry); !cap.Supported || cap.Default != "high" || !stringSlicesEqual(cap.Levels, []string{"auto", "disabled", "low", "high", "max"}) {
			t.Fatalf("opencode %s effort capability = %+v", model, cap)
		}
		if got, err := NormalizeEffort(entry, "low"); err != nil || got != "low" {
			t.Fatalf("opencode %s low = %q/%v, want low/nil", model, got, err)
		}
		for _, alias := range []string{"medium", "xhigh"} {
			if got, err := NormalizeEffort(entry, alias); err != nil || got != "high" {
				t.Fatalf("opencode %s %s = %q/%v, want high/nil", model, alias, got, err)
			}
		}
	}
}

func TestApplyUserConfigUpgradesV7DeepSeekProLowOnce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	pro := Default().Providers[1]
	pro.SupportedEfforts = append([]string(nil), legacyDeepSeekV4Efforts...)
	opencodePreset, _ := CuratedProviderPreset("opencode-go")
	opencode := opencodePreset.Entries[0]
	for _, model := range []string{"deepseek-v4-flash", "deepseek-v4-pro"} {
		override := opencode.ModelOverrides[model]
		override.SupportedEfforts = append([]string(nil), legacyDeepSeekV4Efforts...)
		opencode.ModelOverrides[model] = override
	}
	cfg := &Config{ConfigVersion: 7, Providers: []ProviderEntry{pro, opencode}}
	if err := cfg.SaveTo(path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	changed, err := ApplyUserConfigUpgradesOnStartup(path)
	if err != nil {
		t.Fatalf("ApplyUserConfigUpgradesOnStartup: %v", err)
	}
	if !changed {
		t.Fatal("v7 official Pro config was not upgraded")
	}
	got := LoadForEdit(path)
	if got.ConfigVersion != 8 {
		t.Fatalf("config version = %d, want 8", got.ConfigVersion)
	}
	provider, ok := got.Provider("deepseek-pro")
	if !ok || !reflect.DeepEqual(provider.SupportedEfforts, deepSeekV4Efforts) {
		t.Fatalf("migrated Pro = %+v, want efforts %v", provider, deepSeekV4Efforts)
	}
	provider, ok = got.Provider("opencode-go")
	if !ok {
		t.Fatal("migrated OpenCode Go provider missing")
	}
	for _, model := range []string{"deepseek-v4-flash", "deepseek-v4-pro"} {
		if efforts := provider.ModelOverrides[model].SupportedEfforts; !reflect.DeepEqual(efforts, deepSeekV4Efforts) {
			t.Fatalf("migrated OpenCode Go %s efforts = %v, want %v", model, efforts, deepSeekV4Efforts)
		}
	}

	changed, err = ApplyUserConfigUpgradesOnStartup(path)
	if err != nil {
		t.Fatalf("second ApplyUserConfigUpgradesOnStartup: %v", err)
	}
	if changed {
		t.Fatal("v8 migration was not idempotent")
	}
}
