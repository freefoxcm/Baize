package provider

import (
	"context"
	"time"
)

type cacheGapHintKey struct{}

// WithCacheGapHint attaches how long it has been since this Agent last sent a
// provider request, so a provider with an opt-in longer cache TTL (Anthropic's
// 1h ephemeral breakpoint, versus its cheaper default 5m one) can decide
// whether the default window would likely already have expired. Mirrors the
// WithRetryNotify/WithRequestAttemptCounter context-key idiom in retry.go: a
// side channel for per-request signals that must never enter the request body
// or the cacheable prefix itself.
func WithCacheGapHint(ctx context.Context, gap time.Duration) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, cacheGapHintKey{}, gap)
}

// CacheGapHintFromContext returns the gap since the previous request and
// whether one was set. No value — the first request of an Agent's lifetime, a
// fresh sub-agent session, or a caller that never wired the hint — means
// default to the provider's cheaper, shorter cache TTL.
func CacheGapHintFromContext(ctx context.Context) (time.Duration, bool) {
	if ctx == nil {
		return 0, false
	}
	gap, ok := ctx.Value(cacheGapHintKey{}).(time.Duration)
	return gap, ok
}
