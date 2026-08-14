package skill

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"reasonix/internal/event"
)

// ErrRunModeSelectionCancelled reports that an interactive invocation was
// dismissed before the skill started.
var ErrRunModeSelectionCancelled = errors.New("skill run mode selection cancelled")

// RunModeSelector obtains one host-authoritative choice for a selectable skill.
type RunModeSelector func(context.Context, Skill) (RunAs, error)

const (
	runModeQuestionID    = "skill-run-mode"
	runModeSubagentLabel = "子代理模式"
	runModeInlineLabel   = "当前会话模式"
	runModeRecommended   = "（推荐）"
)

// RunModeQuestion describes the host-level choice shown before a selectable
// skill receives data or tools.
func RunModeQuestion(sk Skill) event.AskQuestion {
	options := make([]event.AskOption, 0, len(sk.RunModes()))
	options = append(options, runModeOption(sk.RunAs, true))
	for _, mode := range sk.RunModes() {
		if mode != sk.RunAs {
			options = append(options, runModeOption(mode, false))
		}
	}
	return event.AskQuestion{
		ID:      runModeQuestionID,
		Header:  "运行方式",
		Prompt:  fmt.Sprintf("本次要使用哪种方式运行技能 %s？", sk.Name),
		Options: options,
	}
}

func runModeOption(mode RunAs, recommended bool) event.AskOption {
	option := event.AskOption{}
	switch mode {
	case RunInline:
		option.Label = runModeInlineLabel
		option.Description = "继承当前对话上下文，并使用当前会话的工具和审批策略运行。"
	case RunSubagent:
		option.Label = runModeSubagentLabel
		option.Description = "在隔离上下文中运行，只允许使用技能明确授权的工具，完成后将最终结果返回当前会话。"
	}
	if recommended {
		option.Label += runModeRecommended
	}
	return option
}

// RunModeFromAnswers validates one answer to RunModeQuestion.
func RunModeFromAnswers(answers []event.AskAnswer) (RunAs, error) {
	for _, answer := range answers {
		if answer.QuestionID != runModeQuestionID || len(answer.Selected) == 0 {
			continue
		}
		selected := strings.TrimSuffix(answer.Selected[0], runModeRecommended)
		switch selected {
		case runModeSubagentLabel:
			return RunSubagent, nil
		case runModeInlineLabel:
			return RunInline, nil
		default:
			return "", fmt.Errorf("unsupported skill run mode selection %q", answer.Selected[0])
		}
	}
	return "", ErrRunModeSelectionCancelled
}

// RunModes returns the modes an invocation may use. A legacy skill has exactly
// its configured RunAs mode.
func (s Skill) RunModes() []RunAs {
	if len(s.AllowedRunModes) == 0 {
		return []RunAs{s.RunAs}
	}
	return append([]RunAs(nil), s.AllowedRunModes...)
}

// SupportsRunMode reports whether the skill author enabled mode.
func (s Skill) SupportsRunMode(mode RunAs) bool {
	return slices.Contains(s.RunModes(), mode)
}

// HasSelectableRunMode reports whether the host must choose between modes.
func (s Skill) HasSelectableRunMode() bool { return len(s.RunModes()) > 1 }

// ResolveRunMode validates an explicit request or asks the host when the skill
// opted into selection. A missing selector uses the declared default.
func ResolveRunMode(ctx context.Context, sk Skill, requested string, selector RunModeSelector) (RunAs, error) {
	if len(sk.InvalidRunModes) > 0 {
		return "", fmt.Errorf("skill %q has invalid allowed-run-modes: %s", sk.Name, strings.Join(sk.InvalidRunModes, ", "))
	}
	if raw := strings.TrimSpace(requested); raw != "" {
		mode, ok := normalizeRunMode(raw)
		if !ok {
			return "", fmt.Errorf("invalid run_mode %q; expected inline or subagent", raw)
		}
		if !sk.SupportsRunMode(mode) {
			return "", fmt.Errorf("skill %q does not allow run_mode=%s; allowed: %s", sk.Name, mode, joinRunModes(sk.RunModes()))
		}
		return mode, nil
	}
	if !sk.HasSelectableRunMode() || selector == nil {
		return sk.RunAs, nil
	}
	mode, err := selector(ctx, sk)
	if err != nil {
		return "", err
	}
	if !sk.SupportsRunMode(mode) {
		return "", fmt.Errorf("run mode selector returned unsupported mode %q for skill %q", mode, sk.Name)
	}
	return mode, nil
}

func parseAllowedRunModes(raw string, defaultMode RunAs) (valid []RunAs, invalid []string) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	seen := map[RunAs]bool{}
	seenInvalid := map[string]bool{}
	for _, value := range parseCSVFrontmatter(raw) {
		mode, ok := normalizeRunMode(value)
		if !ok {
			value = strings.ToLower(strings.TrimSpace(value))
			if value != "" && !seenInvalid[value] {
				seenInvalid[value] = true
				invalid = append(invalid, value)
			}
			continue
		}
		if !seen[mode] {
			seen[mode] = true
			valid = append(valid, mode)
		}
	}
	if !seen[defaultMode] {
		invalid = append(invalid, fmt.Sprintf("default %s is not listed", defaultMode))
	}
	return valid, invalid
}

func normalizeRunMode(raw string) (RunAs, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(RunInline):
		return RunInline, true
	case string(RunSubagent):
		return RunSubagent, true
	default:
		return "", false
	}
}

func joinRunModes(modes []RunAs) string {
	values := make([]string, 0, len(modes))
	for _, mode := range modes {
		values = append(values, string(mode))
	}
	return strings.Join(values, ", ")
}
