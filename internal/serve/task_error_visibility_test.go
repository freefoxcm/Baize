package serve

import (
	"strings"
	"testing"

	"reasonix/internal/provider"
	"reasonix/internal/tool"
)

func TestHistoryMessagesExposeOnlyStructuredToolFailures(t *testing.T) {
	tests := []struct {
		name string
		msg  provider.Message
		want bool
	}{
		{name: "durable marker", msg: provider.Message{Role: provider.RoleTool, ToolFailed: true}, want: true},
		{name: "legacy failed execution", msg: provider.Message{Role: provider.RoleTool, ToolExecution: &provider.ToolExecution{State: tool.ShellStateFailed}}, want: true},
		{name: "legacy not run execution", msg: provider.Message{Role: provider.RoleTool, ToolExecution: &provider.ToolExecution{State: tool.ShellStateNotRun}}, want: true},
		{name: "successful error-like text", msg: provider.Message{Role: provider.RoleTool, Content: "error: this is expected output", ToolExecution: &provider.ToolExecution{State: tool.ShellStateCompleted}}, want: false},
		{name: "unstructured old result", msg: provider.Message{Role: provider.RoleTool, Content: "error: old text only"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := historyMessages([]provider.Message{tt.msg})
			if len(got) != 1 || got[0].Failed != tt.want {
				t.Fatalf("history = %+v, want failed=%v", got, tt.want)
			}
		})
	}
}

func TestTaskErrorVisibilityWebContract(t *testing.T) {
	source := baizeWebSource()
	for _, want := range []string{
		`id="setting-show-task-errors"`,
		`默认隐藏失败工具/子代理卡；任务失败提示、操作与安全错误始终可见。`,
		`showTaskErrors = false`,
		`setToolCardVisibility(card,tool);`,
		`const visibleCalls=allCalls.filter(tc=>!hiddenTranscriptTool(tc.name));`,
		`loadInitialTaskErrorVisibility().finally(reloadHistory);`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("Serve web source missing task-error visibility contract %q", want)
		}
	}
}
