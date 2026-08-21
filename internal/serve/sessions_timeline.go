package serve

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"reasonix/internal/agent"
	"reasonix/internal/store"
)

const (
	timelineSessionsDefaultLimit = 30
	timelineSessionsMinLimit     = 10
	timelineSessionsMaxLimit     = 100
)

type timelineSessionEntry struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Title     string    `json:"title,omitempty"`
	Turns     int       `json:"turns,omitempty"`
	Current   bool      `json:"current,omitempty"`
	Draft     bool      `json:"draft,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type timelineSessionPage struct {
	Items      []timelineSessionEntry `json:"items"`
	NextCursor string                 `json:"nextCursor,omitempty"`
	Total      int                    `json:"total"`
}

type timelineSessionCursor struct {
	Time int64  `json:"t"`
	Name string `json:"n"`
}

func (s *Server) timelineSessions(w http.ResponseWriter, r *http.Request) {
	limit, err := timelineSessionLimit(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, "invalid limit", http.StatusBadRequest)
		return
	}
	cursor, hasCursor, err := decodeTimelineSessionCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		http.Error(w, "invalid cursor", http.StatusBadRequest)
		return
	}
	saved, draft := s.timelineSessionEntries(r)
	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	saved = filterTimelineSessions(saved, query)
	if draft != nil && !timelineSessionMatches(*draft, query) {
		draft = nil
	}
	total := len(saved)
	if draft != nil {
		total++
	}

	if hasCursor {
		saved = sessionsOlderThanCursor(saved, cursor)
		draft = nil
	}
	items := make([]timelineSessionEntry, 0, limit)
	if draft != nil {
		items = append(items, *draft)
	}
	remaining := limit - len(items)
	remaining = min(remaining, len(saved))
	items = append(items, saved[:remaining]...)
	page := timelineSessionPage{Items: items, Total: total}
	if remaining < len(saved) && remaining > 0 {
		page.NextCursor = encodeTimelineSessionCursor(saved[remaining-1])
	}
	writeJSON(w, page)
}

func timelineSessionLimit(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return timelineSessionsDefaultLimit, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return 0, errInvalidTimelineLimit
	}
	if limit < timelineSessionsMinLimit {
		limit = timelineSessionsMinLimit
	}
	if limit > timelineSessionsMaxLimit {
		limit = timelineSessionsMaxLimit
	}
	return limit, nil
}

var errInvalidTimelineLimit = &timelineSessionQueryError{}

type timelineSessionQueryError struct{}

func (*timelineSessionQueryError) Error() string { return "invalid timeline session query" }

func (s *Server) timelineSessionEntries(r *http.Request) ([]timelineSessionEntry, *timelineSessionEntry) {
	dir := s.ctl().SessionDir()
	if dir == "" {
		return nil, nil
	}
	currentPath := strings.TrimSpace(s.ctl().SessionPath())
	current := ""
	if currentPath != "" {
		current = filepath.Clean(currentPath)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}
	currentListed := false
	out := make([]timelineSessionEntry, 0, len(entries))
	for _, file := range entries {
		if file.IsDir() || !store.IsSessionTranscriptName(file.Name()) {
			continue
		}
		path := filepath.Join(dir, file.Name())
		if agent.IsCleanupPending(path) {
			continue
		}
		updatedAt := agent.SessionContentModTime(path)
		if updatedAt.IsZero() {
			if info, infoErr := file.Info(); infoErr == nil {
				updatedAt = info.ModTime()
			}
		}
		name := strings.TrimSuffix(file.Name(), ".jsonl")
		entry := timelineSessionEntry{
			Name:      name,
			Path:      path,
			Current:   current != "" && filepath.Clean(path) == current,
			UpdatedAt: updatedAt.UTC(),
		}
		currentListed = currentListed || entry.Current
		if first, turns := agent.SessionPreview(path); turns > 0 {
			entry.Turns = turns
			entry.Title = s.sessionTitle(r.Context(), file.Name(), first, updatedAt.UnixNano())
		}
		out = append(out, entry)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].Name > out[j].Name
		}
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	if current == "" || currentListed || filepath.Clean(filepath.Dir(current)) != filepath.Clean(dir) {
		return out, nil
	}
	name := strings.TrimSuffix(filepath.Base(current), ".jsonl")
	draft := &timelineSessionEntry{
		Name:      name,
		Path:      currentPath,
		Title:     "New session 新会话",
		Current:   true,
		Draft:     true,
		UpdatedAt: time.Now().UTC(),
	}
	return out, draft
}

func filterTimelineSessions(entries []timelineSessionEntry, query string) []timelineSessionEntry {
	if query == "" {
		return entries
	}
	filtered := make([]timelineSessionEntry, 0, len(entries))
	for _, entry := range entries {
		if timelineSessionMatches(entry, query) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func timelineSessionMatches(entry timelineSessionEntry, query string) bool {
	if query == "" {
		return true
	}
	for _, value := range []string{entry.Title, entry.Name, entry.Path, strconv.Itoa(entry.Turns)} {
		if strings.Contains(strings.ToLower(value), query) {
			return true
		}
	}
	return false
}

func sessionsOlderThanCursor(entries []timelineSessionEntry, cursor timelineSessionCursor) []timelineSessionEntry {
	for index, entry := range entries {
		stamp := entry.UpdatedAt.UnixNano()
		if stamp < cursor.Time || stamp == cursor.Time && entry.Name < cursor.Name {
			return entries[index:]
		}
	}
	return nil
}

func encodeTimelineSessionCursor(entry timelineSessionEntry) string {
	raw, _ := json.Marshal(timelineSessionCursor{Time: entry.UpdatedAt.UnixNano(), Name: entry.Name})
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeTimelineSessionCursor(raw string) (timelineSessionCursor, bool, error) {
	if strings.TrimSpace(raw) == "" {
		return timelineSessionCursor{}, false, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return timelineSessionCursor{}, false, err
	}
	var cursor timelineSessionCursor
	if err := json.Unmarshal(data, &cursor); err != nil || cursor.Time == 0 || cursor.Name == "" {
		return timelineSessionCursor{}, false, errInvalidTimelineLimit
	}
	return cursor, true, nil
}
