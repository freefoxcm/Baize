package control

import (
	"context"
	"strings"
	"testing"

	"reasonix/internal/agent"
	"reasonix/internal/event"
)

func TestTurnOrchestratorRefTurnSeparatesRawDisplayFromModelTask(t *testing.T) {
	const task = "Analyze the attached image."
	const raw = "@[diagram.png](.reasonix/attachments/upload-diagram.png)"
	const refs = "@.reasonix/attachments/upload-diagram.png"
	sess := agent.NewSession("sys")
	runner := &recordingSessionRunner{session: sess}
	c := New(Options{Runner: runner, Executor: agent.New(nil, nil, sess, agent.Options{}, event.Discard)})
	resolve := func(context.Context, string) (string, []string) {
		return `<image path=".reasonix/attachments/upload-diagram.png"></image>`, nil
	}

	if err := c.runRefTurnWithResolverSync(context.Background(), task, raw, refs, raw, "", resolve); err != nil {
		t.Fatal(err)
	}
	if len(runner.inputs) != 1 || !strings.Contains(runner.inputs[0], task) {
		t.Fatalf("provider input = %+v, want hidden attachment task", runner.inputs)
	}
	if len(runner.raw) != 1 || runner.raw[0] != raw {
		t.Fatalf("persisted raw input = %+v, want display reference %q", runner.raw, raw)
	}
}
