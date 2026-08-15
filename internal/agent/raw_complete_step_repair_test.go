package agent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"reasonix/internal/event"
	"reasonix/internal/provider"
	"reasonix/internal/tool"
)

func TestRawCompleteStepJSONRepairsOnceThroughToolChannel(t *testing.T) {
	todoWrite, _ := tool.LookupBuiltin("todo_write")
	completeStep, _ := tool.LookupBuiltin("complete_step")
	reg := tool.NewRegistry()
	reg.Add(todoWrite)
	reg.Add(completeStep)
	reg.Add(fakeTool{name: "aggregate_cases", readOnly: true})
	raw := `{"step_id":"analysis_01","result":"totals confirmed","evidence":[{"kind":"tool","tool":"aggregate_cases","summary":"query returned totals"}]}`
	prov := &scriptedProvider{name: "custom-provider", turns: [][]provider.Chunk{
		{toolCallChunk("todo", "todo_write", `{"todos":[{"content":"Analyze totals","status":"in_progress","step_id":"analysis_01"}]}`), {Type: provider.ChunkDone}},
		{toolCallChunk("query", "aggregate_cases", `{}`), {Type: provider.ChunkDone}},
		{{Type: provider.ChunkText, Text: raw}, {Type: provider.ChunkDone}},
		{toolCallChunk("signoff", "complete_step", raw), {Type: provider.ChunkDone}},
		{{Type: provider.ChunkText, Text: "done"}, {Type: provider.ChunkDone}},
	}}
	a := New(prov, reg, NewSession(""), Options{}, event.Discard)
	if err := a.Run(withClosedLoopContext(context.Background()), "analyze totals"); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if prov.call != 5 {
		t.Fatalf("provider calls = %d, want 5", prov.call)
	}
	if !sessionHasUserMessageContaining(a.sess.conversation, "structured complete_step tool call") {
		t.Fatal("missing corrective tool-channel instruction")
	}
	if got := lastToolResult(a.sess.conversation, "complete_step"); !strings.Contains(got, "host-verified 1") {
		t.Fatalf("complete_step result = %q", got)
	}
}

func TestRawCompleteStepJSONDoesNotLoop(t *testing.T) {
	todoWrite, _ := tool.LookupBuiltin("todo_write")
	completeStep, _ := tool.LookupBuiltin("complete_step")
	reg := tool.NewRegistry()
	reg.Add(todoWrite)
	reg.Add(completeStep)
	reg.Add(fakeTool{name: "aggregate_cases", readOnly: true})
	raw := `{"step":"Analyze totals","result":"totals confirmed","evidence":[{"kind":"tool","tool":"aggregate_cases","summary":"query returned totals"}]}`
	prov := &scriptedProvider{name: "custom-provider", turns: [][]provider.Chunk{
		{toolCallChunk("todo", "todo_write", `{"todos":[{"content":"Analyze totals","status":"in_progress"}]}`), {Type: provider.ChunkDone}},
		{toolCallChunk("query", "aggregate_cases", `{}`), {Type: provider.ChunkDone}},
		{{Type: provider.ChunkText, Text: raw}, {Type: provider.ChunkDone}},
		{{Type: provider.ChunkText, Text: raw}, {Type: provider.ChunkDone}},
	}}
	a := New(prov, reg, NewSession(""), Options{}, event.Discard)
	err := a.Run(withClosedLoopContext(context.Background()), "analyze totals")
	var readiness *FinalReadinessError
	if !errors.As(err, &readiness) {
		t.Fatalf("Run error = %v, want FinalReadinessError", err)
	}
	if prov.call != 4 {
		t.Fatalf("provider calls = %d, want one repair only", prov.call)
	}
}

func TestRawCompleteStepPayloadRejectsOrdinaryJSONAndProse(t *testing.T) {
	valid := `{"step":"Analyze","result":"done","evidence":[{"kind":"manual","summary":"checked"}]}`
	if !rawCompleteStepPayload(valid) {
		t.Fatal("strong complete_step payload was not recognized")
	}
	for _, text := range []string{
		`{"result":"data","items":[1,2]}`,
		`{"step":"Analyze","result":"done","evidence":[{"kind":"manual","summary":"checked"}],"answer":true}`,
		"Here is the result: " + valid,
	} {
		if rawCompleteStepPayload(text) {
			t.Fatalf("ordinary response was misclassified: %s", text)
		}
	}
}
