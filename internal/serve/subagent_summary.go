package serve

import (
	"encoding/json"
	"log/slog"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"reasonix/internal/event"
	"reasonix/internal/fileutil"
	"reasonix/internal/store"
)

type subagentExecutionSummary struct {
	Kind          string   `json:"kind"`
	Name          string   `json:"name"`
	Model         string   `json:"model,omitempty"`
	Effort        string   `json:"effort,omitempty"`
	Status        string   `json:"status"`
	StartedAt     int64    `json:"startedAt,omitempty"`
	EndedAt       int64    `json:"endedAt,omitempty"`
	DurationMs    int64    `json:"durationMs,omitempty"`
	ToolNames     []string `json:"toolNames,omitempty"`
	ToolCallCount int      `json:"toolCallCount,omitempty"`
}

type subagentSummaryFile struct {
	Version int                                  `json:"version"`
	Calls   map[string]*subagentExecutionSummary `json:"calls"`
}

type subagentSummaryRecorder struct {
	mu      sync.Mutex
	files   map[string]*subagentSummaryFile
	toolIDs map[string]map[string]map[string]struct{}
}

func newSubagentSummaryRecorder() *subagentSummaryRecorder {
	return &subagentSummaryRecorder{files: make(map[string]*subagentSummaryFile), toolIDs: make(map[string]map[string]map[string]struct{})}
}

func (r *subagentSummaryRecorder) observe(sessionPath string, e event.Event) {
	sessionPath = strings.TrimSpace(sessionPath)
	if r == nil || sessionPath == "" || e.Tool.ID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	f := r.fileLocked(sessionPath)
	changed, persist := false, false
	switch e.Kind {
	case event.ToolDispatch:
		if e.Tool.ParentID != "" {
			if parent := f.Calls[e.Tool.ParentID]; parent != nil && e.Tool.Name != "" && r.markToolIDLocked(sessionPath, e.Tool.ParentID, e.Tool.ID) {
				parent.ToolCallCount++
				parent.ToolNames = appendUnique(parent.ToolNames, summaryToolName(e.Tool))
				changed = true
			}
		}
		if e.Tool.Profile != nil || isSubagentToolName(e.Tool.Name) {
			summary := f.Calls[e.Tool.ID]
			if summary == nil {
				summary = &subagentExecutionSummary{Kind: subagentKind(e.Tool.Name), Name: subagentDisplayName(e.Tool), Status: "running", StartedAt: time.Now().UnixMilli()}
				f.Calls[e.Tool.ID] = summary
			}
			if e.Tool.Profile != nil {
				summary.Model = e.Tool.Profile.Model
				summary.Effort = e.Tool.Profile.Effort
			}
			changed = true
		}
	case event.ToolProgress:
		if e.Tool.Name != event.SubagentProgressStatusName {
			break
		}
		summary := f.Calls[e.Tool.ID]
		if summary == nil || !terminalSubagentStatus(e.Tool.Output) {
			break
		}
		finishSubagentSummary(summary, e.Tool.Output, e.Tool.DurationMs, time.Now().UnixMilli())
		changed, persist = true, true
	case event.ToolResult:
		summary := f.Calls[e.Tool.ID]
		if summary == nil {
			break
		}
		status := "completed"
		if e.Tool.Err != "" {
			status = "failed"
		}
		if summary.Status == "running" {
			finishSubagentSummary(summary, status, e.Tool.DurationMs, e.Tool.EndedAt)
		}
		changed, persist = true, true
	}
	if changed && persist {
		r.persistLocked(sessionPath, f)
	}
}

func (r *subagentSummaryRecorder) markToolIDLocked(sessionPath, parentID, toolID string) bool {
	if toolID == "" {
		return false
	}
	parents := r.toolIDs[sessionPath]
	if parents == nil {
		parents = make(map[string]map[string]struct{})
		r.toolIDs[sessionPath] = parents
	}
	ids := parents[parentID]
	if ids == nil {
		ids = make(map[string]struct{})
		parents[parentID] = ids
	}
	if _, exists := ids[toolID]; exists {
		return false
	}
	ids[toolID] = struct{}{}
	return true
}

func summaryToolName(tool event.Tool) string {
	name := strings.TrimSpace(tool.ResolvedName)
	if name == "" {
		name = strings.TrimSpace(tool.Name)
	}
	if strings.HasPrefix(name, "mcp__") {
		if parts := strings.Split(name, "__"); len(parts) >= 3 {
			return parts[len(parts)-1]
		}
	}
	return name
}

func (r *subagentSummaryRecorder) summaries(sessionPath string) map[string]*subagentExecutionSummary {
	if r == nil || strings.TrimSpace(sessionPath) == "" {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	f := r.fileLocked(sessionPath)
	out := make(map[string]*subagentExecutionSummary, len(f.Calls))
	for id, summary := range f.Calls {
		copy := *summary
		copy.ToolNames = append([]string(nil), summary.ToolNames...)
		out[id] = &copy
	}
	return out
}

func (r *subagentSummaryRecorder) fileLocked(sessionPath string) *subagentSummaryFile {
	if f := r.files[sessionPath]; f != nil {
		return f
	}
	f := &subagentSummaryFile{Version: 1, Calls: make(map[string]*subagentExecutionSummary)}
	if raw, err := os.ReadFile(store.SessionSubagentSummary(sessionPath)); err == nil {
		if json.Unmarshal(raw, f) != nil || f.Calls == nil {
			f = &subagentSummaryFile{Version: 1, Calls: make(map[string]*subagentExecutionSummary)}
		}
	}
	r.files[sessionPath] = f
	return f
}

func (r *subagentSummaryRecorder) persistLocked(sessionPath string, f *subagentSummaryFile) {
	raw, err := json.Marshal(f)
	if err == nil {
		err = fileutil.AtomicWriteFile(store.SessionSubagentSummary(sessionPath), raw, 0o600)
	}
	if err != nil {
		slog.Warn("serve: persist subagent summary", "err", err)
	}
}

func finishSubagentSummary(summary *subagentExecutionSummary, status string, duration, endedAt int64) {
	if endedAt <= 0 {
		endedAt = time.Now().UnixMilli()
	}
	summary.Status = status
	summary.EndedAt = endedAt
	if duration <= 0 && summary.StartedAt > 0 {
		duration = endedAt - summary.StartedAt
	}
	summary.DurationMs = duration
}

func isSubagentToolName(name string) bool {
	switch name {
	case "task", "read_only_task", "parallel_tasks", "fleet":
		return true
	}
	return false
}

func subagentKind(name string) string {
	switch name {
	case "parallel_tasks":
		return "parallel"
	case "fleet":
		return "fleet"
	case "run_skill":
		return "skill"
	default:
		return "task"
	}
}

func subagentDisplayName(tool event.Tool) string {
	var args map[string]any
	if json.Unmarshal([]byte(tool.Args), &args) == nil {
		for _, key := range []string{"skill", "name", "task_name"} {
			if value, ok := args[key].(string); ok {
				value = strings.TrimSpace(value)
				if value != "" && len(value) <= 120 && !strings.ContainsAny(value, "\r\n\t") {
					return value
				}
			}
		}
	}
	return tool.Name
}

func terminalSubagentStatus(status string) bool {
	return status == "completed" || status == "failed" || status == "cancelled"
}

func appendUnique(values []string, value string) []string {
	if slices.Contains(values, value) {
		return values
	}
	return append(values, value)
}
