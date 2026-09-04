package evidence

import (
	"encoding/json"
	"os"
	"testing"

	"reasonix/internal/effectscope"
)

func TestEvidenceConsumesSharedCommandEffectMatrix(t *testing.T) {
	raw, err := os.ReadFile("../shellsafe/testdata/command_effects.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name, Command                      string
		Certainty                          string
		TaskPolicyBlocked, ContentMutation bool
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			args, _ := json.Marshal(map[string]string{"command": tc.Command})
			got := ClassifyToolCall("bash", args, false)
			if got.Known != (tc.Certainty == "known") || got.StateMutation != tc.TaskPolicyBlocked || got.ContentMutation != tc.ContentMutation {
				t.Fatalf("ClassifyToolCall(%q) = %+v, matrix=%+v", tc.Command, got, tc)
			}
		})
	}
}

func TestClassifyToolCallSeparatesMutationDomains(t *testing.T) {
	tests := []struct {
		name, tool, args string
		readOnly         bool
		want             ToolEffects
	}{
		{"branch listing", "bash", `{"command":"git branch -a"}`, false, ToolEffects{Known: true, Scope: effectscope.Observation}},
		{"tag creation", "bash", `{"command":"git tag v1.2.3"}`, true, ToolEffects{Known: true, StateMutation: true, WorkspaceMutation: true, RepositoryMutation: true, Scope: effectscope.Durable}},
		{"pure commit", "bash", `{"command":"git commit -q -m checkpoint"}`, false, ToolEffects{Known: true, StateMutation: true, WorkspaceMutation: true, RepositoryMutation: true, Scope: effectscope.Durable}},
		{"commit all", "bash", `{"command":"git commit -am checkpoint"}`, false, ToolEffects{Known: true, StateMutation: true, WorkspaceMutation: true, ContentMutation: true, RepositoryMutation: true, Scope: effectscope.Durable}},
		{"host clock", "bash", `{"command":"date --set tomorrow"}`, false, ToolEffects{Known: true, StateMutation: true, Scope: effectscope.Durable}},
		{"audit fix", "bash", `{"command":"npm audit fix"}`, false, ToolEffects{Known: true, StateMutation: true, WorkspaceMutation: true, ContentMutation: true, Scope: effectscope.Durable}},
		{"verification", "bash", `{"command":"go test ./..."}`, false, ToolEffects{Known: true, Scope: effectscope.Observation}},
		{"env prefixed test", "bash", `{"command":"GOROOT=/x go test ./..."}`, false, ToolEffects{Known: true, Scope: effectscope.Observation}},
		{"unknown shell", "bash", `{"command":"custom-tool --run"}`, false, ToolEffects{StateMutation: true, WorkspaceMutation: true, ContentMutation: true, Scope: effectscope.Unknown}},
		{"trusted reader", "read_file", `{}`, true, ToolEffects{Known: true, Scope: effectscope.Observation}},
		{"analysis", "analyze_data", `{}`, true, ToolEffects{Known: true, Scope: effectscope.Scratch}},
		{"fleet meta tool", "fleet", `{}`, false, ToolEffects{Known: true, Scope: effectscope.Observation}},
		{"session title", "set_session_title", `{"title":"Current task"}`, false, ToolEffects{Known: true, StateMutation: true, Scope: effectscope.Durable, Reason: "host session metadata write"}},
		{"background kill", "kill_shell", `{"job_id":"task-1"}`, false, ToolEffects{Known: true, StateMutation: true, Scope: effectscope.Durable}},
		{"generic writer", "edit_file", `{}`, false, ToolEffects{Known: true, StateMutation: true, WorkspaceMutation: true, ContentMutation: true, Scope: effectscope.Durable}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyToolCall(tt.tool, json.RawMessage(tt.args), tt.readOnly)
			got.Reason, tt.want.Reason = "", ""
			if got != tt.want {
				t.Fatalf("ClassifyToolCall(%q, %s) = %+v, want %+v", tt.tool, tt.args, got, tt.want)
			}
		})
	}
}

func TestReceiptMutationTracksContentNotRepositoryCement(t *testing.T) {
	pure := ReceiptFromToolCall("bash", json.RawMessage(`{"command":"git commit -m checkpoint"}`), true, false)
	all := ReceiptFromToolCall("bash", json.RawMessage(`{"command":"git commit -am checkpoint"}`), true, false)
	if pure.Mutation || !all.Mutation {
		t.Fatalf("commit receipt domains: pure=%+v all=%+v", pure, all)
	}
}
