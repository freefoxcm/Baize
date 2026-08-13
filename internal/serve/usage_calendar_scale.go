package serve

import (
	"math"
	"slices"
)

const (
	usageCalendarScaleMethod = "active_p95_linear"
	usageCalendarScaleLevels = 5
)

type usageCalendarScale struct {
	Method    string `json:"method"`
	MaxTokens int64  `json:"maxTokens"`
	Levels    int    `json:"levels"`
}

type usageCalendarTurnMetric struct {
	Kind   string `json:"kind"`
	Source string `json:"source"`
}

func usageCalendarLevels(tokens []int64) (usageCalendarScale, []int) {
	active := make([]int64, 0, len(tokens))
	for _, value := range tokens {
		if value > 0 {
			active = append(active, value)
		}
	}
	slices.Sort(active)
	var ceiling int64
	if len(active) > 0 {
		rank := int(math.Ceil(float64(len(active)) * 0.95))
		ceiling = active[rank-1]
	}

	levels := make([]int, len(tokens))
	for i, value := range tokens {
		if value <= 0 || ceiling <= 0 {
			continue
		}
		if value >= ceiling {
			levels[i] = usageCalendarScaleLevels
			continue
		}
		levels[i] = max(1, int(math.Ceil(float64(value)/float64(ceiling)*usageCalendarScaleLevels)))
	}
	return usageCalendarScale{Method: usageCalendarScaleMethod, MaxTokens: ceiling, Levels: usageCalendarScaleLevels}, levels
}
