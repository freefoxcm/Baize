package agent

import (
	"context"

	"reasonix/internal/event"
	"reasonix/internal/provider"
)

type providerSessionRun func(context.Context, event.Sink, bool) (string, error)

func withProviderSession(ref string, run providerSessionRun) providerSessionRun {
	id := provider.EnsureSessionID(ref)
	return func(ctx context.Context, sink event.Sink, writerRegistered bool) (string, error) {
		return run(provider.WithSessionID(ctx, id), sink, writerRegistered)
	}
}
