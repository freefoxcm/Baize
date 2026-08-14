package control

import (
	"context"
	"strings"
	"testing"
	"time"

	"reasonix/internal/agent"
	"reasonix/internal/event"
	"reasonix/internal/skill"
	"reasonix/internal/tool"
)

func TestParseSkillRunModeFlag(t *testing.T) {
	task, mode, err := parseSkillRunModeFlag("--run-mode=inline build a report")
	if err != nil || task != "build a report" || mode != "inline" {
		t.Fatalf("task=%q mode=%q err=%v", task, mode, err)
	}
	if _, _, err := parseSkillRunModeFlag("--run-mode inline build"); err == nil {
		t.Fatal("ambiguous flag form was accepted")
	}
}

func TestSelectableSlashSkillAsksBeforeInlineExecution(t *testing.T) {
	sess := agent.NewSession("system")
	exec := agent.New(nil, tool.NewRegistry(), sess, agent.Options{}, event.Discard)
	events := make(chan event.Event, 32)
	mainRunner := &fakeTurnRunner{}
	subagentCalls := 0
	c := New(Options{
		Runner:   mainRunner,
		Executor: exec,
		Sink:     event.FuncSink(func(e event.Event) { events <- e }),
		Skills: []skill.Skill{{
			Name: "report", Body: "REPORT_RULE", RunAs: skill.RunSubagent,
			AllowedRunModes: []skill.RunAs{skill.RunInline, skill.RunSubagent}, Scope: skill.ScopeGlobal,
		}},
		SkillRunner: func(context.Context, skill.Skill, string, skill.SubagentRunOptions) (string, error) {
			subagentCalls++
			return "child", nil
		},
	})
	defer c.Close()
	c.EnableInteractiveApproval()
	c.SubmitDisplay("/report build", "/report build")

	var ask event.Ask
	deadline := time.After(10 * time.Second)
	for ask.ID == "" {
		select {
		case e := <-events:
			if e.Kind == event.AskRequest {
				ask = e.Ask
			}
		case <-deadline:
			t.Fatal("run-mode question was not emitted")
		}
	}
	c.AnswerQuestion(ask.ID, []event.AskAnswer{{QuestionID: "skill-run-mode", Selected: []string{"当前会话模式"}}})
	waitForTurnEvents(t, events)
	waitIdle(t, c)

	if subagentCalls != 0 || len(mainRunner.inputs) != 1 || !strings.Contains(mainRunner.inputs[0], "REPORT_RULE") || !strings.Contains(mainRunner.inputs[0], "Arguments: build") {
		t.Fatalf("subagent=%d main=%q", subagentCalls, mainRunner.inputs)
	}
}

func TestStructuredInvocationHonorsExplicitInlineMode(t *testing.T) {
	sess := agent.NewSession("system")
	exec := agent.New(nil, tool.NewRegistry(), sess, agent.Options{}, event.Discard)
	events := make(chan event.Event, 24)
	mainRunner := &fakeTurnRunner{}
	c := New(Options{
		Runner: mainRunner, Executor: exec, Sink: event.FuncSink(func(e event.Event) { events <- e }),
		Skills: []skill.Skill{{
			Name: "report", Body: "REPORT_RULE", RunAs: skill.RunSubagent,
			AllowedRunModes: []skill.RunAs{skill.RunInline, skill.RunSubagent}, Scope: skill.ScopeGlobal,
		}},
		SkillRunner: func(context.Context, skill.Skill, string, skill.SubagentRunOptions) (string, error) {
			t.Fatal("explicit inline invocation used subagent runner")
			return "", nil
		},
	})
	defer c.Close()

	c.SubmitInvocationDisplay("build", "build", []InvocationRequest{{Name: "report", Kind: "subagent", RunMode: "inline"}})
	waitForTurnEvents(t, events)
	waitIdle(t, c)
	if len(mainRunner.inputs) != 1 || !strings.Contains(mainRunner.inputs[0], "REPORT_RULE") {
		t.Fatalf("main inputs=%q", mainRunner.inputs)
	}
}
