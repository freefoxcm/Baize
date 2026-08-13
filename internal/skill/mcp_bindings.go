package skill

import (
	"path"
	"sort"
	"strings"

	"reasonix/internal/tool"
)

func hasPortableMCPToolRef(refs []string) bool {
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref == "*" || strings.HasPrefix(ref, "mcp__") || strings.HasPrefix(ref, "mcp-tool:") {
			return true
		}
	}
	return false
}

func bindAllowedTools(refs []string, bindings []tool.MCPBinding) []string {
	if len(refs) == 0 {
		return refs
	}
	out := make([]string, 0, len(refs))
	seen := map[string]bool{}
	appendOne := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	for _, ref := range refs {
		matches := map[string]tool.MCPBinding{}
		isPattern := strings.ContainsAny(ref, "*?[")
		for _, binding := range bindings {
			aliases := append(tool.MCPBindingAliases(binding), binding.CallableName)
			for _, alias := range aliases {
				matched := ref == alias
				if isPattern {
					matched, _ = path.Match(ref, alias)
				}
				if matched {
					matches[binding.CallableName] = binding
					break
				}
			}
		}
		if isPattern {
			appendOne(ref)
			names := make([]string, 0, len(matches))
			for name := range matches {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				if matched, err := path.Match(ref, name); err != nil || !matched {
					appendOne(name)
				}
				proxyMatched, _ := path.Match(ref, "use_capability")
				if !proxyMatched {
					appendOne(matches[name].CapabilityID)
				}
			}
			continue
		}
		if len(matches) == 1 {
			for name, binding := range matches {
				appendOne(name)
				appendOne(binding.CapabilityID)
			}
			continue
		}
		appendOne(ref)
	}
	return out
}
