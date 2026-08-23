package provider

import "testing"

func TestToolFailedIsDurableButProviderExcluded(t *testing.T) {
	stored := []Message{{
		Role: RoleTool, ToolCallID: "call-1", Name: "task", Content: "failed", ToolFailed: true,
		VisionSummary: &VisionSummary{Version: 1, Summary: "local image summary"},
	}}
	model := ModelMessages(stored)
	if len(model) != 1 || model[0].ToolFailed || model[0].VisionSummary != nil {
		t.Fatalf("ModelMessages leaked local metadata: %+v", model)
	}
	projection := ProjectionMessages(stored)
	if len(projection) != 1 || !projection[0].ToolFailed || projection[0].VisionSummary != nil {
		t.Fatalf("ProjectionMessages local metadata = %+v", projection)
	}
	if !stored[0].ToolFailed || stored[0].VisionSummary == nil {
		t.Fatal("stored local metadata was mutated")
	}
}
