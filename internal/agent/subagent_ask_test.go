package agent

import (
	"context"
	"strings"
	"testing"

	"reasonix/internal/event"
	"reasonix/internal/provider"
	"reasonix/internal/tool"
)

func TestRunSubAgentWithSessionInheritsCallAskerWhenEnabled(t *testing.T) {
	prov := &scriptedProvider{name: "sub", turns: [][]provider.Chunk{
		{
			toolCallChunk("ask-1", "ask", `{"questions":[{"header":"Scope","question":"Which scope?","options":[{"label":"Keep going"},{"label":"Stop"}]}]}`),
			{Type: provider.ChunkDone},
		},
		{{Type: provider.ChunkText, Text: "confirmed"}, {Type: provider.ChunkDone}},
	}}
	reg := tool.NewRegistry()
	reg.Add(NewAskTool())
	asker := &recordingAsker{}
	ctx := withCallContext(context.Background(), "parent-call", event.Discard, asker, false)
	sess := NewSession("sys")

	answer, err := RunSubAgentWithSession(ctx, prov, reg, sess, "prepare report",
		Options{SubagentDepth: 1, InheritCallAsker: true}, event.Discard)
	if err != nil {
		t.Fatalf("RunSubAgentWithSession: %v", err)
	}
	if answer != "confirmed" {
		t.Fatalf("answer = %q, want confirmed", answer)
	}
	if len(asker.questions) != 1 || asker.questions[0].Prompt != "Which scope?" {
		t.Fatalf("questions = %+v, want inherited interactive ask", asker.questions)
	}
	if got := lastToolResult(sess, "ask"); !strings.Contains(got, "Scope: Keep going") {
		t.Fatalf("ask result = %q, want user answer", got)
	}
}

func TestRunSubAgentWithSessionKeepsAskerHeadlessByDefault(t *testing.T) {
	prov := &scriptedProvider{name: "sub", turns: [][]provider.Chunk{
		{
			toolCallChunk("ask-1", "ask", `{"questions":[{"header":"Scope","question":"Which scope?","options":[{"label":"Keep going"},{"label":"Stop"}]}]}`),
			{Type: provider.ChunkDone},
		},
		{{Type: provider.ChunkText, Text: "used fallback"}, {Type: provider.ChunkDone}},
	}}
	reg := tool.NewRegistry()
	reg.Add(NewAskTool())
	asker := &recordingAsker{}
	ctx := withCallContext(context.Background(), "parent-call", event.Discard, asker, false)
	sess := NewSession("sys")

	answer, err := RunSubAgentWithSession(ctx, prov, reg, sess, "prepare report",
		Options{SubagentDepth: 1}, event.Discard)
	if err != nil {
		t.Fatalf("RunSubAgentWithSession: %v", err)
	}
	if answer != "used fallback" {
		t.Fatalf("answer = %q, want used fallback", answer)
	}
	if len(asker.questions) != 0 {
		t.Fatalf("parent asker received %d question(s), want none", len(asker.questions))
	}
	if got := lastToolResult(sess, "ask"); !strings.Contains(got, "No interactive user answered") {
		t.Fatalf("ask result = %q, want headless fallback", got)
	}
}
