package evidence

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// StepEvidenceBoundary returns the latest successful workflow receipt that
// starts a new evidence window for the current plan step.
func (l *Ledger) StepEvidenceBoundary() int {
	if l == nil {
		return -1
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	latest := -1
	for i, r := range l.receipts {
		if r.Success && (r.ToolName == "todo_write" || r.ToolName == "complete_step") {
			latest = i
		}
	}
	return latest
}

// MatchSuccessfulReadToolAfter resolves a tool citation against successful,
// output-producing read receipts in the current step's evidence window.
func (l *Ledger) MatchSuccessfulReadToolAfter(name string, after int) (string, error) {
	name = strings.TrimSpace(name)
	if l == nil || name == "" {
		return "", fmt.Errorf("tool name is required")
	}
	start := max(after+1, 0)
	l.mu.Lock()
	defer l.mu.Unlock()

	exact := map[string]struct{}{}
	aliases := map[string]struct{}{}
	for i := start; i < len(l.receipts); i++ {
		r := l.receipts[i]
		if !successfulObservationReceipt(r) {
			continue
		}
		if strings.EqualFold(r.ToolName, name) {
			exact[r.ToolName] = struct{}{}
			continue
		}
		if observationToolSuffix(r.ToolName, name) {
			aliases[r.ToolName] = struct{}{}
		}
	}
	if len(exact) == 1 {
		for candidate := range exact {
			return candidate, nil
		}
	}
	if len(aliases) == 1 {
		for candidate := range aliases {
			return candidate, nil
		}
	}
	candidates := make([]string, 0, len(exact)+len(aliases))
	for candidate := range exact {
		candidates = append(candidates, candidate)
	}
	for candidate := range aliases {
		candidates = append(candidates, candidate)
	}
	sort.Strings(candidates)
	if len(candidates) > 1 {
		return "", fmt.Errorf("tool reference %q is ambiguous; use one of: %s", name, strings.Join(candidates, ", "))
	}
	return "", fmt.Errorf("tool %q has no matching successful read-only receipt with output after the current step began", name)
}

// HasSuccessfulObservationSignoffAfter reports whether complete_step cited a
// host-observed read tool or a recognized verifier after the given boundary.
func (l *Ledger) HasSuccessfulObservationSignoffAfter(after int) bool {
	if l == nil {
		return false
	}
	start := max(after+1, 0)
	l.mu.Lock()
	receipts := append([]Receipt(nil), l.receipts...)
	l.mu.Unlock()
	for i := start; i < len(receipts); i++ {
		r := receipts[i]
		if !r.Success || r.ToolName != "complete_step" {
			continue
		}
		for _, toolName := range completeStepObservationTools(r.Args) {
			for j := start; j < i; j++ {
				candidate := receipts[j]
				if successfulObservationReceipt(candidate) &&
					(strings.EqualFold(candidate.ToolName, toolName) || observationToolSuffix(candidate.ToolName, toolName)) {
					return true
				}
			}
		}
		for _, command := range completeStepVerificationCommands(r.Args) {
			if !bashCommandIsVerification(command) {
				continue
			}
			for j := start; j < i; j++ {
				candidate := receipts[j]
				if candidate.Success && candidate.ToolName == "bash" && CommandMatches(command, candidate.Command) {
					return true
				}
			}
		}
	}
	return false
}

func successfulObservationReceipt(r Receipt) bool {
	return r.Success && r.Read && !r.Write && !r.Mutation && r.OutputBytes > 0 && !workflowBookkeepingTool(r.ToolName)
}

func observationToolSuffix(candidate, requested string) bool {
	candidate = strings.ToLower(strings.TrimSpace(candidate))
	requested = strings.ToLower(strings.TrimSpace(requested))
	return requested != "" && (strings.HasSuffix(candidate, "__"+requested) || strings.HasSuffix(candidate, "/"+requested))
}

func completeStepObservationTools(args json.RawMessage) []string {
	var payload struct {
		Evidence []struct {
			Kind string `json:"kind"`
			Tool string `json:"tool"`
		} `json:"evidence"`
	}
	if json.Unmarshal(args, &payload) != nil {
		return nil
	}
	var tools []string
	for _, item := range payload.Evidence {
		if item.Kind == "tool" && strings.TrimSpace(item.Tool) != "" {
			tools = append(tools, strings.TrimSpace(item.Tool))
		}
	}
	return tools
}
