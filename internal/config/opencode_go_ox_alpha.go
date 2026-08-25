package config

import "strings"

var (
	preOxAlphaOpenCodeGoModels       = []string{"glm-5.3", "glm-5.2", "glm-5.1", "kimi-k3", "kimi-k2.7-code", "kimi-k2.6", "deepseek-v4-pro", "deepseek-v4-flash", "mimo-v2.5-pro", "mimo-v2.5", "hy3"}
	preOxAlphaOpenCodeGoVisionModels = []string{"kimi-k3"}
)

func migrateLegacyOpenCodeGoVisionModels(current []string) []string {
	if current != nil {
		return migrateKimiK3VisionModels(current, nil)
	}
	return append([]string(nil), opencodeGoVisionModels...)
}

func isMigratableOpenCodeGoChatEntry(p *ProviderEntry) bool {
	if p == nil {
		return false
	}
	switch strings.TrimSpace(p.PresetID) {
	case "opencode-go", "opencode-go-recommended":
		return true
	case "":
		return strings.TrimSpace(p.Name) == "opencode-go"
	default:
		return false
	}
}

// normalizeLegacyOpenCodeGoOxAlphaCatalog upgrades only the previously
// canonical Chat catalog. Edited catalogs and vision choices stay user-owned.
func normalizeLegacyOpenCodeGoOxAlphaCatalog(c *Config) (changed bool) {
	if c == nil {
		return false
	}
	for i := range c.Providers {
		p := &c.Providers[i]
		if !isMigratableOpenCodeGoChatEntry(p) ||
			!strings.EqualFold(strings.TrimSpace(p.Kind), "openai") ||
			normalizedBaseURLForMigration(p.BaseURL) != "https://opencode.ai/zen/go/v1" ||
			!stringSlicesEqual(p.Models, preOxAlphaOpenCodeGoModels) || strings.TrimSpace(p.Model) != "" {
			continue
		}
		p.Models = append([]string(nil), opencodeGoModels...)
		if p.ModelOverrides == nil {
			p.ModelOverrides = map[string]ProviderModelOverride{}
		}
		overrideKey := "ox-alpha-free"
		for key := range p.ModelOverrides {
			if strings.EqualFold(strings.TrimSpace(key), overrideKey) {
				overrideKey = key
				break
			}
		}
		oxAlpha := p.ModelOverrides[overrideKey]
		if oxAlpha.ContextWindow == 0 {
			oxAlpha.ContextWindow = 1_000_000
			p.ModelOverrides[overrideKey] = oxAlpha
		}
		if stringSlicesEqual(p.VisionModels, preOxAlphaOpenCodeGoVisionModels) {
			p.VisionModels = append([]string(nil), opencodeGoVisionModels...)
		}
		changed = true
	}
	return changed
}
