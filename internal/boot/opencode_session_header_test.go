package boot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"reasonix/internal/config"
	"reasonix/internal/provider"
	_ "reasonix/internal/provider/responses"
)

func TestNewProviderSendsOpenCodeSessionHeaderAcrossRoutes(t *testing.T) {
	tests := []struct {
		name string
		kind string
	}{
		{name: "chat", kind: "openai"},
		{name: "messages", kind: "anthropic"},
		{name: "responses", kind: "responses"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			received := make(chan string, 1)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				received <- r.Header.Get("x-opencode-session")
				w.Header().Set("Content-Type", "text/event-stream")
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			p, err := NewProvider(&config.ProviderEntry{
				Name: "opencode-proxy", Kind: test.kind, BaseURL: srv.URL, Model: "model",
			})
			if err != nil {
				t.Fatalf("NewProvider: %v", err)
			}
			ctx := provider.WithSessionID(context.Background(), "session-a")
			stream, err := p.Stream(ctx, provider.Request{
				Messages: []provider.Message{{Role: provider.RoleUser, Content: "hi"}},
			})
			if err != nil {
				t.Fatalf("Stream: %v", err)
			}
			for range stream {
			}
			select {
			case got := <-received:
				if got != "session-a" {
					t.Fatalf("x-opencode-session = %q, want session-a", got)
				}
			case <-time.After(time.Second):
				t.Fatal("provider request was not received")
			}
		})
	}
}
