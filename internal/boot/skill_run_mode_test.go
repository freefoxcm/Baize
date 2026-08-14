package boot

import (
	"context"
	"testing"

	"reasonix/internal/agent"
	"reasonix/internal/event"
	"reasonix/internal/skill"
)

type runModeAsker struct {
	answers []event.AskAnswer
}

func (a runModeAsker) Ask(context.Context, []event.AskQuestion) ([]event.AskAnswer, error) {
	return a.answers, nil
}

func TestSelectSkillRunModeUsesCallAsker(t *testing.T) {
	asker := runModeAsker{answers: []event.AskAnswer{{QuestionID: "skill-run-mode", Selected: []string{"当前会话模式（推荐）"}}}}
	ctx := agent.WithToolCallContext(context.Background(), "call", event.Discard, asker, false)
	got, err := selectSkillRunMode(ctx, skill.Skill{
		Name: "report", RunAs: skill.RunInline,
		AllowedRunModes: []skill.RunAs{skill.RunInline, skill.RunSubagent},
	})
	if err != nil || got != skill.RunInline {
		t.Fatalf("mode=%q err=%v", got, err)
	}
}

func TestSelectSkillRunModeHeadlessUsesDefault(t *testing.T) {
	got, err := selectSkillRunMode(context.Background(), skill.Skill{Name: "report", RunAs: skill.RunSubagent})
	if err != nil || got != skill.RunSubagent {
		t.Fatalf("mode=%q err=%v", got, err)
	}
}
