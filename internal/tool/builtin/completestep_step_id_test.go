package builtin

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"reasonix/internal/evidence"
)

func TestCompleteStepPendingParentReportsSignableSubstepID(t *testing.T) {
	todos := []evidence.TodoItem{
		{Content: "确认报告呈现方式与提纲", Status: "pending", StepID: "plan_step_01"},
		{Content: "按技能第二阶段确认报告呈现方式", Status: "in_progress", Level: 1, StepID: "plan_step_02"},
	}
	ctx := evidence.WithTodoState(context.Background(), todos)

	_, err := (completeStep{}).Execute(ctx, json.RawMessage(`{
		"step_id":"plan_step_01",
		"result":"confirmed",
		"evidence":[{"kind":"manual","summary":"user answered"}]
	}`))
	if err == nil {
		t.Fatal("pending parent sign-off should be rejected")
	}
	for _, want := range []string{`Current signable todo: 2 "按技能第二阶段确认报告呈现方式"`, `step_id "plan_step_02"`, `Retry complete_step with step_id "plan_step_02"`} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("pending parent error %q missing %q", err, want)
		}
	}

	out, err := (completeStep{}).Execute(ctx, json.RawMessage(`{
		"step_id":"plan_step_02",
		"result":"confirmed",
		"evidence":[{"kind":"manual","summary":"user answered"}]
	}`))
	if err != nil {
		t.Fatalf("current sub-step id rejected: %v", err)
	}
	if !strings.Contains(out, `todo-matched 2 ("按技能第二阶段确认报告呈现方式")`) {
		t.Fatalf("sub-step output = %q", out)
	}
}
