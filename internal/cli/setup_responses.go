package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"reasonix/internal/config"
	"reasonix/internal/i18n"
	"reasonix/internal/netclient"
	"reasonix/internal/provider"
)

func promptResponsesProvider(proxy netclient.ProxySpec) (providerPromptResult, error) {
	methodIdx, err := selectOne(i18n.M.ResponsesAddMethodLabel, []menuItem{
		{name: i18n.M.CustomMethodManual},
		{name: i18n.M.CustomMethodURL},
	})
	if err != nil {
		return providerPromptResult{}, err
	}
	if methodIdx == 0 {
		return promptResponsesProviderManual()
	}
	return promptResponsesProviderFromURL(proxy)
}

func promptResponsesProviderManual() (providerPromptResult, error) {
	return promptResponsesProviderManualWith(bufio.NewScanner(os.Stdin), "", "", "")
}

func promptResponsesProviderManualWith(in *bufio.Scanner, baseURL, keyEnv, apiKey string) (providerPromptResult, error) {
	return promptResponsesProviderManualWithName(in, baseURL, "", keyEnv, apiKey)
}

func promptResponsesProviderManualWithName(in *bufio.Scanner, baseURL, providerName, keyEnv, apiKey string) (providerPromptResult, error) {
	fmt.Println()
	if baseURL == "" {
		baseURL = ask(in, os.Stdout, i18n.M.CustomPromptBaseURL, "")
		if baseURL == "" {
			return providerPromptResult{}, fmt.Errorf("base URL is required")
		}
	}
	if providerName == "" {
		providerName = promptProviderName(in, os.Stdout, providerSlug("responses", baseURL))
	}
	modelName := ask(in, os.Stdout, i18n.M.CustomPromptModel, "")
	if modelName == "" {
		return providerPromptResult{}, fmt.Errorf("model name is required")
	}
	if keyEnv == "" {
		keyEnv = promptAPIKeyEnvName(in, os.Stdout, i18n.M.CustomPromptKeyEnv, apiKeyEnvFromProviderName(providerName))
	} else if !config.IsValidCredentialKey(keyEnv) {
		return providerPromptResult{}, fmt.Errorf("invalid API key variable name %q", keyEnv)
	}
	if apiKey == "" {
		apiKey = ask(in, os.Stdout, i18n.M.CustomPromptAPIKey, "")
	}
	entry := newResponsesSetupEntry(providerName, baseURL, []string{modelName}, keyEnv, askContextWindow(in, os.Stdout))
	fmt.Printf("  %s\n", green(fmt.Sprintf(i18n.M.CustomAddedFmt, entry.Name+"/"+modelName)))
	return newProviderPromptResult([]config.ProviderEntry{entry}, keyEnv, apiKey), nil
}

func promptResponsesProviderFromURL(proxy netclient.ProxySpec) (providerPromptResult, error) {
	in := bufio.NewScanner(os.Stdin)
	fmt.Println()
	baseURL := ask(in, os.Stdout, i18n.M.CustomPromptBaseURL, "")
	if baseURL == "" {
		return providerPromptResult{}, fmt.Errorf("base URL is required")
	}
	providerName := promptProviderName(in, os.Stdout, providerSlug("responses", baseURL))
	keyEnv := promptAPIKeyEnvName(in, os.Stdout, i18n.M.CustomPromptKeyEnv, apiKeyEnvFromProviderName(providerName))
	apiKey := ask(in, os.Stdout, i18n.M.CustomPromptAPIKey, "")

	fmt.Printf("  %s\n", dim(fmt.Sprintf(i18n.M.FetchingModelsFmt, "responses")))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	models, err := fetchModelListCompat(ctx, baseURL, apiKey, proxy)
	if err != nil || len(models) == 0 {
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s\n", dim(fmt.Sprintf(i18n.M.FetchModelsFailedFmt, "responses", err)))
		} else {
			fmt.Fprintf(os.Stderr, "  %s\n", dim(i18n.M.CustomFetchEmpty))
		}
		return promptResponsesProviderManualWithName(in, baseURL, providerName, keyEnv, apiKey)
	}
	fmt.Printf("  %s\n", green(fmt.Sprintf(i18n.M.FetchModelsSuccessFmt, len(models), "responses")))

	items := make([]menuItem, len(models))
	for i, model := range models {
		items[i] = menuItem{name: model}
	}
	idxs, err := selectMany(fmt.Sprintf(i18n.M.SelectModelsLabel, "responses"), items)
	if err != nil || len(idxs) == 0 {
		return providerPromptResult{}, fmt.Errorf("no models selected")
	}
	selected := make([]string, 0, len(idxs))
	for _, idx := range idxs {
		selected = append(selected, models[idx])
	}
	entry := newResponsesSetupEntry(providerName, baseURL, selected, keyEnv, askContextWindow(in, os.Stdout))
	fmt.Printf("  %s\n", green(fmt.Sprintf(i18n.M.CustomAddedFmt, entry.Name+"/"+selected[0])))
	return newProviderPromptResult([]config.ProviderEntry{entry}, keyEnv, apiKey), nil
}

func newResponsesSetupEntry(name, baseURL string, models []string, keyEnv string, contextWindow int) config.ProviderEntry {
	entry := config.ProviderEntry{
		Name: name, Kind: "responses", BaseURL: baseURL, Models: append([]string(nil), models...),
		Default: models[0], APIKeyEnv: keyEnv, ContextWindow: contextWindow,
	}
	if _, ok := provider.OfficialOpenCodeGoRoute("responses", baseURL); !ok {
		return entry
	}
	preset, ok := config.CuratedProviderPreset("opencode-go-responses")
	if !ok || len(preset.Entries) != 1 {
		return entry
	}
	canonical := preset.Entries[0]
	entry.ResponsesMode = canonical.ResponsesMode
	entry.MaxOutputTokens = canonical.MaxOutputTokens
	entry.BillingMode = canonical.BillingMode
	entry.PresetID = preset.ID
	entry.PresetVersion = config.ProviderPresetVersion
	entry.ModelOverrides = make(map[string]config.ProviderModelOverride)
	for _, model := range models {
		if override, exists := canonical.ModelOverrides[model]; exists {
			entry.ModelOverrides[model] = override
		}
		if canonical.HasVisionModel(model) {
			entry.VisionModels = append(entry.VisionModels, model)
		}
	}
	return entry
}
