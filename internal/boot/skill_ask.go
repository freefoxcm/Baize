package boot

import (
	"strings"

	"reasonix/internal/skill"
)

func skillInheritsCallAsker(sk skill.Skill) bool {
	for _, name := range sk.AllowedTools {
		if strings.TrimSpace(name) == "ask" {
			return true
		}
	}
	return false
}
