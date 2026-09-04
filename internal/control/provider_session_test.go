package control

import (
	"context"
	"path/filepath"
	"testing"

	"reasonix/internal/provider"
)

type providerSessionRunner struct {
	sessionIDs []string
}

func (r *providerSessionRunner) Run(ctx context.Context, _ string) error {
	r.sessionIDs = append(r.sessionIDs, provider.SessionID(ctx))
	return nil
}

func TestTurnOrchestratorBindsStableProviderSessionID(t *testing.T) {
	runner := &providerSessionRunner{}
	c := New(Options{Runner: runner})
	c.SetFreshSessionPath(filepath.Join(t.TempDir(), "session-a.jsonl"))
	o := newTurnOrchestrator(c)
	for range 2 {
		if err := o.runTurnWithRawDisplay(context.Background(), "hi", "hi", ""); err != nil {
			t.Fatal(err)
		}
	}
	c.SetFreshSessionPath(filepath.Join(t.TempDir(), "session-b.jsonl"))
	if err := o.runTurnWithRawDisplay(context.Background(), "hi", "hi", ""); err != nil {
		t.Fatal(err)
	}
	if len(runner.sessionIDs) != 3 || runner.sessionIDs[0] != "session-a" || runner.sessionIDs[1] != "session-a" || runner.sessionIDs[2] != "session-b" {
		t.Fatalf("provider session ids = %v, want [session-a session-a session-b]", runner.sessionIDs)
	}
}
