package cli

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"reasonix/internal/i18n"
)

func TestPromptProviderNameUsesDefaultAndRejectsReservedOrAmbiguousNames(t *testing.T) {
	i18n.DetectLanguage("en")

	t.Run("default", func(t *testing.T) {
		var out bytes.Buffer
		got := promptProviderName(bufio.NewScanner(strings.NewReader("\n")), &out, "custom-opencode-ai")
		if got != "custom-opencode-ai" {
			t.Fatalf("provider name = %q, want generated default", got)
		}
	})

	t.Run("validation", func(t *testing.T) {
		var out bytes.Buffer
		got := promptProviderName(
			bufio.NewScanner(strings.NewReader("custom\nanthropic\nbad/name\nopencode-go\n")),
			&out,
			"custom-opencode-ai",
		)
		if got != "opencode-go" {
			t.Fatalf("provider name = %q, want opencode-go", got)
		}
		if count := strings.Count(out.String(), "is invalid"); count != 3 {
			t.Fatalf("validation messages = %d, want 3: %q", count, out.String())
		}
	})
}

func TestPromptCustomProviderManualAcceptsCustomName(t *testing.T) {
	result, err := promptCustomProviderManualWith(
		bufio.NewScanner(strings.NewReader("opencode-go\ndeepseek-v4-flash\n\n\n\n")),
		"https://opencode.ai/zen/go/v1",
		"",
		"",
	)
	if err != nil {
		t.Fatalf("promptCustomProviderManualWith: %v", err)
	}
	entry := result.entries[0]
	if entry.Name != "opencode-go" {
		t.Fatalf("Name = %q, want opencode-go", entry.Name)
	}
	if entry.APIKeyEnv != "OPENCODE_GO_API_KEY" {
		t.Fatalf("APIKeyEnv = %q, want OPENCODE_GO_API_KEY", entry.APIKeyEnv)
	}
}

func TestPromptAnthropicProviderManualAcceptsCustomName(t *testing.T) {
	result, err := promptAnthropicProviderManualWith(
		bufio.NewScanner(strings.NewReader("opencode-go-anthropic\nqwen3.7-plus\n\n\n\n")),
		"https://opencode.ai/zen/go",
		"",
		"",
	)
	if err != nil {
		t.Fatalf("promptAnthropicProviderManualWith: %v", err)
	}
	entry := result.entries[0]
	if entry.Name != "opencode-go-anthropic" {
		t.Fatalf("Name = %q, want opencode-go-anthropic", entry.Name)
	}
	if entry.APIKeyEnv != "ANTHROPIC_API_KEY" {
		t.Fatalf("APIKeyEnv = %q, want ANTHROPIC_API_KEY", entry.APIKeyEnv)
	}
}
