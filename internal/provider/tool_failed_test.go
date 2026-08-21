package provider

import "testing"

func TestToolFailedIsDurableButProviderExcluded(t *testing.T) {
	stored := []Message{{Role: RoleTool, ToolCallID: "call-1", Name: "task", Content: "failed", ToolFailed: true}}
	model := ModelMessages(stored)
	if len(model) != 1 || model[0].ToolFailed {
		t.Fatalf("ModelMessages leaked ToolFailed: %+v", model)
	}
	projection := ProjectionMessages(stored)
	if len(projection) != 1 || !projection[0].ToolFailed {
		t.Fatalf("ProjectionMessages lost ToolFailed: %+v", projection)
	}
	if !stored[0].ToolFailed {
		t.Fatal("stored ToolFailed marker was mutated")
	}
}
