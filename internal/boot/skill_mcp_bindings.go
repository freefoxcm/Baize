package boot

import (
	"strings"

	"reasonix/internal/skill"
	"reasonix/internal/tool"
)

func skillMCPBindingRequested(sk skill.Skill, binding tool.MCPBinding) bool {
	if sk.Plugin != "" {
		return binding.Package == sk.Plugin
	}
	aliases := append(tool.MCPBindingAliases(binding), binding.CallableName)
	for _, ref := range sk.AllowedTools {
		ref = strings.TrimSpace(ref)
		if ref == "*" {
			return true
		}
		if !strings.HasPrefix(ref, "mcp__") && !strings.HasPrefix(ref, "mcp-tool:") {
			continue
		}
		for _, alias := range aliases {
			if ref == alias {
				return true
			}
		}
	}
	return false
}
