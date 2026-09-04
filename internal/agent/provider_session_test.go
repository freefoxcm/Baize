package agent

import (
	"context"
	"strings"
	"testing"

	"reasonix/internal/event"
	"reasonix/internal/provider"
)

func TestWithProviderSessionUsesTranscriptRef(t *testing.T) {
	var got string
	run := withProviderSession("subagent-ref", func(ctx context.Context, _ event.Sink, _ bool) (string, error) {
		got = provider.SessionID(ctx)
		return "", nil
	})
	if _, err := run(context.Background(), nil, false); err != nil {
		t.Fatal(err)
	}
	if got != "subagent-ref" {
		t.Fatalf("provider session id = %q, want subagent-ref", got)
	}
}

func TestWithProviderSessionKeepsGeneratedIDStable(t *testing.T) {
	var got []string
	run := withProviderSession("", func(ctx context.Context, _ event.Sink, _ bool) (string, error) {
		got = append(got, provider.SessionID(ctx))
		return "", nil
	})
	for range 2 {
		if _, err := run(context.Background(), nil, false); err != nil {
			t.Fatal(err)
		}
	}
	if len(got) != 2 || !strings.HasPrefix(got[0], "ses_") || got[0] != got[1] {
		t.Fatalf("provider session ids = %v, want one stable generated ses_ id", got)
	}
}
