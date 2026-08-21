package serve

import (
	"reasonix/internal/provider"
	"reasonix/internal/tool"
)

func populateToolHistoryMetadata(history *historyMessage, message provider.Message) {
	history.ToolCallID = message.ToolCallID
	history.ToolName = message.Name
	history.DurationMs = message.ToolDurationMs
	history.Failed = message.ToolFailed || legacyToolExecutionFailed(message.ToolExecution)
}

func legacyToolExecutionFailed(execution *provider.ToolExecution) bool {
	if execution == nil {
		return false
	}
	switch execution.State {
	case tool.ShellStateFailed, tool.ShellStateTimedOut, tool.ShellStateCancelled, tool.ShellStateNotRun:
		return true
	default:
		return false
	}
}
