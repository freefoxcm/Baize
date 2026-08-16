package evidence

import (
	"encoding/json"
	"strings"
	"testing"
)

func observationReceipt(name, output string, success, read bool) Receipt {
	r := ReceiptFromToolCall(name, json.RawMessage(`{}`), success, read)
	r.ObserveOutput(output)
	return r
}

func TestLedgerWorkReceiptExcludesWorkflowBookkeeping(t *testing.T) {
	for _, name := range []string{"ask", "todo_write", "complete_step", "update_goal"} {
		ledger := NewLedger()
		ledger.Record(ReceiptFromToolCall(name, json.RawMessage(`{}`), true, true))
		if ledger.HasSuccessfulWorkReceipt() {
			t.Fatalf("workflow tool %q was treated as host-observed task work", name)
		}
	}

	ledger := NewLedger()
	ledger.Record(ReceiptFromToolCall("read_file", json.RawMessage(`{"path":"main.go"}`), true, true))
	if !ledger.HasSuccessfulWorkReceipt() {
		t.Fatal("successful read was not treated as host-observed task work")
	}
}

func TestMatchSuccessfulReadToolAfter(t *testing.T) {
	ledger := NewLedger()
	ledger.Record(Receipt{ToolName: "todo_write", Success: true})
	boundary := ledger.StepEvidenceBoundary()
	ledger.Record(observationReceipt("mcp__ipap__aggregate_cases", "42", true, true))
	if got, err := ledger.MatchSuccessfulReadToolAfter("aggregate_cases", boundary); err != nil || got != "mcp__ipap__aggregate_cases" {
		t.Fatalf("match = %q, %v", got, err)
	}

	ledger.Record(observationReceipt("mcp__archive__aggregate_cases", "12", true, true))
	if _, err := ledger.MatchSuccessfulReadToolAfter("aggregate_cases", boundary); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("ambiguous match error = %v", err)
	}
	if got, err := ledger.MatchSuccessfulReadToolAfter("mcp__ipap__aggregate_cases", boundary); err != nil || got == "" {
		t.Fatalf("exact match = %q, %v", got, err)
	}
}

func TestMatchSuccessfulReadToolRejectsStaleFailedEmptyAndWriter(t *testing.T) {
	tests := []struct {
		name    string
		receipt Receipt
		stale   bool
	}{
		{name: "failed", receipt: observationReceipt("query", "data", false, true)},
		{name: "empty", receipt: observationReceipt("query", "", true, true)},
		{name: "writer", receipt: observationReceipt("query", "data", true, false)},
		{name: "workflow", receipt: observationReceipt("update_goal", "complete", true, true)},
		{name: "stale", receipt: observationReceipt("query", "data", true, true), stale: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ledger := NewLedger()
			if tc.stale {
				ledger.Record(tc.receipt)
				ledger.Record(Receipt{ToolName: "todo_write", Success: true})
			} else {
				ledger.Record(Receipt{ToolName: "todo_write", Success: true})
				ledger.Record(tc.receipt)
			}
			if _, err := ledger.MatchSuccessfulReadToolAfter("query", ledger.StepEvidenceBoundary()); err == nil {
				t.Fatal("invalid receipt satisfied tool evidence")
			}
		})
	}
}

func TestHasSuccessfulObservationSignoffAfter(t *testing.T) {
	ledger := NewLedger()
	ledger.Record(Receipt{ToolName: "todo_write", Success: true})
	ledger.Record(observationReceipt("mcp__ipap__aggregate_cases", "42", true, true))
	ledger.Record(ReceiptFromToolCall("complete_step", json.RawMessage(`{
		"step":"Analyze", "result":"done",
		"evidence":[{"kind":"tool","tool":"aggregate_cases","summary":"totals returned"}]
	}`), true, true))
	if !ledger.HasSuccessfulObservationSignoffAfter(-1) {
		t.Fatal("host-backed tool evidence did not satisfy observation signoff")
	}
}
