package cli

import (
	"testing"

	"reasonix/internal/config"
	"reasonix/internal/i18n"
)

func TestApplyServeLanguagePrefersDesktopSetting(t *testing.T) {
	t.Cleanup(func() { i18n.DetectLanguage("en") })

	i18n.DetectLanguage("en") // no CLI language, no locale
	cfg := &config.Config{Desktop: config.DesktopConfig{Language: "zh"}}
	applyServeLanguage(cfg)
	if got := i18n.CurrentLanguage(); got != "zh" {
		t.Fatalf("desktop zh: current = %q, want zh", got)
	}

	i18n.DetectLanguage("zh-TW")
	cfg = &config.Config{Desktop: config.DesktopConfig{Language: "en"}}
	applyServeLanguage(cfg)
	if got := i18n.CurrentLanguage(); got != "zh-TW" {
		t.Fatalf("existing non-en catalogue must not be downgraded: current = %q, want zh-TW", got)
	}

	i18n.DetectLanguage("en")
	cfg = &config.Config{Language: "en", Desktop: config.DesktopConfig{Language: "zh"}}
	applyServeLanguage(cfg)
	if got := i18n.CurrentLanguage(); got != "en" {
		t.Fatalf("pinned CLI language must win over desktop: current = %q, want en", got)
	}

	i18n.DetectLanguage("en")
	applyServeLanguage(&config.Config{})
	if got := i18n.CurrentLanguage(); got != "en" {
		t.Fatalf("no desktop setting must leave en: current = %q", got)
	}

	applyServeLanguage(nil)
	if got := i18n.CurrentLanguage(); got != "en" {
		t.Fatalf("nil config must be a no-op: current = %q", got)
	}
}
