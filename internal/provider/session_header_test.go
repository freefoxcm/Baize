package provider

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestApplyOpenCodeSessionHeader(t *testing.T) {
	tests := []struct {
		name         string
		providerName string
		baseURL      string
		want         string
	}{
		{name: "go preset", providerName: "opencode-go", baseURL: "https://opencode.ai/zen/go/v1", want: "session-a"},
		{name: "proxy", providerName: "opencode-proxy", baseURL: "https://proxy.example/v1", want: "session-a"},
		{name: "renamed official route", providerName: "custom", baseURL: "https://opencode.ai/zen/go/v1", want: "session-a"},
		{name: "unrelated provider", providerName: "custom", baseURL: "https://example.com/v1"},
	}
	ctx := WithSessionID(context.Background(), " session-a ")
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			header := http.Header{"X-Opencode-Session": []string{"static"}}
			ApplyOpenCodeSessionHeader(ctx, header, test.providerName, test.baseURL)
			got := header.Get(openCodeSessionHeader)
			if test.want == "" {
				if got != "static" {
					t.Fatalf("unrelated provider header = %q, want configured value untouched", got)
				}
				return
			}
			if got != test.want {
				t.Fatalf("session header = %q, want %q", got, test.want)
			}
		})
	}
}

func TestApplyOpenCodeSessionHeaderOmitsMissingIdentity(t *testing.T) {
	header := http.Header{}
	ApplyOpenCodeSessionHeader(context.Background(), header, "opencode-proxy", "https://proxy.example/v1")
	if got := header.Get(openCodeSessionHeader); got != "" {
		t.Fatalf("session header = %q, want omitted", got)
	}
}

func TestEnsureSessionID(t *testing.T) {
	if got := EnsureSessionID(" session-a "); got != "session-a" {
		t.Fatalf("durable session id = %q, want session-a", got)
	}
	first, second := EnsureSessionID(""), EnsureSessionID("")
	if !strings.HasPrefix(first, "ses_") || !strings.HasPrefix(second, "ses_") || first == second {
		t.Fatalf("ephemeral session ids = %q/%q, want distinct ses_ values", first, second)
	}
}
