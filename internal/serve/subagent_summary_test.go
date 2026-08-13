package serve

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"reasonix/internal/event"
	"reasonix/internal/provider"
	"reasonix/internal/store"
)

func TestSubagentSummaryPersistsOnlyExecutionMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	recorder := newSubagentSummaryRecorder()
	recorder.observe(path, event.Event{Kind: event.ToolDispatch, Tool: event.Tool{
		ID: "task-1", Name: "task", Args: `{"task":"sensitive task prompt"}`,
		Profile: &event.Profile{Model: "local/model", Effort: "high"},
	}})
	// Partial and full dispatches for the same nested call count once.
	for _, partial := range []bool{true, false} {
		recorder.observe(path, event.Event{Kind: event.ToolDispatch, Tool: event.Tool{
			ID: "nested-1", ParentID: "task-1", Name: "use_capability",
			ResolvedName: "mcp__ipap__aggregate_cases", Partial: partial,
			Args: `{"token":"must-not-persist"}`,
		}})
	}
	recorder.observe(path, event.Event{Kind: event.ToolProgress, Tool: event.Tool{
		ID: "task-1", Name: event.SubagentProgressReasoningName, Output: "private live reasoning",
	}})
	recorder.observe(path, event.Event{Kind: event.ToolProgress, Tool: event.Tool{
		ID: "task-1", Name: event.SubagentProgressStatusName, Output: "completed", DurationMs: 1250,
	}})

	raw, err := os.ReadFile(store.SessionSubagentSummary(path))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"sensitive task prompt", "must-not-persist", "private live reasoning"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("summary persisted forbidden content %q: %s", forbidden, raw)
		}
	}
	var file subagentSummaryFile
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	got := file.Calls["task-1"]
	if got == nil || got.Status != "completed" || got.ToolCallCount != 1 || len(got.ToolNames) != 1 || got.ToolNames[0] != "aggregate_cases" {
		t.Fatalf("summary = %#v", got)
	}

	history := historyMessages([]provider.Message{{Role: provider.RoleAssistant, ToolCalls: []provider.ToolCall{{ID: "task-1", Name: "task"}}}}, recorder.summaries(path))
	if history[0].ToolCalls[0].SubagentSummary == nil || history[0].ToolCalls[0].SubagentSummary.DurationMs != 1250 {
		t.Fatalf("history summary = %#v", history[0].ToolCalls[0].SubagentSummary)
	}
}
