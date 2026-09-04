package agent

import (
	"fmt"
	"strings"
)

// summarizeReadinessGaps keeps repeated host-observed obligations compact while
// preserving their first-seen order in completion diagnostics.
func summarizeReadinessGaps(gaps []string) string {
	counts := make(map[string]int, len(gaps))
	ordered := make([]string, 0, len(gaps))
	for _, gap := range gaps {
		gap = strings.TrimSpace(gap)
		if gap == "" {
			continue
		}
		if counts[gap] == 0 {
			ordered = append(ordered, gap)
		}
		counts[gap]++
	}
	for i, gap := range ordered {
		if counts[gap] > 1 {
			ordered[i] = fmt.Sprintf("%s (%d obligations)", gap, counts[gap])
		}
	}
	return strings.Join(ordered, "; ")
}
