package serve

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"reasonix/internal/config"
	"reasonix/internal/control"
)

func TestServeUsageCalendarPublishesScaleAndTurnMetric(t *testing.T) {
	ctrl := control.New(control.Options{})
	t.Cleanup(ctrl.Close)
	srv := New(ctrl, NewBroadcaster(), config.ServeConfig{})
	srv.statsDir = t.TempDir
	srv.now = func() time.Time { return time.Date(2026, time.August, 13, 12, 0, 0, 0, time.Local) }
	httpServer := httptest.NewServer(srv.Handler())
	t.Cleanup(httpServer.Close)

	resp, err := http.Get(httpServer.URL + "/usage/calendar?range=3m")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var got struct {
		Max   int64 `json:"max"`
		Scale usageCalendarScale
		Turn  usageCalendarTurnMetric `json:"turnMetric"`
		Days  []map[string]any        `json:"days"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Max != 0 || got.Scale.Method != usageCalendarScaleMethod || got.Scale.MaxTokens != 0 || got.Scale.Levels != 5 {
		t.Fatalf("calendar scale = max %d, %+v", got.Max, got.Scale)
	}
	if got.Turn.Kind != "completed_execution" || got.Turn.Source != "serve" {
		t.Fatalf("turn metric = %+v", got.Turn)
	}
	if len(got.Days) == 0 {
		t.Fatal("calendar returned no days")
	}
	if level, ok := got.Days[0]["level"].(float64); !ok || level != 0 {
		t.Fatalf("inactive day level = %#v", got.Days[0]["level"])
	}
}
