package evidence

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"reasonix/internal/effectscope"
)

// MatchSuccessfulComputationAfter verifies a computation citation inside the
// current plan step's evidence window.
func (l *Ledger) MatchSuccessfulComputationAfter(name string, after int) (string, error) {
	name = strings.TrimSpace(name)
	if l == nil || name == "" {
		return "", fmt.Errorf("computation tool name is required")
	}
	start := max(after+1, 0)
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := start; i < len(l.receipts); i++ {
		r := l.receipts[i]
		if successfulComputationReceipt(r) && strings.EqualFold(r.ToolName, name) {
			return r.ToolName, nil
		}
	}
	return "", fmt.Errorf("tool %q has no matching successful computation with output after the current step began", name)
}

// LatestSuccessfulComputationIndex returns the newest output-producing,
// host-confined computation receipt.
func (l *Ledger) LatestSuccessfulComputationIndex() (int, bool) {
	if l == nil {
		return -1, false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, receipt := range slices.Backward(l.receipts) {
		if successfulComputationReceipt(receipt) {
			return i, true
		}
	}
	return -1, false
}

// HasSuccessfulComputationSignoffAfter reports whether complete_step cited a
// successful computation after the supplied receipt boundary.
func (l *Ledger) HasSuccessfulComputationSignoffAfter(after int) bool {
	if l == nil {
		return false
	}
	start := max(after+1, 0)
	candidateStart := max(after, 0)
	l.mu.Lock()
	receipts := append([]Receipt(nil), l.receipts...)
	l.mu.Unlock()
	for i := start; i < len(receipts); i++ {
		r := receipts[i]
		if !r.Success || r.ToolName != "complete_step" {
			continue
		}
		for _, toolName := range completeStepComputationTools(r.Args) {
			for j := candidateStart; j < i; j++ {
				candidate := receipts[j]
				if successfulComputationReceipt(candidate) && strings.EqualFold(candidate.ToolName, toolName) {
					return true
				}
			}
		}
	}
	return false
}

func successfulComputationReceipt(r Receipt) bool {
	return r.Success && !r.Mutation && r.EffectScope == effectscope.Scratch && r.OutputBytes > 0
}

func completeStepComputationTools(args json.RawMessage) []string {
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
		if item.Kind == "computation" && strings.TrimSpace(item.Tool) != "" {
			tools = append(tools, strings.TrimSpace(item.Tool))
		}
	}
	return tools
}
