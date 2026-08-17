package agent

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"reasonix/internal/event"
	"reasonix/internal/provider"
	"reasonix/internal/tool"
)

func TestClosedLoopAllowsHostBackedReadOnlyAnalysis(t *testing.T) {
	todoWrite, ok := tool.LookupBuiltin("todo_write")
	if !ok {
		t.Fatal("todo_write builtin not registered")
	}
	completeStep, ok := tool.LookupBuiltin("complete_step")
	if !ok {
		t.Fatal("complete_step builtin not registered")
	}
	reg := tool.NewRegistry()
	reg.Add(fakeTool{name: "read_file", readOnly: true})
	reg.Add(todoWrite)
	reg.Add(completeStep)
	prov := &scriptedProvider{name: "delivery", turns: [][]provider.Chunk{
		{toolCallChunk("todo", "todo_write", `{"todos":[{"content":"Analyze main.go","status":"in_progress","step_id":"analysis_01"}]}`), {Type: provider.ChunkDone}},
		{toolCallChunk("read", "read_file", `{"path":"main.go"}`), {Type: provider.ChunkDone}},
		{toolCallChunk("signoff", "complete_step", `{"step_id":"analysis_01","result":"analysis complete","evidence":[{"kind":"tool","tool":"read_file","summary":"main.go was inspected"}]}`), {Type: provider.ChunkDone}},
		{{Type: provider.ChunkText, Text: "analysis"}, {Type: provider.ChunkDone}},
	}}
	a := New(prov, reg, NewSession(""), Options{}, event.Discard)
	if err := a.Run(withClosedLoopContext(context.Background()), "analyze main.go"); err != nil {
		t.Fatalf("read-only analysis should close with host-backed tool evidence: %v", err)
	}
	if got := lastToolResult(a.sess.conversation, "complete_step"); !strings.Contains(got, "host-verified 1") {
		t.Fatalf("complete_step result = %q", got)
	}
}

func TestClosedLoopReadOnlyGapRequestsSignoffNotReview(t *testing.T) {
	todoWrite, _ := tool.LookupBuiltin("todo_write")
	reg := tool.NewRegistry()
	reg.Add(fakeTool{name: "aggregate_cases", readOnly: true})
	reg.Add(todoWrite)
	prov := &scriptedProvider{name: "delivery", turns: [][]provider.Chunk{
		{toolCallChunk("todo", "todo_write", `{"todos":[{"content":"Aggregate cases","status":"in_progress"}]}`), {Type: provider.ChunkDone}},
		{toolCallChunk("read", "aggregate_cases", `{}`), {Type: provider.ChunkDone}},
		{{Type: provider.ChunkText, Text: "analysis"}, {Type: provider.ChunkDone}},
	}}
	a := New(prov, reg, NewSession(""), Options{}, event.Discard)
	err := a.Run(withClosedLoopContext(context.Background()), "analyze case totals")
	var readiness *FinalReadinessError
	if !errors.As(err, &readiness) {
		t.Fatalf("Run error = %v, want FinalReadinessError", err)
	}
	if !slices.Contains(readiness.Missing, "todo") || slices.Contains(readiness.Missing, "review") {
		t.Fatalf("missing = %v, want todo sign-off without review", readiness.Missing)
	}
	if strings.Contains(readiness.Reason, "git diff") || strings.Contains(readiness.Reason, "reviewed_paths") {
		t.Fatalf("read-only readiness leaked mutation review guidance: %s", readiness.Reason)
	}
}

func TestClosedLoopAllowsHostBackedComputation(t *testing.T) {
	todoWrite, _ := tool.LookupBuiltin("todo_write")
	completeStep, _ := tool.LookupBuiltin("complete_step")
	reg := tool.NewRegistry()
	reg.Add(fakeTool{name: "analyze_data", readOnly: true})
	reg.Add(todoWrite)
	reg.Add(completeStep)
	prov := &scriptedProvider{name: "delivery", turns: [][]provider.Chunk{
		{toolCallChunk("todo", "todo_write", `{"todos":[{"content":"Calculate overlap","status":"in_progress","step_id":"analysis_01"}]}`), {Type: provider.ChunkDone}},
		{toolCallChunk("compute", "analyze_data", `{}`), {Type: provider.ChunkDone}},
		{toolCallChunk("signoff", "complete_step", `{"step_id":"analysis_01","result":"calculation complete","evidence":[{"kind":"computation","tool":"analyze_data","summary":"overlap calculated"}]}`), {Type: provider.ChunkDone}},
		{{Type: provider.ChunkText, Text: "analysis"}, {Type: provider.ChunkDone}},
	}}
	a := New(prov, reg, NewSession(""), Options{}, event.Discard)
	if err := a.Run(withClosedLoopContext(context.Background()), "calculate overlap"); err != nil {
		t.Fatalf("computation should close with host-backed evidence: %v", err)
	}
}

func TestClosedLoopComputationGapRequestsSignoffNotReview(t *testing.T) {
	todoWrite, _ := tool.LookupBuiltin("todo_write")
	reg := tool.NewRegistry()
	reg.Add(fakeTool{name: "analyze_data", readOnly: true})
	reg.Add(todoWrite)
	prov := &scriptedProvider{name: "delivery", turns: [][]provider.Chunk{
		{toolCallChunk("todo", "todo_write", `{"todos":[{"content":"Calculate overlap","status":"in_progress"}]}`), {Type: provider.ChunkDone}},
		{toolCallChunk("compute", "analyze_data", `{}`), {Type: provider.ChunkDone}},
		{{Type: provider.ChunkText, Text: "analysis"}, {Type: provider.ChunkDone}},
	}}
	a := New(prov, reg, NewSession(""), Options{}, event.Discard)
	err := a.Run(withClosedLoopContext(context.Background()), "calculate overlap")
	var readiness *FinalReadinessError
	if !errors.As(err, &readiness) {
		t.Fatalf("Run error = %v, want FinalReadinessError", err)
	}
	if !slices.Contains(readiness.Missing, "todo") || slices.Contains(readiness.Missing, "review") {
		t.Fatalf("missing = %v, want todo sign-off without review", readiness.Missing)
	}
}
