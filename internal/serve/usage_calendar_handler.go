package serve

import (
	"net/http"

	"reasonix/internal/stats"
)

func (s *Server) usageCalendar(w http.ResponseWriter, r *http.Request) {
	key, from, to, err := usageCalendarRange(s.now(), r.URL.Query().Get("range"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rs, err := stats.NewWriter(s.statsDir()).Query(stats.SourceFilter{Source: "serve", From: from, To: to})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type dayEntry struct {
		Day      string           `json:"day"`
		Tokens   int64            `json:"tokens"`
		Requests int              `json:"requests"`
		Turns    int              `json:"turns"`
		Level    int              `json:"level"`
		ByModel  map[string]int64 `json:"byModel,omitempty"`
	}
	days := make([]dayEntry, 0, len(rs.Daily))
	tokens := make([]int64, 0, len(rs.Daily))
	var trueMax int64
	for _, day := range rs.Daily {
		value := int64(day.Total)
		trueMax = max(trueMax, value)
		tokens = append(tokens, value)
		days = append(days, dayEntry{Day: day.Day, Tokens: value, Requests: day.Requests, Turns: day.Turns, ByModel: day.ByModel})
	}
	scale, levels := usageCalendarLevels(tokens)
	for i := range days {
		days[i].Level = levels[i]
	}
	writeJSON(w, map[string]any{
		"range": key, "from": from.Format(usageCalendarDateLayout), "to": to.Format(usageCalendarDateLayout),
		"days": days, "max": trueMax, "scale": scale, "total": rs.Tokens, "turns": rs.Turns,
		"turnMetric": usageCalendarTurnMetric{Kind: "completed_execution", Source: "serve"}, "activeDays": rs.ActiveDays,
	})
}
