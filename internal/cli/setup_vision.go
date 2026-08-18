package cli

import (
	"fmt"
	"strings"

	"reasonix/internal/config"
	"reasonix/internal/i18n"
)

func promptProviderVision(entry config.ProviderEntry) (config.ProviderEntry, error) {
	models := entry.ModelList()
	if !config.CanConfigureVision(&entry) {
		fmt.Printf("  %s\n", dim(fmt.Sprintf(i18n.M.SetupVisionUnsupported, entry.Name)))
		return applyProviderVisionSelection(entry, nil), nil
	}

	defaults := providerVisionDefaults(entry, models)
	checked := make([]bool, len(models))
	items := make([]menuItem, len(models))
	for i, model := range models {
		items[i] = menuItem{name: model}
		checked[i] = containsFold(defaults, model)
	}
	idxs, err := selectManyWithOptions(
		fmt.Sprintf(i18n.M.SetupVisionModelsFmt, entry.Name),
		items,
		selectManyOptions{checked: checked, allowEmpty: true},
	)
	if err != nil {
		return entry, err
	}
	selected := make([]string, 0, len(idxs))
	for _, idx := range idxs {
		selected = append(selected, models[idx])
	}
	return applyProviderVisionSelection(entry, selected), nil
}

func providerVisionDefaults(entry config.ProviderEntry, models []string) []string {
	if entry.Vision {
		return append([]string(nil), models...)
	}
	if entry.VisionModels == nil {
		return config.InferVisionModels(models)
	}
	selected := make([]string, 0, len(entry.VisionModels))
	for _, model := range models {
		if containsFold(entry.VisionModels, model) {
			selected = append(selected, model)
		}
	}
	return selected
}

func applyProviderVisionSelection(entry config.ProviderEntry, selected []string) config.ProviderEntry {
	entry.Vision = false
	entry.VisionModels = append([]string{}, selected...)
	return entry
}

func containsFold(values []string, want string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(want)) {
			return true
		}
	}
	return false
}
