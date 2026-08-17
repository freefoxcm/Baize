package config

import "strings"

var (
	legacyDeepSeekV4Efforts = []string{"disabled", "high", "max"}
	deepSeekV4Efforts       = []string{"disabled", "low", "high", "max"}
)

func migrateOfficialDeepSeekProLowDefault(c *Config) {
	if c == nil {
		return
	}
	for i := range c.Providers {
		p := &c.Providers[i]
		models := p.ModelList()
		if strings.TrimSpace(p.Name) != "deepseek-pro" ||
			!isOfficialDeepSeekBillingEndpoint(p) ||
			!isStandardDeepSeekProviderTemplate(p) ||
			len(models) != 1 || strings.TrimSpace(models[0]) != "deepseek-v4-pro" ||
			!stringSlicesEqual(p.SupportedEfforts, legacyDeepSeekV4Efforts) ||
			strings.TrimSpace(p.DefaultEffort) != "high" {
			continue
		}
		p.SupportedEfforts = append([]string(nil), deepSeekV4Efforts...)
	}
}

func migrateVerifiedOpenCodeGoDeepSeekLowDefaults(c *Config) {
	if c == nil {
		return
	}
providerLoop:
	for i := range c.Providers {
		p := &c.Providers[i]
		if strings.TrimSpace(p.PresetID) == "opencode-go" &&
			strings.EqualFold(strings.TrimSpace(p.Kind), "openai") &&
			normalizedBaseURLForMigration(p.BaseURL) == "https://opencode.ai/zen/go/v1" &&
			strings.TrimSpace(p.Model) == "" && stringSlicesEqual(p.Models, opencodeGoModels) &&
			p.Price == nil && len(p.Prices) == 0 && strings.TrimSpace(p.ReasoningProtocol) == "" &&
			len(p.SupportedEfforts) == 0 && strings.TrimSpace(p.DefaultEffort) == "" {
			models := []string{"deepseek-v4-flash", "deepseek-v4-pro"}
			for _, model := range models {
				override, ok := p.ModelOverrides[model]
				if !ok || strings.TrimSpace(override.ReasoningProtocol) != ReasoningProtocolDeepSeek ||
					!stringSlicesEqual(override.SupportedEfforts, legacyDeepSeekV4Efforts) ||
					strings.TrimSpace(override.DefaultEffort) != "high" {
					continue providerLoop
				}
			}
			for _, model := range models {
				override := p.ModelOverrides[model]
				override.SupportedEfforts = append([]string(nil), deepSeekV4Efforts...)
				p.ModelOverrides[model] = override
			}
		}

		if strings.TrimSpace(p.PresetID) == "opencode-go-deepseek-anthropic" &&
			strings.EqualFold(strings.TrimSpace(p.Kind), "anthropic") &&
			normalizedBaseURLForMigration(p.BaseURL) == "https://opencode.ai/zen/go" &&
			strings.TrimSpace(p.Model) == "" && stringSlicesEqual(p.Models, []string{"deepseek-v4-flash"}) &&
			p.Price == nil && len(p.Prices) == 0 && strings.TrimSpace(p.ReasoningProtocol) == "" &&
			stringSlicesEqual(p.SupportedEfforts, legacyDeepSeekV4Efforts) &&
			strings.TrimSpace(p.DefaultEffort) == "high" {
			p.SupportedEfforts = append([]string(nil), deepSeekV4Efforts...)
		}
	}
}
