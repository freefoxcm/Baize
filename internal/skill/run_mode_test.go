package skill

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"reasonix/internal/event"
)

func TestParseAllowedRunModes(t *testing.T) {
	valid, invalid := parseAllowedRunModes("inline, subagent, INLINE", RunSubagent)
	if !reflect.DeepEqual(valid, []RunAs{RunInline, RunSubagent}) || len(invalid) != 0 {
		t.Fatalf("valid=%v invalid=%v", valid, invalid)
	}
	valid, invalid = parseAllowedRunModes("inline, remote", RunSubagent)
	if !reflect.DeepEqual(valid, []RunAs{RunInline}) || !reflect.DeepEqual(invalid, []string{"remote", "default subagent is not listed"}) {
		t.Fatalf("valid=%v invalid=%v", valid, invalid)
	}
}

func TestResolveRunMode(t *testing.T) {
	sk := Skill{Name: "report", RunAs: RunSubagent, AllowedRunModes: []RunAs{RunInline, RunSubagent}}
	calls := 0
	selector := func(context.Context, Skill) (RunAs, error) {
		calls++
		return RunInline, nil
	}
	if got, err := ResolveRunMode(context.Background(), sk, "subagent", selector); err != nil || got != RunSubagent || calls != 0 {
		t.Fatalf("explicit mode=%q calls=%d err=%v", got, calls, err)
	}
	if got, err := ResolveRunMode(context.Background(), sk, "", selector); err != nil || got != RunInline || calls != 1 {
		t.Fatalf("selected mode=%q calls=%d err=%v", got, calls, err)
	}
	if _, err := ResolveRunMode(context.Background(), sk, "remote", selector); err == nil {
		t.Fatal("invalid mode was accepted")
	}
	fixed := Skill{Name: "fixed", RunAs: RunInline}
	if _, err := ResolveRunMode(context.Background(), fixed, "subagent", selector); err == nil {
		t.Fatal("unsupported override was accepted")
	}
}

func TestRunModeAnswers(t *testing.T) {
	question := RunModeQuestion(Skill{
		Name: "ipap-analysis-report", RunAs: RunInline,
		AllowedRunModes: []RunAs{RunInline, RunSubagent},
	})
	if question.Header != "运行方式" || question.Prompt != "本次要使用哪种方式运行技能 ipap-analysis-report？" {
		t.Fatalf("question not localized: %+v", question)
	}
	if len(question.Options) != 2 || question.Options[0].Label != "当前会话模式（推荐）" || question.Options[1].Label != "子代理模式" {
		t.Fatalf("options not localized: %+v", question.Options)
	}
	subagentDefault := RunModeQuestion(Skill{
		Name: "report", RunAs: RunSubagent,
		AllowedRunModes: []RunAs{RunInline, RunSubagent},
	})
	if len(subagentDefault.Options) != 2 || subagentDefault.Options[0].Label != "子代理模式（推荐）" || subagentDefault.Options[1].Label != "当前会话模式" {
		t.Fatalf("subagent default options = %+v", subagentDefault.Options)
	}
	mode, err := RunModeFromAnswers([]event.AskAnswer{{QuestionID: "skill-run-mode", Selected: []string{"当前会话模式（推荐）"}}})
	if err != nil || mode != RunInline {
		t.Fatalf("mode=%q err=%v", mode, err)
	}
	if _, err := RunModeFromAnswers(nil); !errors.Is(err, ErrRunModeSelectionCancelled) {
		t.Fatalf("dismiss err=%v", err)
	}
}

func TestIndexMarksSelectableRunMode(t *testing.T) {
	out := IndexBlock([]Skill{{
		Name: "report", Description: "build report", RunAs: RunSubagent,
		AllowedRunModes: []RunAs{RunInline, RunSubagent},
	}})
	if !strings.Contains(out, "report [↔ inline|subagent; default subagent]") {
		t.Fatalf("selectable tag missing:\n%s", out)
	}
}
