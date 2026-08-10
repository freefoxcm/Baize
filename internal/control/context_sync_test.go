package control

import (
	"context"
	"path/filepath"
	"testing"

	"reasonix/internal/agent"
	"reasonix/internal/event"
	"reasonix/internal/provider"
	"reasonix/internal/tool"
)

// TestResumeSyncsContextSnapshot: after Resume, ContextSnapshot must report
// the loaded session's estimated size, not the previous session's usage, so
// the serve gauge is correct immediately after switching sessions.
func TestResumeSyncsContextSnapshot(t *testing.T) {
	dir := t.TempDir()
	execProv := &recordingProvider{name: "executor", streams: [][]provider.Chunk{{
		{Type: provider.ChunkText, Text: "old done"},
		{Type: provider.ChunkUsage, Usage: &provider.Usage{PromptTokens: 12345, CompletionTokens: 678, TotalTokens: 13023}},
		{Type: provider.ChunkDone},
	}}}
	exec := agent.New(execProv, tool.NewRegistry(), agent.NewSession("exec sys"), agent.Options{ContextWindow: 1_000_000}, event.Discard)
	c := New(Options{Runner: exec, Executor: exec, SystemPrompt: "exec sys", SessionDir: dir, SessionPath: filepath.Join(dir, "old.jsonl"), Label: "test"})

	if err := c.Run(context.Background(), "old task"); err != nil {
		t.Fatal(err)
	}
	before, window := c.ContextSnapshot()
	if before == 0 {
		t.Fatal("expected non-zero context after a turn")
	}

	resumed := agent.NewSession("exec sys")
	resumed.Add(provider.Message{Role: provider.RoleUser, Content: "saved task gamma"})
	resumed.Add(provider.Message{Role: provider.RoleAssistant, Content: "saved answer gamma"})
	c.Resume(resumed, filepath.Join(dir, "resumed.jsonl"))

	used, gotWindow := c.ContextSnapshot()
	if used == 0 {
		t.Fatal("ContextSnapshot used = 0 after Resume")
	}
	if used >= before {
		t.Errorf("ContextSnapshot used = %d after Resume, want < pre-resume %d", used, before)
	}
	if gotWindow != window {
		t.Errorf("ContextSnapshot window = %d after Resume, want %d", gotWindow, window)
	}
}
