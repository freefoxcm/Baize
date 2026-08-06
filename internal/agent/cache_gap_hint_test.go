package agent

import (
	"context"
	"testing"
	"time"

	"reasonix/internal/event"
	"reasonix/internal/provider"
	"reasonix/internal/tool"
)

// gapCaptureProvider records whether/what cache-gap hint arrived on each
// Stream call, so tests can prove Agent.stream threads it into Provider.Stream
// and freezes it across body retries.
type gapCaptureProvider struct {
	gaps          []time.Duration
	ok            []bool
	interruptNext bool
}

func (p *gapCaptureProvider) Name() string { return "gap-capture" }

func (p *gapCaptureProvider) Stream(ctx context.Context, _ provider.Request) (<-chan provider.Chunk, error) {
	gap, ok := provider.CacheGapHintFromContext(ctx)
	p.gaps = append(p.gaps, gap)
	p.ok = append(p.ok, ok)
	ch := make(chan provider.Chunk, 2)
	if p.interruptNext {
		p.interruptNext = false
		ch <- provider.Chunk{Type: provider.ChunkError, Err: provider.StreamInterrupt(context.Canceled, provider.StreamInterruptConnectionReset)}
		close(ch)
		return ch, nil
	}
	ch <- provider.Chunk{Type: provider.ChunkText, Text: "ok"}
	ch <- provider.Chunk{Type: provider.ChunkUsage, Usage: &provider.Usage{PromptTokens: 10, CompletionTokens: 1, TotalTokens: 11}}
	close(ch)
	return ch, nil
}

func TestStreamRecoveryFreezesCacheGapHint(t *testing.T) {
	prov := &gapCaptureProvider{}
	a := New(prov, tool.NewRegistry(), NewSession("sys"), Options{}, event.Discard)
	if err := a.Run(context.Background(), "prime"); err != nil {
		t.Fatalf("prime Run: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	prov.interruptNext = true
	if err := a.Run(context.Background(), "retry"); err != nil {
		t.Fatalf("recovered Run: %v", err)
	}

	if len(prov.gaps) != 3 {
		t.Fatalf("Stream calls = %d, want prime + failed attempt + retry", len(prov.gaps))
	}
	if !prov.ok[1] || !prov.ok[2] || prov.gaps[1] <= 0 {
		t.Fatalf("retry round hints = ok %v/%v gaps %v/%v", prov.ok[1], prov.ok[2], prov.gaps[1], prov.gaps[2])
	}
	if prov.gaps[2] != prov.gaps[1] {
		t.Fatalf("retry cache gap changed provider-visible TTL input: first=%v retry=%v", prov.gaps[1], prov.gaps[2])
	}
}

func TestStreamCarriesCacheGapHintAfterFirstTurn(t *testing.T) {
	prov := &gapCaptureProvider{}
	a := New(prov, tool.NewRegistry(), NewSession("sys"), Options{}, event.Discard)
	if err := a.Run(context.Background(), "one"); err != nil {
		t.Fatalf("first Run: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	if err := a.Run(context.Background(), "two"); err != nil {
		t.Fatalf("second Run: %v", err)
	}

	if len(prov.ok) != 2 {
		t.Fatalf("got %d Stream calls, want 2", len(prov.ok))
	}
	if prov.ok[0] {
		t.Fatalf("first request should carry no gap hint, got gap=%v", prov.gaps[0])
	}
	if !prov.ok[1] || prov.gaps[1] <= 0 {
		t.Fatalf("second request should carry a positive gap hint, got ok=%v gap=%v", prov.ok[1], prov.gaps[1])
	}
}
