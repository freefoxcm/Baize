package boot

import (
	"context"
	"fmt"

	"reasonix/internal/agent"
	"reasonix/internal/event"
	"reasonix/internal/skill"
)

func selectSkillRunMode(ctx context.Context, sk skill.Skill) (skill.RunAs, error) {
	_, sink, asker, ok := agent.CallContext(ctx)
	if !ok || asker == nil {
		emitSkillRunModeNotice(sink, sk.Name, sk.RunAs, true)
		return sk.RunAs, nil
	}
	answers, err := asker.Ask(ctx, []event.AskQuestion{skill.RunModeQuestion(sk)})
	if err != nil {
		return "", fmt.Errorf("select skill run mode: %w", err)
	}
	mode, err := skill.RunModeFromAnswers(answers)
	if err != nil {
		return "", err
	}
	emitSkillRunModeNotice(sink, sk.Name, mode, false)
	return mode, nil
}

func emitSkillRunModeNotice(sink event.Sink, name string, mode skill.RunAs, usedDefault bool) {
	if sink == nil {
		return
	}
	text := fmt.Sprintf("技能 %s 本次将以 %s 模式运行。", name, localizedRunMode(mode))
	if usedDefault {
		text = fmt.Sprintf("当前环境无法交互选择运行方式；技能 %s 将使用默认的%s。", name, localizedRunMode(mode))
	}
	sink.Emit(event.Event{Kind: event.Notice, Level: event.LevelInfo, Text: text})
}

func localizedRunMode(mode skill.RunAs) string {
	if mode == skill.RunInline {
		return "当前会话模式"
	}
	return "子代理模式"
}
