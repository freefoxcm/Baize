package serve

import "context"

// controllerLifecycleContext preserves request values without letting an HTTP
// disconnect cancel resources owned by the replacement controller.
func controllerLifecycleContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return context.WithoutCancel(ctx)
}
