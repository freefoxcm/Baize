package cli

import (
	"reflect"
	"testing"

	"reasonix/internal/config"
)

func TestProviderVisionDefaults(t *testing.T) {
	models := []string{"plain-chat", "qwen-vl-plus", "new-vision", "stale-safe"}
	tests := []struct {
		name  string
		entry config.ProviderEntry
		want  []string
	}{
		{
			name:  "unconfigured infers conservative defaults",
			entry: config.ProviderEntry{},
			want:  []string{"qwen-vl-plus", "new-vision"},
		},
		{
			name:  "legacy provider-wide vision selects all",
			entry: config.ProviderEntry{Vision: true},
			want:  models,
		},
		{
			name:  "explicit selection filters stale models case-insensitively",
			entry: config.ProviderEntry{VisionModels: []string{"NEW-VISION", "removed-model"}},
			want:  []string{"new-vision"},
		},
		{
			name:  "explicit empty remains empty",
			entry: config.ProviderEntry{VisionModels: []string{}},
			want:  []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := providerVisionDefaults(tt.entry, models)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("providerVisionDefaults() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestApplyProviderVisionSelectionNormalizesProviderWideFlag(t *testing.T) {
	entry := config.ProviderEntry{Vision: true, VisionModels: []string{"old"}}
	got := applyProviderVisionSelection(entry, []string{"image-a", "image-b"})
	if got.Vision {
		t.Fatal("provider-wide vision remained enabled")
	}
	if !reflect.DeepEqual(got.VisionModels, []string{"image-a", "image-b"}) {
		t.Fatalf("vision models = %#v", got.VisionModels)
	}

	empty := applyProviderVisionSelection(entry, nil)
	if empty.VisionModels == nil || len(empty.VisionModels) != 0 {
		t.Fatalf("empty selection = %#v, want explicit empty slice", empty.VisionModels)
	}
}

func TestOfficialDeepSeekTextModelRemainsTextOnly(t *testing.T) {
	entry := config.ProviderEntry{
		Name: "deepseek", BaseURL: "https://api.deepseek.com", Model: "deepseek-v4-flash",
		Vision: true, VisionModels: []string{"deepseek-v4-flash"},
	}
	if config.EffectiveVision(&entry) {
		t.Fatal("official DeepSeek text model accepted image input")
	}
}

func TestSelectedManyIndicesAllowsExplicitEmptyWithoutChangingDefaultContract(t *testing.T) {
	if got, ok := selectedManyIndices([]bool{false, false}, false); ok || got != nil {
		t.Fatalf("required selection = %#v, %v", got, ok)
	}
	if got, ok := selectedManyIndices([]bool{false, false}, true); !ok || got != nil {
		t.Fatalf("allowed empty selection = %#v, %v", got, ok)
	}
	if got, ok := selectedManyIndices([]bool{true, false, true}, false); !ok || !reflect.DeepEqual(got, []int{0, 2}) {
		t.Fatalf("selected indices = %#v, %v", got, ok)
	}
}
