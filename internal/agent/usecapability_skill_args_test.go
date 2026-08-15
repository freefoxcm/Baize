package agent

import (
	"context"
	"encoding/json"
	"testing"

	"reasonix/internal/tool"
)

func TestUseCapabilitySkillCallWrapsFreeformArguments(t *testing.T) {
	reg := tool.NewRegistry()
	reg.Add(fakeTool{name: "run_skill", readOnly: true})
	tl := NewUseCapabilityTool(context.Background(), nil, nil, reg, nil, nil, nil)

	resolved, err := tl.ResolveCall(context.Background(), json.RawMessage(`{"action":"call","capability_id":"skill:ipap-person-analysis","arguments":"analyze the selected person"}`))
	if err != nil {
		t.Fatal(err)
	}
	if resolved.TargetName != "run_skill" {
		t.Fatalf("target = %q, want run_skill", resolved.TargetName)
	}
	var got struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	}
	if err := json.Unmarshal(resolved.Args, &got); err != nil {
		t.Fatalf("resolved args = %s: %v", resolved.Args, err)
	}
	if got.Name != "ipap-person-analysis" || got.Arguments != "analyze the selected person" {
		t.Fatalf("resolved args = %+v", got)
	}
}
