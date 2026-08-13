package boot

import (
	"context"
	"testing"

	"reasonix/internal/plugin"
	"reasonix/internal/skill"
	"reasonix/internal/tool"
)

func TestSkillMCPBindingsProjectSkillUsesOnlyExplicitTools(t *testing.T) {
	specs := []plugin.Spec{{Name: "ipap"}, {Name: "other"}}
	cached := map[string][]plugin.CachedTool{
		"ipap":  {{Name: "aggregate_cases"}, {Name: "search_cases"}, {Name: "delete_case"}},
		"other": {{Name: "aggregate_cases"}},
	}
	sk := skill.Skill{Scope: skill.ScopeProject, AllowedTools: []string{"ask", "mcp__ipap__aggregate_cases", "mcp-tool:ipap/search_cases"}, Requires: []string{"mcp-server:ipap"}}
	got := skillMCPBindings(sk, nil, specs, cached, map[string]bool{"ipap": true, "other": true})
	if len(got) != 2 || got[0].Server != "ipap" || got[0].RawName != "aggregate_cases" || got[1].Server != "ipap" || got[1].RawName != "search_cases" {
		t.Fatalf("project skill bindings = %+v, want only explicit IPAP tools", got)
	}

	serverOnly := skillMCPBindings(skill.Skill{Scope: skill.ScopeProject, Requires: []string{"mcp-server:ipap"}}, nil, specs, cached, map[string]bool{"ipap": true, "other": true})
	if len(serverOnly) != 0 {
		t.Fatalf("server requirement widened project skill bindings: %+v", serverOnly)
	}

	malformed := skillMCPBindings(skill.Skill{Scope: skill.ScopeProject, AllowedTools: []string{"mcp-server:ipap", "mcp-server:ipap/aggregate_cases", "mcp-tool:ipap/*"}}, nil, specs, cached, map[string]bool{"ipap": true})
	if len(malformed) != 0 {
		t.Fatalf("server-level or wildcard references widened project skill bindings: %+v", malformed)
	}

	reg := tool.NewRegistry()
	host := plugin.NewHost()
	t.Cleanup(host.Close)
	for _, live := range plugin.LazyToolset(plugin.Spec{Name: "ipap"}, &plugin.CachedSchema{Tools: []plugin.CachedTool{{Name: "aggregate_cases"}, {Name: "delete_case"}}}, host, reg, context.Background(), false) {
		reg.Add(live)
	}
	live := skillMCPBindings(skill.Skill{Scope: skill.ScopeProject, AllowedTools: []string{"mcp__ipap__aggregate_cases"}}, reg, []plugin.Spec{{Name: "ipap"}}, cached, map[string]bool{"ipap": true})
	if len(live) != 1 || live[0].RawName != "aggregate_cases" || live[0].CapabilityID != "mcp-tool:ipap/aggregate_cases" {
		t.Fatalf("live project skill bindings = %+v, want aggregate_cases only", live)
	}
}
