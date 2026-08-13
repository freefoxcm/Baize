package boot

import (
	"testing"

	"reasonix/internal/skill"
)

func TestSkillInheritsCallAskerOnlyWhenExplicitlyAllowed(t *testing.T) {
	tests := []struct {
		name    string
		allowed []string
		want    bool
	}{
		{name: "explicit", allowed: []string{"ask", "mcp-tool:ipap/aggregate_cases"}, want: true},
		{name: "trimmed", allowed: []string{" ask "}, want: true},
		{name: "absent", allowed: []string{"mcp-tool:ipap/aggregate_cases"}},
		{name: "wildcard", allowed: []string{"*"}},
		{name: "empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := skillInheritsCallAsker(skill.Skill{AllowedTools: tt.allowed}); got != tt.want {
				t.Fatalf("skillInheritsCallAsker() = %v, want %v", got, tt.want)
			}
		})
	}
}
