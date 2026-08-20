package serve

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"reasonix/internal/agent"
	"reasonix/internal/config"
	"reasonix/internal/control"
)

func TestTimelineSessionsPaginatesSearchesAndSkipsCleanup(t *testing.T) {
	dir := t.TempDir()
	base := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
	paths := make([]string, 31)
	for index := range paths {
		name := fmt.Sprintf("session-%02d.jsonl", index)
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(fmt.Sprintf("{\"role\":\"user\",\"content\":\"message %02d\"}\n", index)), 0o644); err != nil {
			t.Fatal(err)
		}
		stamp := base.Add(time.Duration(index) * time.Minute)
		if err := os.Chtimes(path, stamp, stamp); err != nil {
			t.Fatal(err)
		}
		paths[index] = path
	}
	pending := filepath.Join(dir, "pending.jsonl")
	if err := os.WriteFile(pending, []byte("{\"role\":\"user\",\"content\":\"pending\"}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := agent.MarkCleanupPending(pending, "delete"); err != nil {
		t.Fatal(err)
	}

	controller := control.New(control.Options{SessionDir: dir, SessionPath: paths[30]})
	server := New(controller, NewBroadcaster(), config.ServeConfig{})
	first := timelinePageRequest(t, server, "/sessions/timeline?limit=30")
	if first.Total != 31 || len(first.Items) != 30 || first.NextCursor == "" {
		t.Fatalf("first page = %+v", first)
	}
	if first.Items[0].Name != "session-30" || !first.Items[0].UpdatedAt.Equal(base.Add(30*time.Minute)) || !first.Items[0].Current || first.Items[0].Turns != 1 {
		t.Fatalf("first item = %+v", first.Items[0])
	}
	second := timelinePageRequest(t, server, "/sessions/timeline?limit=30&cursor="+url.QueryEscape(first.NextCursor))
	if second.Total != 31 || len(second.Items) != 1 || second.Items[0].Name != "session-00" || second.NextCursor != "" {
		t.Fatalf("second page = %+v", second)
	}
	seen := map[string]bool{}
	for _, item := range append(first.Items, second.Items...) {
		if seen[item.Name] {
			t.Fatalf("duplicate item %q", item.Name)
		}
		seen[item.Name] = true
	}
	search := timelinePageRequest(t, server, "/sessions/timeline?q=session-05")
	if search.Total != 1 || len(search.Items) != 1 || search.Items[0].Name != "session-05" {
		t.Fatalf("search page = %+v", search)
	}
}

func TestTimelineSessionsRejectsInvalidCursorAndLimit(t *testing.T) {
	server := New(control.New(control.Options{SessionDir: t.TempDir()}), NewBroadcaster(), config.ServeConfig{})
	for _, target := range []string{"/sessions/timeline?cursor=broken", "/sessions/timeline?limit=nope"} {
		recorder := httptest.NewRecorder()
		server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d, want 400", target, recorder.Code)
		}
	}
}

func TestTimelineSessionsPinsCurrentDraft(t *testing.T) {
	dir := t.TempDir()
	draftPath := filepath.Join(dir, "session-draft.jsonl")
	server := New(control.New(control.Options{SessionDir: dir, SessionPath: draftPath}), NewBroadcaster(), config.ServeConfig{})
	page := timelinePageRequest(t, server, "/sessions/timeline")
	if page.Total != 1 || len(page.Items) != 1 || !page.Items[0].Draft || !page.Items[0].Current || page.Items[0].UpdatedAt.IsZero() {
		t.Fatalf("draft page = %+v", page)
	}
	if _, err := os.Stat(draftPath); !os.IsNotExist(err) {
		t.Fatalf("draft listing created transcript: %v", err)
	}
}

func TestTimelineSessionsExcludesDraftFromLaterPages(t *testing.T) {
	dir := t.TempDir()
	base := time.Date(2026, time.August, 1, 12, 0, 0, 0, time.UTC)
	for index := 0; index < 11; index++ {
		path := filepath.Join(dir, fmt.Sprintf("saved-%02d.jsonl", index))
		if err := os.WriteFile(path, []byte("{\"role\":\"user\",\"content\":\"saved\"}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		stamp := base.Add(time.Duration(index) * time.Minute)
		if err := os.Chtimes(path, stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}
	draftPath := filepath.Join(dir, "current-draft.jsonl")
	server := New(control.New(control.Options{SessionDir: dir, SessionPath: draftPath}), NewBroadcaster(), config.ServeConfig{})
	first := timelinePageRequest(t, server, "/sessions/timeline?limit=10")
	if first.Total != 12 || len(first.Items) != 10 || !first.Items[0].Draft || first.NextCursor == "" {
		t.Fatalf("first page = %+v", first)
	}
	second := timelinePageRequest(t, server, "/sessions/timeline?limit=10&cursor="+url.QueryEscape(first.NextCursor))
	if second.Total != 12 || len(second.Items) != 2 {
		t.Fatalf("second page = %+v", second)
	}
	for _, item := range second.Items {
		if item.Draft {
			t.Fatalf("draft leaked into cursor page: %+v", item)
		}
	}
}

func timelinePageRequest(t *testing.T, server *Server, target string) timelineSessionPage {
	t.Helper()
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("%s status = %d: %s", target, recorder.Code, recorder.Body.String())
	}
	var page timelineSessionPage
	if err := json.Unmarshal(recorder.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	return page
}
