package serve

import (
	"context"
	"testing"
	"time"

	"reasonix/internal/history"
	"reasonix/internal/stats"
)

// isolateServeHome fences process-wide SQLite projections around a temporary
// REASONIX_HOME. Windows cannot remove the directory while a catalog is open.
func isolateServeHome(t *testing.T, home string) {
	t.Helper()
	closeServeTestCatalogs(t)
	t.Setenv("REASONIX_HOME", home)
	t.Cleanup(func() { closeServeTestCatalogs(t) })
}

func closeServeTestCatalogs(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := history.CloseSharedCatalog(ctx); err != nil {
		t.Fatalf("close shared history catalog: %v", err)
	}
	if err := stats.CloseUsageCatalogs(ctx); err != nil {
		t.Fatalf("close shared usage catalogs: %v", err)
	}
}
