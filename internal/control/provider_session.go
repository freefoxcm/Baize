package control

import (
	"context"

	"reasonix/internal/agent"
	"reasonix/internal/jobs"
	"reasonix/internal/provider"
)

func withProviderSession(ctx context.Context, parentSession string) context.Context {
	ctx = provider.WithSessionID(ctx, provider.EnsureSessionID(parentSession))
	ctx = agent.WithParentSession(ctx, parentSession)
	return jobs.WithSession(ctx, parentSession)
}
