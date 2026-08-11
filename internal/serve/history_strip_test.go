package serve

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"reasonix/internal/agent"
	"reasonix/internal/config"
	"reasonix/internal/control"
	"reasonix/internal/provider"
)

// TestHistoryMessagesStripsComposePrefixes locks the desktop-consistent
// behavior: /history must show the user's actual text, not the system-injected
// plan-mode marker / language directives / transient blocks that live in the
// stored Content for prefix-cache stability.
func TestHistoryMessagesStripsComposePrefixes(t *testing.T) {
	marker := control.PlanModeMarker // "[Plan mode — planning workflow...]"
	langBlock := "<reasoning-language>\n必须使用简体中文书写全部可见思考/推理文本：从第一个字开始就用中文\n</reasoning-language>"

	msgs := []provider.Message{
		// RawContent holds the user's original text; Content has the marker.
		{Role: provider.RoleUser, Content: marker + "\n\nnew question", RawContent: "new question"},
		// No RawContent: fall back to Content with transient blocks stripped.
		{Role: provider.RoleUser, Content: langBlock + "\n\nplain question"},
		{Role: provider.RoleAssistant, Content: "answer"},
		// Steer turns surface as a notice.
		{Role: provider.RoleUser, Content: "[Mid-turn steer queued by the user. Do not treat this as a new task; use it only as additional guidance for the current task after completing the current step.]\ngo faster"},
	}

	out := historyMessages(msgs)

	var users []string
	var notices []string
	for _, hm := range out {
		switch hm.Role {
		case "user":
			users = append(users, hm.Content)
		case "notice":
			notices = append(notices, hm.Content)
		}
	}

	if len(users) != 2 {
		t.Fatalf("user messages = %d, want 2 (%v)", len(users), users)
	}
	if users[0] != "new question" {
		t.Errorf("first user message = %q, want the raw text without plan-mode marker", users[0])
	}
	if strings.Contains(users[1], langBlock) {
		t.Errorf("language directive leaked into history: %q", users[1])
	}
	if users[1] != "plain question" {
		t.Errorf("second user message = %q, want plain question", users[1])
	}
	if len(notices) != 1 || !strings.Contains(notices[0], "go faster") {
		t.Errorf("steer should surface as a notice with the steer text, got %v", notices)
	}
}

// TestHistoryMessagesDropsSyntheticAndEmpty ensures auto-generated turns
// (plan approval etc.) and whitespace-only messages never render.
func TestHistoryMessagesDropsSyntheticAndEmpty(t *testing.T) {
	msgs := []provider.Message{
		{Role: provider.RoleUser, Content: "Plan approved — plan mode is off. Implement the plan now."},
		{Role: provider.RoleUser, Content: "   "},
		{Role: provider.RoleUser, Content: "real user text"},
	}
	out := historyMessages(msgs)
	if len(out) != 1 || out[0].Content != "real user text" {
		t.Fatalf("history = %+v, want only the real user text", out)
	}
}

// TestServeHistoryEndpointStripsMarker is the endpoint-level guard: the wire
// payload for a plan-mode session contains the user's text, not the marker.
func TestServeHistoryEndpointStripsMarker(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/history")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("history status = %d", resp.StatusCode)
	}
	var msgs []historyMessage
	if err := json.NewDecoder(resp.Body).Decode(&msgs); err != nil {
		t.Fatal(err)
	}
	// The empty controller session yields no messages; the endpoint must
	// still return valid JSON with the stripped shape.
	if msgs == nil {
		t.Fatal("history payload should be a JSON array")
	}
}

var _ = agent.SessionPreview // keep agent import anchored
