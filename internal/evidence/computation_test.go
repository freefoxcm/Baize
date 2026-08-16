package evidence

import (
	"encoding/json"
	"testing"

	"reasonix/internal/effectscope"
)

func computationReceipt(name, output string) Receipt {
	r := Receipt{ToolName: name, Success: true, EffectScope: effectscope.Scratch}
	r.ObserveOutput(output)
	return r
}

func TestComputationEvidenceMatchesCurrentStep(t *testing.T) {
	ledger := NewLedger()
	ledger.Record(Receipt{ToolName: "todo_write", Success: true})
	boundary := ledger.StepEvidenceBoundary()
	ledger.Record(computationReceipt("analyze_data", `{"total":12}`))
	if _, err := ledger.MatchSuccessfulComputationAfter("analyze_data", boundary); err != nil {
		t.Fatal(err)
	}
	ledger.Record(ReceiptFromToolCall("complete_step", json.RawMessage(`{"result":"done","evidence":[{"kind":"computation","tool":"analyze_data","summary":"calculated"}]}`), true, true))
	index, ok := ledger.LatestSuccessfulComputationIndex()
	if !ok || !ledger.HasSuccessfulComputationSignoffAfter(index) {
		t.Fatal("computation sign-off was not recognized")
	}
}

func TestComputationEvidenceRejectsFailedEmptyAndDurableReceipts(t *testing.T) {
	tests := []Receipt{
		{ToolName: "analyze_data", Success: false, EffectScope: effectscope.Scratch, OutputBytes: 4},
		{ToolName: "analyze_data", Success: true, EffectScope: effectscope.Scratch},
		{ToolName: "bash", Success: true, EffectScope: effectscope.Durable, OutputBytes: 4},
		{ToolName: "bash", Success: true, EffectScope: effectscope.Scratch, Mutation: true, OutputBytes: 4},
	}
	for _, receipt := range tests {
		ledger := NewLedger()
		ledger.Record(receipt)
		if _, ok := ledger.LatestSuccessfulComputationIndex(); ok {
			t.Fatalf("receipt should not count: %+v", receipt)
		}
	}
}

func TestComputationEvidenceCannotCrossStepBoundary(t *testing.T) {
	ledger := NewLedger()
	ledger.Record(computationReceipt("analyze_data", `{"total":12}`))
	ledger.Record(Receipt{ToolName: "todo_write", Success: true})
	if _, err := ledger.MatchSuccessfulComputationAfter("analyze_data", ledger.StepEvidenceBoundary()); err == nil {
		t.Fatal("a computation from before the current step must not be reusable")
	}
}

func TestScratchClassificationDoesNotWhitelistOrdinaryTmpWrites(t *testing.T) {
	normal, _ := ClassifyBashToolCall(json.RawMessage(`{"command":"python -c 'open(\"/tmp/a.py\",\"w\").write(\"x\")'"}`))
	if !normal.ContentMutation || normal.Scope != effectscope.Unknown {
		t.Fatalf("ordinary tmp classification = %+v", normal)
	}
	scratch, _ := ClassifyBashToolCall(json.RawMessage(`{"command":"python /tmp/a.py","execution_scope":"scratch"}`))
	if scratch.ContentMutation || scratch.Scope != effectscope.Unknown {
		t.Fatalf("scratch preflight classification = %+v", scratch)
	}
}
