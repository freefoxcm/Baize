package agent

import (
	"context"
	"encoding/json"
	"strings"

	"reasonix/internal/provider"
	"reasonix/internal/tool"
)

func (a *Agent) repairRawCompleteStepJSON(ctx context.Context, state *turnRuntime, text string) bool {
	if a == nil || state == nil || state.completeStepJSONRepairs > 0 || !a.hasActiveTodo() || !a.completeStepVisible(ctx) {
		return false
	}
	if !rawCompleteStepPayload(text) {
		return false
	}
	state.completeStepJSONRepairs++
	msg := "Your previous response serialized complete_step arguments as assistant text. Do not execute or repeat that JSON in prose. Reissue it now as exactly one structured complete_step tool call through the tool channel."
	a.sess.conversation.Add(provider.Message{Role: provider.RoleUser, Content: a.withTurnPreferences(msg)})
	return true
}

func (a *Agent) hasActiveTodo() bool {
	for _, item := range a.sess.todoState {
		if strings.TrimSpace(item.Status) == "in_progress" {
			return true
		}
	}
	return false
}

func (a *Agent) completeStepVisible(ctx context.Context) bool {
	if a.svc.tools == nil || !a.svc.tools.ProviderVisible("complete_step") {
		return false
	}
	target, _, ambiguous := a.svc.tools.ResolveCall("complete_step")
	if target == nil || len(ambiguous) > 0 {
		return false
	}
	contextual, ok := target.(tool.ContextualTool)
	return !ok || contextual.ProviderVisible(ctx)
}

func rawCompleteStepPayload(text string) bool {
	trimmed := strings.TrimSpace(text)
	if len(trimmed) < 2 || trimmed[0] != '{' || trimmed[len(trimmed)-1] != '}' {
		return false
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal([]byte(trimmed), &fields) != nil {
		return false
	}
	allowed := map[string]bool{
		"step_id": true, "step": true, "step_index": true,
		"result": true, "evidence": true, "notes": true,
	}
	for key := range fields {
		if !allowed[key] {
			return false
		}
	}
	if !completeStepIdentityPresent(fields) || !nonEmptyJSONString(fields["result"]) {
		return false
	}
	var evidenceItems []map[string]json.RawMessage
	if json.Unmarshal(fields["evidence"], &evidenceItems) != nil || len(evidenceItems) == 0 {
		return false
	}
	for _, item := range evidenceItems {
		if !nonEmptyJSONString(item["kind"]) || !nonEmptyJSONString(item["summary"]) {
			return false
		}
	}
	return true
}

func completeStepIdentityPresent(fields map[string]json.RawMessage) bool {
	if nonEmptyJSONString(fields["step_id"]) || nonEmptyJSONString(fields["step"]) {
		return true
	}
	var index int
	return json.Unmarshal(fields["step_index"], &index) == nil && index > 0
}

func nonEmptyJSONString(raw json.RawMessage) bool {
	var value string
	return json.Unmarshal(raw, &value) == nil && strings.TrimSpace(value) != ""
}
