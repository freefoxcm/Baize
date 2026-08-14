package capdiag

import (
	"strings"

	"reasonix/internal/skill"
)

func skillAssetEntry(candidate skill.Candidate, displayPath func(string) string) AssetEntry {
	modes := make([]string, len(candidate.RunModes))
	for i, mode := range candidate.RunModes {
		modes[i] = string(mode)
	}
	return AssetEntry{
		Name: candidate.Name, Description: candidate.Description, Scope: string(candidate.Scope),
		Path: displayPath(candidate.Path), Status: string(candidate.Status), RunAs: string(candidate.RunAs), RunModes: modes,
	}
}

func invalidSkillRunModesIssue(candidate skill.Candidate, displayPath func(string) string) *Issue {
	if len(candidate.InvalidRunModes) == 0 {
		return nil
	}
	return &Issue{
		Severity: "error", Code: "skill.invalid_run_modes", Subsystem: "skills",
		Name: candidate.Name, Source: displayPath(candidate.Path),
		Message:     "skill has invalid allowed-run-modes: " + strings.Join(candidate.InvalidRunModes, ", "),
		Remediation: "Use only inline and subagent, and include the declared runAs default",
		SettingsTab: "skills",
	}
}
