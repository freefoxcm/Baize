package control

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"

	"reasonix/internal/event"
	"reasonix/internal/skill"
)

type interactionState struct {
	enabled atomic.Bool
}

func (c *Controller) selectSkillRunMode(ctx context.Context, sk skill.Skill) (skill.RunAs, error) {
	if !c.interaction.enabled.Load() {
		c.notice(fmt.Sprintf("当前环境无法交互选择运行方式；技能 %s 将使用默认的%s。", sk.Name, localizedSkillRunMode(sk.RunAs)))
		return sk.RunAs, nil
	}
	answers, err := c.Ask(ctx, []event.AskQuestion{skill.RunModeQuestion(sk)})
	if err != nil {
		return "", fmt.Errorf("select skill run mode: %w", err)
	}
	mode, err := skill.RunModeFromAnswers(answers)
	if err != nil {
		return "", err
	}
	c.notice(fmt.Sprintf("技能 %s 本次将以%s运行。", sk.Name, localizedSkillRunMode(mode)))
	return mode, nil
}

func localizedSkillRunMode(mode skill.RunAs) string {
	if mode == skill.RunInline {
		return "当前会话模式"
	}
	return "子代理模式"
}

func parseSkillRunModeFlag(task string) (string, string, error) {
	fields := strings.Fields(task)
	if len(fields) == 0 || !strings.HasPrefix(fields[0], "--run-mode") {
		return task, "", nil
	}
	const prefix = "--run-mode="
	if !strings.HasPrefix(fields[0], prefix) || len(fields[0]) == len(prefix) {
		return "", "", fmt.Errorf("use --run-mode=inline or --run-mode=subagent")
	}
	return strings.Join(fields[1:], " "), strings.TrimPrefix(fields[0], prefix), nil
}

func (c *Controller) submitSkillInvocation(input, display string) bool {
	sk, task, ok := c.resolveSkillInvocation(input)
	if !ok {
		return false
	}
	task, requestedMode, err := parseSkillRunModeFlag(task)
	if err != nil {
		c.notice(err.Error())
		return true
	}
	if strings.TrimSpace(task) == "" && (sk.RunAs == skill.RunSubagent || sk.HasSelectableRunMode()) {
		usage := "usage: /" + sk.Name + " <task>"
		if sk.HasSelectableRunMode() {
			usage = "usage: /" + sk.Name + " [--run-mode=inline|subagent] <task>"
		}
		c.notice(usage)
		return true
	}
	c.runResolvedSkillSlash(sk, task, input, display, requestedMode)
	return true
}

func (c *Controller) runResolvedSkillSlash(sk skill.Skill, task, raw, display, requestedMode string) {
	c.runGuarded(func(ctx context.Context) error {
		if err := c.skills.validate(sk); err != nil {
			return err
		}
		mode, err := skill.ResolveRunMode(ctx, sk, requestedMode, c.selectSkillRunMode)
		if err != nil {
			return err
		}
		sk.RunAs = mode
		if mode == skill.RunSubagent {
			runner := c.skillRunner
			if runner == nil {
				return fmt.Errorf("subagent skill runner is unavailable for /%s", sk.Name)
			}
			return newTurnOrchestrator(c).runSubagentSkillGoalLoop(ctx, c.skills.prepare(sk), task, raw, display, runner, c.PlanMode())
		}
		return c.runGoalLoopWithRawDisplay(ctx, c.skills.render(sk, task), raw, display)
	})
}
