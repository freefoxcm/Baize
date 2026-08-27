package config

import "strings"

var (
	preOxAlphaOpenCodeGoModels       = []string{"glm-5.3", "glm-5.2", "glm-5.1", "kimi-k3", "kimi-k2.7-code", "kimi-k2.6", "deepseek-v4-pro", "deepseek-v4-flash", "mimo-v2.5-pro", "mimo-v2.5", "hy3"}
	preOxAlphaOpenCodeGoVisionModels = []string{"kimi-k3"}
	preGLM53FlashOpenCodeGoModels    = []string{"ox-alpha-free", "glm-5.3", "glm-5.2", "glm-5.1", "kimi-k3", "kimi-k2.7-code", "kimi-k2.6", "deepseek-v4-pro", "deepseek-v4-flash", "mimo-v2.5-pro", "mimo-v2.5", "hy3"}
	preGLM53FlashVisionModels        = []string{"ox-alpha-free", "kimi-k3"}
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

// normalizeLegacyOpenCodeGoChatCatalog upgrades only previously canonical Chat
// catalogs. Edited catalogs and vision choices stay user-owned.
func normalizeLegacyOpenCodeGoChatCatalog(c *Config) (changed bool) {
	if c == nil {
		return false
	}
	for i := range c.Providers {
		p := &c.Providers[i]
		if !isMigratableOpenCodeGoChatEntry(p) ||
			!strings.EqualFold(strings.TrimSpace(p.Kind), "openai") ||
			normalizedBaseURLForMigration(p.BaseURL) != "https://opencode.ai/zen/go/v1" ||
			(!stringSlicesEqual(p.Models, preOxAlphaOpenCodeGoModels) &&
				!stringSlicesEqual(p.Models, preGLM53FlashOpenCodeGoModels)) || strings.TrimSpace(p.Model) != "" {
			continue
		}
		p.Models = append([]string(nil), opencodeGoModels...)
		if stringSlicesEqual(p.VisionModels, preOxAlphaOpenCodeGoVisionModels) ||
			stringSlicesEqual(p.VisionModels, preGLM53FlashVisionModels) {
			p.VisionModels = append([]string(nil), opencodeGoVisionModels...)
		}
		changed = true
	}
	return changed
}
