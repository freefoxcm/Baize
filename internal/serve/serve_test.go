package serve

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"reasonix/internal/agent"
	"reasonix/internal/config"
	"reasonix/internal/control"
	"reasonix/internal/event"
	"reasonix/internal/eventwire"
	"reasonix/internal/jobs"
	"reasonix/internal/permission"
	"reasonix/internal/provider"
	_ "reasonix/internal/provider/anthropic" // register the kind: boot.Build resolves the default deepseek-flash entry, whose kind is now anthropic
	"reasonix/internal/tool"
)

func TestTitlePromptRequiresUserMessageLanguage(t *testing.T) {
	if !strings.Contains(titlePrompt, "same language as the user's message") {
		t.Fatalf("title prompt does not preserve the user's language: %q", titlePrompt)
	}
}

type titleUsageProvider struct{}

func (titleUsageProvider) Name() string { return "title" }
func (titleUsageProvider) Stream(context.Context, provider.Request) (<-chan provider.Chunk, error) {
	ch := make(chan provider.Chunk, 3)
	ch <- provider.Chunk{Type: provider.ChunkText, Text: "Short title"}
	ch <- provider.Chunk{Type: provider.ChunkUsage, Usage: &provider.Usage{PromptTokens: 10, CompletionTokens: 2, TotalTokens: 12}}
	ch <- provider.Chunk{Type: provider.ChunkDone}
	close(ch)
	return ch, nil
}

type titleUsageSink struct{ events []event.Event }

func (s *titleUsageSink) Emit(e event.Event) { s.events = append(s.events, e) }

func TestGenerateTitleRecordsUsageWithModelIdentity(t *testing.T) {
	sink := &titleUsageSink{}
	s := &Server{
		titleProv:      titleUsageProvider{},
		titleModelRef:  "deepseek/deepseek-v4-flash",
		titleUsageSink: sink,
	}
	if got := s.generateTitle(context.Background(), "hello"); got != "Short title" {
		t.Fatalf("title = %q", got)
	}
	if len(sink.events) != 1 || sink.events[0].Kind != event.Usage || sink.events[0].ModelRef != "deepseek/deepseek-v4-flash" {
		t.Fatalf("title usage event = %+v", sink.events)
	}
}

// fakeRunner stands in for an agent.Runner: it records the composed input and
// returns without emitting model events, so the controller's TurnDone is the
// observable signal.
type fakeRunner struct{ got chan string }

func (f fakeRunner) Run(_ context.Context, input string) error { f.got <- input; return nil }

type deliveryRecoveryController struct {
	control.SessionAPI
	goal         string
	goalStatus   string
	resumeResult bool
	resumed      bool
	submitted    chan [2]string
}

func (c *deliveryRecoveryController) Goal() string       { return c.goal }
func (c *deliveryRecoveryController) GoalStatus() string { return c.goalStatus }
func (c *deliveryRecoveryController) ResumeGoal() bool {
	c.resumed = true
	if c.resumeResult {
		c.goalStatus = control.GoalStatusRunning
	}
	return c.resumeResult
}
func (c *deliveryRecoveryController) SubmitDeliveryRecovery(display, input string) {
	c.submitted <- [2]string{display, input}
}

func TestServeSubmitRunsAndBroadcastsTurnDone(t *testing.T) {
	bc := NewBroadcaster()
	got := make(chan string, 1)
	ctrl := control.New(control.Options{Runner: fakeRunner{got: got}, Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	sub, cancel := bc.Subscribe() // observe the broadcast deterministically
	defer cancel()

	resp, err := http.Post(srv.URL+"/submit", "application/json", strings.NewReader(`{"input":"hi"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("submit status = %d, want 202", resp.StatusCode)
	}

	select {
	case in := <-got:
		if in != "hi" {
			t.Errorf("runner ran %q, want hi", in)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runner never ran")
	}

	deadline := time.After(2 * time.Second)
	for {
		select {
		case data := <-sub:
			var w eventwire.Event
			if err := json.Unmarshal(data, &w); err == nil && w.Kind == "turn_done" {
				return
			}
		case <-deadline:
			t.Fatal("never saw turn_done on the stream")
		}
	}
}

func TestServeDeliveryRecoveryEndpoint(t *testing.T) {
	baseBC := NewBroadcaster()
	base := control.New(control.Options{Sink: baseBC})
	wrapped := &deliveryRecoveryController{
		SessionAPI:   base,
		goal:         "ship the change",
		goalStatus:   control.GoalStatusBlocked,
		resumeResult: true,
		submitted:    make(chan [2]string, 1),
	}
	srv := httptest.NewServer(New(wrapped, baseBC, config.ServeConfig{}).Handler())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/delivery-recovery", "application/json", strings.NewReader(`{"input":" Continue checks "}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("delivery recovery status = %d, want 202", resp.StatusCode)
	}
	if !wrapped.resumed {
		t.Fatal("blocked Goal was not resumed before delivery recovery")
	}
	select {
	case got := <-wrapped.submitted:
		if got != [2]string{"Continue checks", "Continue checks"} {
			t.Fatalf("delivery recovery submission = %#v", got)
		}
	default:
		t.Fatal("delivery recovery was not submitted")
	}
}

func TestServeDeliveryRecoveryRejectsInvalidOrUnresumableRequests(t *testing.T) {
	baseBC := NewBroadcaster()
	base := control.New(control.Options{Sink: baseBC})
	wrapped := &deliveryRecoveryController{
		SessionAPI: base,
		goal:       "already complete",
		goalStatus: control.GoalStatusComplete,
		submitted:  make(chan [2]string, 1),
	}
	handler := New(wrapped, baseBC, config.ServeConfig{}).Handler()

	cases := []struct {
		body string
		want int
	}{
		{body: `{}`, want: http.StatusBadRequest},
		{body: `{"input":"!rm file"}`, want: http.StatusForbidden},
		{body: `{"input":"continue"}`, want: http.StatusConflict},
	}
	for _, tc := range cases {
		recorder := httptest.NewRecorder()
		req := localTestRequest(http.MethodPost, "/delivery-recovery", strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(recorder, req)
		if recorder.Code != tc.want {
			t.Fatalf("delivery recovery body %s status = %d, want %d", tc.body, recorder.Code, tc.want)
		}
	}
	select {
	case got := <-wrapped.submitted:
		t.Fatalf("invalid recovery unexpectedly submitted: %#v", got)
	default:
	}
}

// TestHistoryMessagesCarriesToolDuration verifies /history serializes the
// persisted tool-execution duration so the web UI can show it after a rebuild.
func TestHistoryMessagesCarriesToolDuration(t *testing.T) {
	msgs := []provider.Message{
		{Role: provider.RoleAssistant, Content: "plan", ToolCalls: []provider.ToolCall{{ID: "c1", Name: "bash", Arguments: "echo"}, {ID: "c2", Name: "edit_file", Arguments: "{}", Added: 3, Removed: 1}}},
		{Role: provider.RoleTool, ToolCallID: "c1", Name: "bash", Content: "out", ToolDurationMs: 2500},
	}
	hm := historyMessages(msgs)
	var toolMsg historyMessage
	for _, m := range hm {
		if m.Role == "tool" {
			toolMsg = m
		}
	}
	if toolMsg.ToolCallID != "c1" {
		t.Fatalf("tool message missing: %+v", hm)
	}
	if toolMsg.DurationMs != 2500 {
		t.Fatalf("durationMs = %d, want 2500", toolMsg.DurationMs)
	}
	b, err := json.Marshal(toolMsg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"durationMs":2500`) {
		t.Fatalf("history JSON omits durationMs: %s", b)
	}

	// The assistant message's tool calls carry the +/- line tallies the web UI
	// shows on history-rebuilt diff cards.
	var assistantMsg historyMessage
	for _, m := range hm {
		if m.Role == "assistant" {
			assistantMsg = m
		}
	}
	if len(assistantMsg.ToolCalls) != 2 {
		t.Fatalf("assistant tool calls = %+v", assistantMsg.ToolCalls)
	}
	if assistantMsg.ToolCalls[1].Added != 3 || assistantMsg.ToolCalls[1].Removed != 1 {
		t.Fatalf("diff tallies = +%d -%d, want +3 -1", assistantMsg.ToolCalls[1].Added, assistantMsg.ToolCalls[1].Removed)
	}
	b, err = json.Marshal(assistantMsg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"added":3`) || !strings.Contains(string(b), `"removed":1`) {
		t.Fatalf("history JSON omits diff tallies: %s", b)
	}
	// A call without tallies omits both fields (old sessions).
	if strings.Contains(string(b), `"added":0`) || strings.Contains(string(b), `"removed":0`) {
		t.Fatalf("zero tallies should be omitted: %s", b)
	}

	// Zero duration (old sessions) is omitted from the wire form.
	msgs[1].ToolDurationMs = 0
	hm = historyMessages(msgs)
	toolMsg = historyMessage{}
	for _, m := range hm {
		if m.Role == "tool" {
			toolMsg = m
		}
	}
	b, err = json.Marshal(toolMsg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "durationMs") {
		t.Fatalf("zero duration should be omitted: %s", b)
	}
}

// TestHistoryMessagesDurationFallback verifies that tool results from
// sessions predating ToolDurationMs persistence estimate their duration from
// the CreatedAt delta with the issuing assistant message.
func TestHistoryMessagesDurationFallback(t *testing.T) {
	msgs := []provider.Message{
		{Role: provider.RoleAssistant, Content: "plan", CreatedAt: 1000, ToolCalls: []provider.ToolCall{{ID: "c1", Name: "bash", Arguments: "echo"}}},
		{Role: provider.RoleTool, ToolCallID: "c1", Name: "bash", Content: "out", CreatedAt: 4234},
	}
	hm := historyMessages(msgs)
	if hm[1].DurationMs != 3234 {
		t.Fatalf("fallback duration = %d, want 3234 (4234-1000)", hm[1].DurationMs)
	}

	// Recorded durations win over the fallback.
	msgs[1].ToolDurationMs = 2500
	hm = historyMessages(msgs)
	if hm[1].DurationMs != 2500 {
		t.Fatalf("recorded duration = %d, want 2500", hm[1].DurationMs)
	}
	msgs[1].ToolDurationMs = 0

	// A parallel batch shares the issuing assistant's start; each result gets
	// its own span.
	msgs = append(msgs,
		provider.Message{Role: provider.RoleAssistant, Content: "plan2", CreatedAt: 5000, ToolCalls: []provider.ToolCall{{ID: "c2", Name: "bash", Arguments: "x"}, {ID: "c3", Name: "bash", Arguments: "y"}}},
		provider.Message{Role: provider.RoleTool, ToolCallID: "c2", Name: "bash", Content: "a", CreatedAt: 6400},
		provider.Message{Role: provider.RoleTool, ToolCallID: "c3", Name: "bash", Content: "b", CreatedAt: 7200},
	)
	hm = historyMessages(msgs)
	if hm[3].DurationMs != 1400 || hm[4].DurationMs != 2200 {
		t.Fatalf("batch fallback durations = %d/%d, want 1400/2200", hm[3].DurationMs, hm[4].DurationMs)
	}

	// A standalone tool result without any issuing assistant omits the field.
	hm = historyMessages([]provider.Message{{Role: provider.RoleTool, ToolCallID: "x", Name: "bash", Content: "o", CreatedAt: 1234}})
	if hm[0].DurationMs != 0 {
		t.Fatalf("standalone duration = %d, want 0", hm[0].DurationMs)
	}
}

func TestServeEndpoints(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc}) // no runner needed for these
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	if resp, err := http.Get(srv.URL + "/history"); err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("history = %v / %v", resp, err)
	}

	if resp, _ := http.Get(srv.URL + "/context"); resp.StatusCode != http.StatusOK {
		t.Errorf("context status = %d", resp.StatusCode)
	}

	resp, err := http.Post(srv.URL+"/plan", "application/json", strings.NewReader(`{"on":true}`))
	if err != nil || resp.StatusCode != http.StatusNoContent {
		t.Fatalf("plan = %v / status %d", err, resp.StatusCode)
	}
	if c := ctrl.Compose("x"); !strings.Contains(c, "Plan mode") {
		t.Error("/plan {on:true} should have enabled plan mode (Compose would prepend the marker)")
	}

	resp, err = http.Post(srv.URL+"/tool-approval-mode", "application/json", strings.NewReader(`{"mode":"auto"}`))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("tool approval mode auto status = %d, want 204", resp.StatusCode)
	}
	resp.Body.Close()
	if got := ctrl.ToolApprovalMode(); got != control.ToolApprovalAuto {
		t.Fatalf("tool approval mode = %q, want auto", got)
	}
	resp, err = http.Post(srv.URL+"/tool-approval-mode", "application/json", strings.NewReader(`{"mode":"surprise"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid tool approval mode status = %d, want 400", resp.StatusCode)
	}

	if resp, _ := http.Post(srv.URL+"/submit", "application/json", strings.NewReader(`{}`)); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("empty submit should be 400, got %d", resp.StatusCode)
	}
}

// steerBlockingProvider keeps the model stream open until the turn's context
// is cancelled, so steer admission can be observed while the turn is active.
// It signals through started once the agent has entered the tool loop (the
// steer intake is open only from that point on).
type steerBlockingProvider struct {
	started chan struct{}
}

func (steerBlockingProvider) Name() string { return "steer-test" }

func (p steerBlockingProvider) Stream(ctx context.Context, _ provider.Request) (<-chan provider.Chunk, error) {
	select {
	case p.started <- struct{}{}:
	default:
	}
	ch := make(chan provider.Chunk)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// TestServeSteer covers the /steer endpoint contract: whitespace text is
// rejected (400), steer without an active turn is refused (409, the client
// keeps the text queued and retries), and an active turn accepts the guidance
// (202) until the turn ends.
func TestServeSteer(t *testing.T) {
	bc := NewBroadcaster()
	started := make(chan struct{}, 1)
	ag := agent.New(steerBlockingProvider{started: started}, tool.NewRegistry(), agent.NewSession(""), agent.Options{}, bc)
	ctrl := control.New(control.Options{Runner: ag, Executor: ag, Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	postSteer := func(text string) int {
		resp, err := http.Post(srv.URL+"/steer", "application/json", strings.NewReader(`{"text":`+strconv.Quote(text)+`}`))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}

	// Whitespace-only text is rejected before any steer admission logic.
	if got := postSteer("  "); got != http.StatusBadRequest {
		t.Fatalf("empty steer status = %d, want 400", got)
	}

	// No active turn: the steer is not queued (client keeps it and retries).
	if got := postSteer("late guidance"); got != http.StatusConflict {
		t.Fatalf("idle steer status = %d, want 409", got)
	}

	// Start a turn and keep it alive, then steer into it.
	resp, err := http.Post(srv.URL+"/submit", "application/json", strings.NewReader(`{"input":"work"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("submit status = %d, want 202", resp.StatusCode)
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("agent never entered the tool loop")
	}
	if got := postSteer("keep the diff small"); got != http.StatusAccepted {
		t.Fatalf("active steer status = %d, want 202", got)
	}

	// Once the turn ends, steer admission closes again.
	resp, err = http.Post(srv.URL+"/cancel", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	waitNotRunning(t, ctrl)
	if got := postSteer("after the turn"); got != http.StatusConflict {
		t.Fatalf("post-turn steer status = %d, want 409", got)
	}
}

func TestServeSubmitRejectsShellShortcut(t *testing.T) {
	bc := NewBroadcaster()
	got := make(chan string, 1)
	ctrl := control.New(control.Options{Runner: fakeRunner{got: got}, Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/submit", "application/json", strings.NewReader(`{"input":"!echo nope"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("shell submit status = %d, want 403", resp.StatusCode)
	}
	select {
	case in := <-got:
		t.Fatalf("runner should not run shell submit, got %q", in)
	default:
	}
}

func TestServeSubmitValidatesFormat(t *testing.T) {
	bc := NewBroadcaster()
	got := make(chan string, 1)
	ctrl := control.New(control.Options{Runner: fakeRunner{got: got}, Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	post := func(body string) int {
		resp, err := http.Post(srv.URL+"/submit", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		return resp.StatusCode
	}

	// Unsupported format is rejected with 400 and the runner never runs.
	if code := post(`{"input":"hi","format":"xml"}`); code != http.StatusBadRequest {
		t.Fatalf("unsupported format status = %d, want 400", code)
	}
	select {
	case in := <-got:
		t.Fatalf("runner must not run for rejected format, got %q", in)
	default:
	}

	// Whitespace-padded json_object is normalized and accepted.
	if code := post(`{"input":"hi","format":"  json_object  "}`); code != http.StatusAccepted {
		t.Fatalf("padded json_object status = %d, want 202", code)
	}
	select {
	case in := <-got:
		if in != "hi" {
			t.Fatalf("runner ran %q, want hi", in)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runner never ran for padded json_object")
	}
}

func TestHistoryMessagesPreserveToolDetails(t *testing.T) {
	got := historyMessages([]provider.Message{
		{Role: provider.RoleUser, Content: "run command"},
		{Role: provider.RoleAssistant, Content: "checking", ReasoningContent: "think", ToolCalls: []provider.ToolCall{{
			ID: "call_1", Name: "bash", Arguments: `{"command":"pwd"}`,
		}}},
		{Role: provider.RoleTool, Name: "bash", ToolCallID: "call_1", Content: "/tmp/project\n"},
	})

	if len(got) != 3 {
		t.Fatalf("history length = %d, want 3", len(got))
	}
	if got[1].Reasoning != "think" {
		t.Fatalf("assistant reasoning = %q, want think", got[1].Reasoning)
	}
	if len(got[1].ToolCalls) != 1 || got[1].ToolCalls[0].ID != "call_1" || got[1].ToolCalls[0].Name != "bash" || got[1].ToolCalls[0].Arguments != `{"command":"pwd"}` {
		t.Fatalf("assistant tool calls not preserved: %+v", got[1].ToolCalls)
	}
	if got[2].ToolCallID != "call_1" || got[2].ToolName != "bash" || got[2].Content != "/tmp/project\n" {
		t.Fatalf("tool result details not preserved: %+v", got[2])
	}
}

func TestHistoryMessagesStripTransientReasoningLanguageBlock(t *testing.T) {
	got := historyMessages([]provider.Message{
		{Role: provider.RoleUser, Content: "<reasoning-language>\nVisible reasoning/thinking text preference: use English.\n</reasoning-language>\n\nExplain this module"},
		{Role: provider.RoleAssistant, Content: "ok"},
	})
	if len(got) != 2 {
		t.Fatalf("history length = %d, want 2: %+v", len(got), got)
	}
	if got[0].Role != "user" || got[0].Content != "Explain this module" {
		t.Fatalf("user history = %+v, want plain user text without reasoning-language", got[0])
	}
	if strings.Contains(got[0].Content, "<reasoning-language>") {
		t.Fatalf("reasoning-language leaked into /history user content: %q", got[0].Content)
	}
}

func TestSessionsListPreviewStripsTransientReasoningLanguageBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.jsonl")
	s := agent.NewSession("system")
	s.Add(provider.Message{Role: provider.RoleUser, Content: "<reasoning-language>\nVisible reasoning/thinking text preference: use English.\n</reasoning-language>\n\nExplain this module"})
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}

	preview, turns := agent.SessionPreview(path)
	if turns != 1 {
		t.Errorf("turns = %d, want 1", turns)
	}
	if preview != "Explain this module" {
		t.Errorf("preview = %q, want user prompt", preview)
	}
}

func TestSessionsListPreviewSeesEventLogTurns(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.jsonl")
	s := agent.NewSession("system")
	s.Add(provider.Message{Role: provider.RoleUser, Content: "first"})
	if err := s.SaveSnapshot(path); err != nil {
		t.Fatal(err)
	}
	s.Add(provider.Message{Role: provider.RoleAssistant, Content: "reply"})
	s.Add(provider.Message{Role: provider.RoleUser, Content: "second"})
	if err := s.SaveSnapshot(path); err != nil {
		t.Fatal(err)
	}

	// The second turn lives only in the event log; a checkpoint-only reader
	// would still report one turn.
	if _, turns := agent.SessionPreview(path); turns != 2 {
		t.Errorf("turns = %d, want 2 (event log turns visible)", turns)
	}
	if mod := agent.SessionContentModTime(path); mod.IsZero() {
		t.Error("SessionContentModTime returned zero for a live session")
	}
}

func TestServeCancelEndpoint(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/cancel", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("cancel status = %d, want 204", resp.StatusCode)
	}
}

func TestServeApproveMissingID(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	// Missing id should return 400.
	resp, err := http.Post(srv.URL+"/approve", "application/json", strings.NewReader(`{"allow":true}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("approve missing id = %d, want 400", resp.StatusCode)
	}

	// Malformed JSON should return 400.
	resp2, _ := http.Post(srv.URL+"/approve", "application/json", strings.NewReader(`{bad`))
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusBadRequest {
		t.Errorf("approve bad json = %d, want 400", resp2.StatusCode)
	}
}

func TestServeCompactEndpoint(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/compact", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("compact = %d, want 204", resp.StatusCode)
	}
}

func TestServeIndexPage(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("index status = %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("index content-type = %q, want text/html", ct)
	}
}

func TestServeEditResubmitsMessage(t *testing.T) {
	bc := NewBroadcaster()
	got := make(chan string, 1)
	ctrl := control.New(control.Options{Runner: fakeRunner{got: got}, Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	sub, cancel := bc.Subscribe()
	defer cancel()

	resp, err := http.Post(srv.URL+"/edit", "application/json",
		strings.NewReader(`{"display":"fixed text","input":"fixed text","original":"original text"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("edit status = %d, want 202", resp.StatusCode)
	}

	select {
	case in := <-got:
		if in != "fixed text" {
			t.Errorf("runner ran %q, want the edited input", in)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runner never ran")
	}

	deadline := time.After(2 * time.Second)
	for {
		select {
		case data := <-sub:
			var w eventwire.Event
			if err := json.Unmarshal(data, &w); err == nil && w.Kind == "turn_done" {
				return
			}
		case <-deadline:
			t.Fatal("never saw turn_done on the stream")
		}
	}
}

func TestServeEditRejectsEmptyAndShell(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	for name, body := range map[string]string{
		"empty input": `{"display":"","input":"","original":""}`,
		"shell input": `{"display":"!ls","input":"!ls","original":"x"}`,
		"malformed":   `not json`,
	} {
		resp, err := http.Post(srv.URL+"/edit", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusForbidden {
			t.Errorf("%s: status = %d, want 400/403", name, resp.StatusCode)
		}
	}
}

func TestServeIndexDefinesQueryHelpers(t *testing.T) {
	html := baizeWebSource()
	for _, want := range []string{
		"const $ = s => document.querySelector(s);",
		"const $$ = s => document.querySelectorAll(s);",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("serve index missing query helper %q", want)
		}
	}
}

func TestServeIndexKeepsToolCardHeaderOnOneRow(t *testing.T) {
	html := baizeWebSource()
	for _, want := range []string{
		".card-main{display:flex;align-items:center;gap:8px;min-width:0}",
		".card-title{display:flex;align-items:center;gap:7px;min-width:0;flex:1 1 auto}",
		".card-head .subject{font-family:var(--mono);font-size:11.5px;color:var(--muted);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;min-width:0;max-width:60%}",
		".card-meta{display:flex;align-items:center;gap:8px;flex:0 0 auto;white-space:nowrap;font-family:var(--mono);font-size:10.5px;color:var(--muted-2)}",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("serve index missing single-row tool-card header CSS %q", want)
		}
	}
	for _, unwanted := range []string{
		".card-main{display:grid;gap:2px;min-width:0}",
		".card-head .name{color:var(--fg);font-weight:500;font-size:12.5px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;flex:0 1 auto;max-width:55%}",
	} {
		if strings.Contains(html, unwanted) {
			t.Fatalf("serve index restored superseded tool-card header CSS %q", unwanted)
		}
	}
}

func TestServeBrandingAndAssets(t *testing.T) {
	html := baizeWebSource()
	for _, want := range []string{
		"<title>Baize</title>",
		"href=\"/assets/logo-symbol.svg\"",
		"class=\"activity-rail__brand\"><img src=\"/assets/logo-symbol.svg\" alt=\"Baize\"",
		"alt=\"Baize\" class=\"brand-wordmark brand-wordmark--welcome\"",
		"'placeholder': 'Message Baize...  / for commands'",
		"'placeholder': '给 Baize 发消息...  / 查看命令'",
		"Automatic retries paused. Baize stopped repeated attempts",
		"已暂停自动重试。Baize 已停止重复尝试",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("serve index missing Baize branding %q", want)
		}
	}
	for _, old := range []string{"<title>Reasonix</title>", "alt=\"Reasonix\"", "Message Reasonix", "给 Reasonix 发消息", "Reasonix stopped repeated attempts"} {
		if strings.Contains(html, old) {
			t.Fatalf("serve index still contains retired display branding %q", old)
		}
	}

	login := string(loginHTML)
	for _, want := range []string{"<title>{{.Title}}</title>", "href=\"/assets/logo-symbol.svg\"", "href=\"/assets/login.css\"", "src=\"/assets/logo-wordmark.svg\" alt=\"Baize\""} {
		if !strings.Contains(login, want) {
			t.Fatalf("login page missing Baize branding %q", want)
		}
	}
	if strings.Contains(login, "Reasonix") {
		t.Fatalf("login page still contains retired display branding: %s", login)
	}

	for name, asset := range map[string][]byte{"wordmark": logoWordmarkSVG, "symbol": logoSymbolSVG} {
		body := string(asset)
		if !strings.Contains(body, "aria-label=\"Baize\"") || !strings.Contains(body, "<path ") {
			t.Fatalf("%s SVG is missing Baize accessibility metadata or path geometry", name)
		}
		if strings.Contains(body, "<image") || strings.Contains(body, "Reasonix") {
			t.Fatalf("%s SVG embeds raster content or retired branding", name)
		}
	}
}

func TestServeIndexTokenActivityAndWorkspaceLabel(t *testing.T) {
	html := baizeWebSource()
	for _, want := range []string{
		`'usage_calendar': 'Token activity'`,
		`'cal_range_year': 'This year'`,
		`'cal_range_6m': 'Last 6 months'`,
		`'cal_range_3m': 'Last 3 months'`,
		`data-cal-range="6m" aria-pressed="true"`,
		`'usage_calendar': 'Token活动'`,
		`data-cal-range="year"`,
		`data-cal-range="6m"`,
		`data-cal-range="3m"`,
		`fetch('/usage/calendar?range='+encodeURIComponent(calRange)`,
		`grid.style.gridTemplateColumns='repeat('+calWeeks`,
		`role="tooltip"`,
		`cell.onfocus=()=>showCalTip`,
		`if(e.key==='Escape'&&calSelected)`,
		`const parts=trimmed.split(/[\\/]/);`,
		`const trimmed=raw.replace(/[\\/]+$/,'');`,
		`const cwd=String(s.workspaceRoot||s.cwd||'-');`,
		`.welcome__pill strong{flex:0 0 auto;`,
		`showWelcome(){if(welcome)welcome.style.display='';setUsageCalendarRange('6m',true);}`,
		`.welcome__calendar{width:fit-content;min-width:min(360px,100%);max-width:min(600px,100%);`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("serve index missing Token activity/workspace support %q", want)
		}
	}
	yearPos := strings.Index(html, `data-cal-range="year"`)
	sixMonthPos := strings.Index(html, `data-cal-range="6m"`)
	threeMonthPos := strings.Index(html, `data-cal-range="3m"`)
	if yearPos < 0 || sixMonthPos < 0 || threeMonthPos < 0 || !(yearPos < sixMonthPos && sixMonthPos < threeMonthPos) {
		t.Fatalf("Token activity ranges are not ordered year, 6m, 3m: %d, %d, %d", yearPos, sixMonthPos, threeMonthPos)
	}
	for _, old := range []string{"CAL_DAYS = 120", "/usage/calendar?days=", `data-cal-range="month"`, "AI coding agent", "AI 编码助手"} {
		if strings.Contains(html, old) {
			t.Fatalf("serve index still contains removed welcome content %q", old)
		}
	}
}
func TestServeIndexReportsSessionDeleteFailures(t *testing.T) {
	html := baizeWebSource()
	for _, want := range []string{
		"'cannot_delete_active': 'Cannot delete the active session'",
		"'cannot_delete_active': '无法删除当前会话'",
		"'delete_failed': 'Could not delete the session. Check your connection and try again.'",
		"'delete_failed': '无法删除会话，请检查连接后重试'",
		"if(target&&target.current){showAppToast(__('cannot_delete_active'),'warn');return;}",
		"if(!r.ok){showAppToast((await r.text()).trim()||('HTTP '+r.status),'danger');}",
		"}).catch(()=>showAppToast(__('delete_failed'),'danger'));",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("serve index missing session delete failure handling %q", want)
		}
	}
}

func TestServeIndexHandlesRetryingEvents(t *testing.T) {
	html := baizeWebSource()
	for _, want := range []string{
		"case 'retrying': setRetrying(e.retryAttempt,e.retryMax); break;",
		"if(e.kind!=='retrying')clearRetrying();",
		"'retrying_status': 'Retrying ({attempt}/{max})...'",
		"'retrying_status': '正在重试 ({attempt}/{max})...'",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("serve index missing retrying support %q", want)
		}
	}
}

func TestServeIndexPresentsRecoveryPauseAsNotice(t *testing.T) {
	html := baizeWebSource()
	for _, want := range []string{
		"e.outcome==='recovery_paused'",
		"appendTranscriptNotice('⏸ '+__('recovery_paused'))",
		"'recovery_paused': 'Automatic retries paused. Baize stopped repeated attempts and kept completed work. Send “Continue” to start a fresh attempt, or add instructions to change direction.'",
		"'recovery_paused': '已暂停自动重试。Baize 已停止重复尝试，并保留已完成的工作。发送“继续”即可开始新一轮，也可以补充要求来调整方向。'",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("serve index missing recovery pause support %q", want)
		}
	}
}

func TestServeIndexPresentsFinalReadinessAsRecoverableNotice(t *testing.T) {
	html := baizeWebSource()
	for _, want := range []string{
		"e.outcome==='final_readiness'",
		"showDeliveryReadiness(e)",
		"post('/submit',{input:prompt,action:'final_readiness_recovery'})",
		"m.role==='final_readiness'",
		"'delivery_incomplete_title': 'Delivery checks are not complete'",
		"'delivery_incomplete_title': '交付检查尚未完成'",
		"project_check:'delivery_requirement_project_check'",
		"observation:'delivery_requirement_observation'",
		"'delivery_requirement_observation': 'query evidence'",
		"'delivery_requirement_observation': '查询证据'",
		"computation:'delivery_requirement_computation'",
		"'delivery_requirement_computation': 'computation evidence'",
		"'delivery_requirement_computation': '分析计算证据'",
		"clearDeliveryCards()",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("serve index missing final-readiness recovery support %q", want)
		}
	}
}

func TestServeIndexHidesInternalTodoToolsAndMarksSignableTodo(t *testing.T) {
	html := baizeWebSource()
	for _, want := range []string{
		"return n==='todo_write'||n==='exit_plan_mode';",
		"if(hiddenTranscriptTool(tool&&tool.name))return;",
		"filter(tc=>!hiddenTranscriptTool(tc.name))",
		"'todo_signable': 'sign off now'",
		"'todo_signable': '当前可签收'",
		"todoIsPhaseSignoff(todosState,i)?__('todo_phase_signoff'):__('todo_signable')",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("serve index missing internal-tool/todo presentation support %q", want)
		}
	}
}

func TestServeIndexRendersAndReloadsExtensions(t *testing.T) {
	html := baizeWebSource()
	for _, want := range []string{
		"case 'extension_surface': if(e.extension)renderExtensionSurface(e.extension); break;",
		"case 'extension_status': if(e.extension)renderExtensionSurface(e.extension); break;",
		"const node=el('div','notice'",
		"post('/extensions/reload',{})",
		"{cmd:'reload',sig:'/reload',desc:__('cmd_reload'),group:'system'}",
		"{cmd:'reload-cmd',sig:'/reload-cmd'",
		"'cmd_reload': 'Reload tools, skills, MCP and extensions'",
		"'cmd_reload': '重新加载工具、技能、MCP 和扩展'",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("serve index missing extension support %q", want)
		}
	}
	if strings.Contains(html, "p.card.markdown+'</") {
		t.Fatal("extension Markdown must not be inserted as HTML")
	}
}

func TestServeIndexPagePassesLanguagePreferenceToClient(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	html := string(body)
	if !strings.Contains(html, `data-language="auto"`) {
		t.Fatalf("default language preference was not passed as auto:\n%s", html)
	}
	if !strings.Contains(string(baizeJS), "applyStaticI18n();") {
		t.Fatal("index should translate static __('key') placeholders on the client")
	}

	cfgPath := config.UserConfigPath()
	if cfgPath == "" {
		t.Fatal("user config path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, []byte("[desktop]\nlanguage = \"en\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	resp, err = http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `data-language="en"`) {
		t.Fatalf("pinned desktop language was not passed through:\n%s", string(body))
	}
}

func baizeWebSource() string {
	return string(indexHTML) + string(baizeCSS) + string(baizeJS)
}

func TestServeModelsMarksActiveByModelRef(t *testing.T) {
	writeServeModelConfig(t)

	bc := NewBroadcaster()
	ctrl := control.New(control.Options{
		Sink:     bc,
		Label:    "shared-chat",
		ModelRef: "alternate/shared-chat",
	})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/models")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("models status = %d, want 200", resp.StatusCode)
	}
	var body struct {
		Current string `json:"current"`
		Models  []struct {
			Ref    string `json:"ref"`
			Active bool   `json:"active"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode models: %v", err)
	}
	if body.Current != "alternate/shared-chat" {
		t.Fatalf("current = %q, want alternate/shared-chat", body.Current)
	}
	active := map[string]bool{}
	for _, m := range body.Models {
		active[m.Ref] = m.Active
	}
	if active["default/shared-chat"] {
		t.Fatal("default provider was marked active even though the controller is on alternate/shared-chat")
	}
	if !active["alternate/shared-chat"] {
		t.Fatal("alternate/shared-chat was not marked active")
	}
}

func TestServeModelsIncludesExtensionProviderCatalog(t *testing.T) {
	writeServeModelConfig(t)

	bc := NewBroadcaster()
	ref := "plugin/demo/cloud/extension-chat"
	ctrl := control.New(control.Options{
		Sink:     bc,
		Label:    "extension-chat",
		ModelRef: ref,
		ProviderResolver: &provider.StaticResolver{Descriptors: []provider.Descriptor{{
			Ref: ref, Model: "extension-chat", DisplayName: "Extension Chat",
		}},
		},
	})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/models")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body struct {
		Models []struct {
			Ref      string `json:"ref"`
			Provider string `json:"provider"`
			Kind     string `json:"kind"`
			Active   bool   `json:"active"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	for _, model := range body.Models {
		if model.Ref == ref {
			if model.Provider != "plugin/demo/cloud" || model.Kind != "extension" || !model.Active {
				t.Fatalf("extension model = %+v", model)
			}
			return
		}
	}
	t.Fatalf("extension provider %q missing from models: %+v", ref, body.Models)
}

func TestServeExtensionReloadPublishesOnlySuccessfulReplacement(t *testing.T) {
	bc := NewBroadcaster()
	old := control.New(control.Options{Sink: bc, ModelRef: "default/model"})
	s := New(old, bc, config.ServeConfig{})

	wantErr := errors.New("sidecar did not initialize")
	s.rebuildController = func(context.Context, *control.Controller, string) (*control.Controller, error) {
		return nil, wantErr
	}
	if err := s.reloadExtensions(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("reload error = %v, want %v", err, wantErr)
	}
	if s.ctl() != old {
		t.Fatal("failed reload replaced the working controller")
	}

	replacement := control.New(control.Options{Sink: bc, ModelRef: "default/model"})
	s.rebuildController = func(_ context.Context, gotOld *control.Controller, ref string) (*control.Controller, error) {
		if gotOld != old || ref != "default/model" {
			t.Fatalf("rebuild inputs old=%p ref=%q", gotOld, ref)
		}
		return replacement, nil
	}
	if err := s.reloadExtensions(context.Background()); err != nil {
		t.Fatalf("reload: %v", err)
	}
	if s.ctl() != replacement {
		t.Fatal("successful reload did not publish the replacement")
	}
}

func TestServeSwitchEffortUsesModelRefForDuplicateModelNames(t *testing.T) {
	writeServeModelConfig(t)

	bc := NewBroadcaster()
	ctrl := control.New(control.Options{
		Sink:       bc,
		Label:      "shared-chat",
		ModelRef:   "alternate/shared-chat",
		SessionDir: t.TempDir(),
	})
	server := New(ctrl, bc, config.ServeConfig{})
	var builtRef string
	server.buildController = func(_ context.Context, ref string) (*control.Controller, error) {
		builtRef = ref
		return control.New(control.Options{
			Sink:       bc,
			Label:      "shared-chat",
			ModelRef:   ref,
			SessionDir: t.TempDir(),
		}), nil
	}

	if err := server.switchEffort(context.Background(), "high"); err != nil {
		t.Fatalf("switchEffort: %v", err)
	}
	if builtRef != "alternate/shared-chat" {
		t.Fatalf("rebuilt model ref = %q, want alternate/shared-chat", builtRef)
	}
	edit := config.LoadForEdit(config.UserConfigPath())
	def, _ := edit.Provider("default")
	if def.Effort != "" {
		t.Fatalf("default effort = %q, want unchanged", def.Effort)
	}
	alt, _ := edit.Provider("alternate")
	if alt.Effort != "high" {
		t.Fatalf("alternate effort = %q, want high", alt.Effort)
	}
}

func writeServeModelConfig(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	isolateServeHome(t, home)
	cfgPath := config.UserConfigPath()
	if cfgPath == "" {
		t.Fatal("user config path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `default_model = "default/shared-chat"

[[providers]]
name = "default"
kind = "openai"
base_url = "http://127.0.0.1:1/v1"
models = ["shared-chat"]
default = "shared-chat"
supported_efforts = ["low", "high"]

[[providers]]
name = "alternate"
kind = "openai"
base_url = "http://127.0.0.1:2/v1"
models = ["shared-chat"]
default = "shared-chat"
supported_efforts = ["low", "high"]
`
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestResumeRequiresSessionPathInsideSessionDir(t *testing.T) {
	dir := t.TempDir()
	active := filepath.Join(dir, "active.jsonl")
	inside := filepath.Join(dir, "inside.jsonl")
	outsideDir := t.TempDir()
	outside := filepath.Join(outsideDir, "outside.jsonl")
	for _, path := range []string{active, inside, outside} {
		if err := os.WriteFile(path, []byte(`{"role":"user","content":"hi"}`+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc, SessionDir: dir, SessionPath: active})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	post := func(path string) int {
		body, err := json.Marshal(map[string]string{"path": path})
		if err != nil {
			t.Fatal(err)
		}
		resp, err := http.Post(srv.URL+"/resume", "application/json", strings.NewReader(string(body)))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	if got := post(outside); got != http.StatusForbidden {
		t.Fatalf("outside resume status = %d, want 403", got)
	}
	if got := post(inside); got != http.StatusNoContent {
		t.Fatalf("inside resume status = %d, want 204", got)
	}
	want, err := filepath.EvalSymlinks(inside)
	if err != nil {
		t.Fatal(err)
	}
	if got := filepath.Clean(ctrl.SessionPath()); got != filepath.Clean(want) {
		t.Fatalf("session path = %q, want %q", got, want)
	}
}

func TestResumeRejectsCleanupPendingSession(t *testing.T) {
	dir := t.TempDir()
	active := filepath.Join(dir, "active.jsonl")
	pending := filepath.Join(dir, "pending.jsonl")
	for _, path := range []string{active, pending} {
		if err := os.WriteFile(path, []byte(`{"role":"user","content":"hi"}`+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := agent.MarkCleanupPending(pending, "delete"); err != nil {
		t.Fatal(err)
	}

	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc, SessionDir: dir, SessionPath: active})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	body, err := json.Marshal(map[string]string{"path": pending})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(srv.URL+"/resume", "application/json", strings.NewReader(string(body)))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("cleanup-pending resume status = %d, want 400", resp.StatusCode)
	}
	if got := filepath.Clean(ctrl.SessionPath()); got != filepath.Clean(active) {
		t.Fatalf("session path after rejected resume = %q, want active %q", got, active)
	}
}

func TestSessionsSkipsCleanupPending(t *testing.T) {
	dir := t.TempDir()
	active := filepath.Join(dir, "active.jsonl")
	pending := filepath.Join(dir, "pending.jsonl")
	for _, path := range []string{active, pending} {
		if err := os.WriteFile(path, []byte(`{"role":"user","content":"hi"}`+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := agent.MarkCleanupPending(pending, "delete"); err != nil {
		t.Fatal(err)
	}

	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc, SessionDir: dir, SessionPath: active})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/sessions")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var got []struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "active" || got[0].Path != agent.CanonicalSessionPath(active) {
		t.Fatalf("/sessions = %+v, want only active session", got)
	}
}

func TestDeleteSessionRequiresSessionNameInsideSessionDir(t *testing.T) {
	dir := t.TempDir()
	active := filepath.Join(dir, "active.jsonl")
	old := filepath.Join(dir, "old.jsonl")
	for _, path := range []string{active, old} {
		if err := os.WriteFile(path, []byte(`{"role":"user","content":"hi"}`+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	ref := "sa_20260102_030405_000000000_aabbccddeeff"
	writeServeSubagentArtifact(t, dir, ref, agent.BranchID(old))
	oldJobsDir := jobs.ArtifactDir(old)
	if err := os.MkdirAll(oldJobsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldJobsDir, "bash-1.log"), []byte("output"), 0o644); err != nil {
		t.Fatal(err)
	}
	sibling := dir + "-other"
	if err := os.MkdirAll(sibling, 0o755); err != nil {
		t.Fatal(err)
	}
	escape := filepath.Join(sibling, "escape.jsonl")
	if err := os.WriteFile(escape, []byte("keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc, SessionDir: dir, SessionPath: active})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	post := func(body string) int {
		resp, err := http.Post(srv.URL+"/delete-session", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	if got := post(`{"path":"` + escape + `"}`); got != http.StatusBadRequest {
		t.Fatalf("legacy path delete status = %d, want 400", got)
	}
	if got := post(`{"name":"../` + filepath.Base(sibling) + `/escape"}`); got != http.StatusBadRequest {
		t.Fatalf("sibling traversal status = %d, want 400", got)
	}
	if _, err := os.Stat(escape); err != nil {
		t.Fatalf("sibling session was removed: %v", err)
	}
	if got := post(`{"name":"active"}`); got != http.StatusConflict {
		t.Fatalf("active delete status = %d, want 409", got)
	}
	if got := post(`{"name":"old"}`); got != http.StatusNoContent {
		t.Fatalf("valid delete status = %d, want 204", got)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatalf("old session still exists or stat failed unexpectedly: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "subagents", ref+".jsonl")); !os.IsNotExist(err) {
		t.Fatalf("old session subagent jsonl still exists or stat failed unexpectedly: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "subagents", ref+".meta.json")); !os.IsNotExist(err) {
		t.Fatalf("old session subagent meta still exists or stat failed unexpectedly: %v", err)
	}
	if _, err := os.Stat(oldJobsDir); !os.IsNotExist(err) {
		t.Fatalf("old session jobs sidecar still exists or stat failed unexpectedly: %v", err)
	}
}

func writeServeSubagentArtifact(t *testing.T, dir, ref, parentSession string) {
	t.Helper()
	subagentDir := filepath.Join(dir, "subagents")
	if err := os.MkdirAll(subagentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subagentDir, ref+".jsonl"), []byte(`{"role":"user","content":"sub"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(agent.SubagentMeta{
		Ref:           ref,
		Status:        agent.SubagentCompleted,
		Kind:          "task",
		Name:          "task",
		ParentSession: parentSession,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subagentDir, ref+".meta.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestServeSubmitMalformedJSON(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/submit", "application/json", strings.NewReader(`{not json`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("malformed submit = %d, want 400", resp.StatusCode)
	}
}

func TestServePlanMalformedJSON(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/plan", "application/json", strings.NewReader(`{bad`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("malformed plan = %d, want 400", resp.StatusCode)
	}
}

func TestServeContextEndpoint(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/context")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("context status = %d", resp.StatusCode)
	}
	var body map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode context: %v", err)
	}
	// Before any turn, used should be 0.
	if body["used"] != 0 {
		t.Errorf("used = %d, want 0", body["used"])
	}
}

// planApprovingRunner simulates the planner runner: it requests host approval
// for the plan (via the controller-wired PlannerPlanApprover) before executing
// the plan body, exactly like boot's planner path.
type planApprovingRunner struct {
	approver agent.PlannerPlanApprover
}

func (r *planApprovingRunner) SetPlannerPlanApprover(a agent.PlannerPlanApprover) { r.approver = a }
func (r *planApprovingRunner) Run(ctx context.Context, input string) error {
	executed := false
	err := r.approver.RunWithPlannerApproval(ctx, "plan text", func(ctx context.Context) error {
		executed = true
		return nil
	})
	if err != nil {
		return err
	}
	_ = executed
	return nil
}

// TestServePlanApprovalPostureMatrix drives a plan-mode turn through the serve
// HTTP stack and proves the plan approval (exit_plan_mode, a fresh human
// decision) surfaces as an approval_request in every tool-approval mode —
// auto/yolo must never silently approve a plan.
func TestServePlanApprovalPostureMatrix(t *testing.T) {
	for _, mode := range []string{"ask", "auto", "yolo"} {
		t.Run(mode, func(t *testing.T) {
			bc := NewBroadcaster()
			runner := &planApprovingRunner{}
			ctrl := control.New(control.Options{Runner: runner, Sink: bc})
			ctrl.EnableInteractiveApproval()
			ctrl.SetPlanMode(true)
			ctrl.SetToolApprovalMode(mode)
			srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
			defer srv.Close()

			sub, cancel := bc.Subscribe()
			defer cancel()

			if _, err := http.Post(srv.URL+"/submit", "application/json", strings.NewReader(`{"input":"draft a plan"}`)); err != nil {
				t.Fatal(err)
			}
			approvalID := ""
			deadline := time.After(5 * time.Second)
		wait:
			for {
				select {
				case data := <-sub:
					var w eventwire.Event
					if err := json.Unmarshal(data, &w); err == nil && w.Kind == "approval_request" && w.Approval != nil && w.Approval.Tool == "exit_plan_mode" {
						approvalID = w.Approval.ID
						break wait
					}
				case <-deadline:
					t.Fatalf("mode %s: no exit_plan_mode approval_request received", mode)
				}
			}
			// Approve the plan through the HTTP endpoint; the plan body must
			// then run to completion.
			payload := `{"id":"` + approvalID + `","allow":true,"session":false}`
			resp, err := http.Post(srv.URL+"/approve", "application/json", strings.NewReader(payload))
			if err != nil {
				t.Fatal(err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusNoContent {
				t.Fatalf("approve status = %d", resp.StatusCode)
			}
			deadline = time.After(5 * time.Second)
			for {
				select {
				case data := <-sub:
					var w eventwire.Event
					if err := json.Unmarshal(data, &w); err == nil && w.Kind == "turn_done" {
						return // plan approved and executed
					}
				case <-deadline:
					t.Fatalf("mode %s: turn did not complete after approval", mode)
				}
			}
		})
	}
}

func TestUsageCalendarRange(t *testing.T) {
	loc := time.FixedZone("test", 8*60*60)
	tests := []struct {
		name, now, preset, wantKey, wantFrom, wantTo string
		wantErr                                      bool
	}{
		{name: "default six months", now: "2026-08-04", wantKey: "6m", wantFrom: "2026-02-04", wantTo: "2026-08-04"},
		{name: "month is no longer supported", now: "2026-08-04", preset: "month", wantErr: true},
		{name: "year", now: "2026-08-04", preset: "year", wantKey: "year", wantFrom: "2026-01-01", wantTo: "2026-08-04"},
		{name: "three months", now: "2026-08-04", preset: "3m", wantKey: "3m", wantFrom: "2026-05-04", wantTo: "2026-08-04"},
		{name: "six months crosses year", now: "2026-03-04", preset: "6m", wantKey: "6m", wantFrom: "2025-09-04", wantTo: "2026-03-04"},
		{name: "month end clamps", now: "2025-05-31", preset: "3m", wantKey: "3m", wantFrom: "2025-02-28", wantTo: "2025-05-31"},
		{name: "leap month end clamps", now: "2024-05-31", preset: "3m", wantKey: "3m", wantFrom: "2024-02-29", wantTo: "2024-05-31"},
		{name: "invalid", now: "2026-08-04", preset: "90", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			now, err := time.ParseInLocation(usageCalendarDateLayout, tc.now, loc)
			if err != nil {
				t.Fatal(err)
			}
			key, from, to, err := usageCalendarRange(now, tc.preset)
			if tc.wantErr {
				if err == nil {
					t.Fatal("usageCalendarRange error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if key != tc.wantKey || from.Format(usageCalendarDateLayout) != tc.wantFrom || to.Format(usageCalendarDateLayout) != tc.wantTo {
				t.Fatalf("range = %q %s..%s, want %q %s..%s", key, from.Format(usageCalendarDateLayout), to.Format(usageCalendarDateLayout), tc.wantKey, tc.wantFrom, tc.wantTo)
			}
		})
	}
}

// TestServeUsageCalendar drives GET /usage/calendar against a temp stats dir
// seeded with daily stats files (stats record JSONL). Usage/turn rows must
// aggregate into the preset range contract, month boundaries must be honored,
// and rows from other sources (desktop/cli) must be excluded.
func TestServeUsageCalendar(t *testing.T) {
	dir := t.TempDir()
	dayLayout := "2006-01-02"
	writeRow := func(day time.Time, model, source string, total int, turn bool) {
		line := map[string]any{"ts": day.Format(time.RFC3339), "model": model, "source": source, "total": total}
		if turn {
			line = map[string]any{"ts": day.Format(time.RFC3339), "source": source, "turn": true}
		}
		b, err := json.Marshal(line)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, day.Format(dayLayout)+".jsonl")
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(append(b, '\n')); err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
	now := time.Date(2026, time.August, 4, 12, 0, 0, 0, time.FixedZone("test", 8*60*60))
	writeRow(now, "deepseek/deepseek-v4-flash", "serve", 300, false)
	writeRow(now, "deepseek/deepseek-v4-flash", "serve", 700, false)
	writeRow(now, "", "serve", 0, true)
	writeRow(now.AddDate(0, 0, -1), "opencode-go/glm-5.2", "serve", 500, false)
	writeRow(now.AddDate(0, 0, -1), "opencode-go/glm-5.2", "serve", 0, true)
	writeRow(now.AddDate(0, 0, -3), "deepseek/deepseek-v4-pro", "serve", 200, false)
	writeRow(now.AddDate(0, 0, -3), "", "serve", 0, true)
	writeRow(now.AddDate(0, 0, -3), "", "serve", 0, true)
	writeRow(now.AddDate(0, 0, -1), "opencode-go/glm-5.2", "cli", 9999, false)             // other source: excluded
	writeRow(now.AddDate(0, 0, -200), "deepseek/deepseek-v4-flash", "serve", 12345, false) // outside window

	ctrl := control.New(control.Options{})
	bc := NewBroadcaster()
	srv := New(ctrl, bc, config.ServeConfig{})
	srv.statsDir = func() string { return dir }
	srv.now = func() time.Time { return now }
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	var got struct {
		Days []struct {
			Day      string           `json:"day"`
			Tokens   int64            `json:"tokens"`
			Requests int              `json:"requests"`
			Turns    int              `json:"turns"`
			ByModel  map[string]int64 `json:"byModel"`
		} `json:"days"`
		Range      string `json:"range"`
		From       string `json:"from"`
		To         string `json:"to"`
		Max        int64  `json:"max"`
		Total      int64  `json:"total"`
		Turns      int64  `json:"turns"`
		ActiveDays int    `json:"activeDays"`
	}
	resp, err := http.Get(ts.URL + "/usage/calendar")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Range != "6m" || got.From != "2026-02-04" || got.To != "2026-08-04" {
		t.Fatalf("range = %q %s..%s, want 6m 2026-02-04..2026-08-04", got.Range, got.From, got.To)
	}
	if got.Total != 1000+500+200 {
		t.Fatalf("total = %d, want 1700", got.Total)
	}
	if got.Max != 1000 {
		t.Fatalf("max = %d, want 1000", got.Max)
	}
	if got.Turns != 1+1+2 {
		t.Fatalf("turns = %d, want 4", got.Turns)
	}
	if got.ActiveDays != 3 {
		t.Fatalf("activeDays = %d, want 3", got.ActiveDays)
	}
	type daySummary struct {
		tokens          int64
		requests, turns int
		byModel         map[string]int64
	}
	byDay := map[string]daySummary{}
	for _, d := range got.Days {
		byDay[d.Day] = daySummary{tokens: d.Tokens, requests: d.Requests, turns: d.Turns, byModel: d.ByModel}
	}
	today := byDay[now.Format(dayLayout)]
	if today.tokens != 1000 || today.requests != 2 || today.turns != 1 {
		t.Fatalf("today = %+v, want 1000 tokens, 2 requests, 1 turn", today)
	}
	if today.byModel["deepseek/deepseek-v4-flash"] != 1000 {
		t.Fatalf("today model split = %#v, want flash=1000", today.byModel)
	}
	yesterday := byDay[now.AddDate(0, 0, -1).Format(dayLayout)]
	if yesterday.tokens != 500 || yesterday.requests != 1 || yesterday.turns != 1 {
		t.Fatalf("yesterday = %+v, want 500 tokens, 1 request, 1 turn (cli row excluded)", yesterday)
	}
	if yesterday.byModel["opencode-go/glm-5.2"] != 500 {
		t.Fatalf("yesterday model split = %#v, want glm=500", yesterday.byModel)
	}
	if _, ok := byDay[now.AddDate(0, 0, -200).Format(dayLayout)]; ok {
		t.Fatal("out-of-window day leaked")
	}
	// Day ordering must be ascending and contiguous (Query contract).
	if len(got.Days) != 182 {
		t.Fatalf("days = %d, want 182 for February 4..August 4", len(got.Days))
	}
	for i := 1; i < len(got.Days); i++ {
		if got.Days[i].Day <= got.Days[i-1].Day {
			t.Fatalf("days not ascending at %d: %s <= %s", i, got.Days[i].Day, got.Days[i-1].Day)
		}
	}
	for _, tc := range []struct {
		rangeKey string
		from     string
		days     int
	}{
		{rangeKey: "year", from: "2026-01-01", days: 216},
		{rangeKey: "6m", from: "2026-02-04", days: 182},
		{rangeKey: "3m", from: "2026-05-04", days: 93},
	} {
		resp, err := http.Get(ts.URL + "/usage/calendar?range=" + tc.rangeKey)
		if err != nil {
			t.Fatal(err)
		}
		var ranged struct {
			Range string            `json:"range"`
			From  string            `json:"from"`
			To    string            `json:"to"`
			Days  []json.RawMessage `json:"days"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&ranged); err != nil {
			resp.Body.Close()
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("range %q status = %d", tc.rangeKey, resp.StatusCode)
		}
		if ranged.Range != tc.rangeKey || ranged.From != tc.from || ranged.To != "2026-08-04" || len(ranged.Days) != tc.days {
			t.Fatalf("range %q response = %+v (%d days)", tc.rangeKey, ranged, len(ranged.Days))
		}
	}
	bad, err := http.Get(ts.URL + "/usage/calendar?range=bogus")
	if err != nil {
		t.Fatal(err)
	}
	bad.Body.Close()
	if bad.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid range status = %d, want 400", bad.StatusCode)
	}
}

type serveApprovalWriter struct{}

func (serveApprovalWriter) Name() string        { return "serve_write" }
func (serveApprovalWriter) Description() string { return "write a test file" }
func (serveApprovalWriter) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}}}`)
}
func (serveApprovalWriter) ReadOnly() bool { return false }
func (serveApprovalWriter) Execute(context.Context, json.RawMessage) (string, error) {
	return "ok", nil
}

type serveApprovalProvider struct {
	mu   sync.Mutex
	turn int
}

func (p *serveApprovalProvider) Name() string { return "serve-approval-test" }
func (p *serveApprovalProvider) Stream(context.Context, provider.Request) (<-chan provider.Chunk, error) {
	p.mu.Lock()
	turn := p.turn
	p.turn++
	p.mu.Unlock()

	ch := make(chan provider.Chunk, 2)
	if turn == 0 {
		ch <- provider.Chunk{Type: provider.ChunkToolCall, ToolCall: &provider.ToolCall{
			ID: "serve-approval-1", Name: "serve_write", Arguments: `{"path":"a.txt"}`,
		}}
	} else {
		ch <- provider.Chunk{Type: provider.ChunkText, Text: "done"}
	}
	ch <- provider.Chunk{Type: provider.ChunkDone}
	close(ch)
	return ch, nil
}

// receives a still-blocked ask_request. Without replay, the browser attaches to
// a healthy-looking session that never surfaces the parked prompt (#7643).
func TestServeEventsReplaysPendingAskOnAttach(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	ctrl.EnableInteractiveApproval()
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	firstSub, cancelFirst := bc.Subscribe()
	defer cancelFirst()

	askCtx, cancelAsk := context.WithCancel(context.Background())
	askDone := make(chan error, 1)
	go func() {
		_, err := ctrl.Ask(askCtx, []event.AskQuestion{{
			ID: "q1", Prompt: "pick one", Options: []event.AskOption{{Label: "A"}, {Label: "B"}},
		}})
		askDone <- err
	}()

	select {
	case data := <-firstSub:
		if !strings.Contains(string(data), `"kind":"ask_request"`) {
			t.Fatalf("initial subscriber got %s, want ask_request", data)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for initial ask_request")
	}

	resp, err := http.Get(srv.URL + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/events status = %d", resp.StatusCode)
	}

	replayed := make(chan string, 1)
	go func() {
		buf := make([]byte, 0, 4096)
		tmp := make([]byte, 512)
		for {
			n, readErr := resp.Body.Read(tmp)
			if n > 0 {
				buf = append(buf, tmp[:n]...)
				if strings.Contains(string(buf), `"kind":"ask_request"`) {
					replayed <- string(buf)
					return
				}
			}
			if readErr != nil {
				return
			}
		}
	}()

	select {
	case <-replayed:
	case <-time.After(2 * time.Second):
		t.Fatal("late SSE attach never received replayed ask_request")
	}

	select {
	case err := <-askDone:
		t.Fatalf("ask resolved before the late client answered: %v", err)
	default:
	}

	// Reconnect recovery must be connection-local: the existing subscriber
	// must not receive the same prompt a second time.
	select {
	case data := <-firstSub:
		t.Fatalf("existing subscriber got duplicate replay: %s", data)
	default:
	}

	cancelAsk()
	select {
	case <-askDone:
	case <-time.After(2 * time.Second):
		t.Fatal("blocked ask did not exit after test cancellation")
	}
}

// TestServeEventsReplayHandoffSerializesPromptEmission proves the controller's
// attach handoff can register a subscriber and replay while prompt emission is
// serialized, so a prompt cannot land between those two operations.
func TestServeEventsReplayHandoffSerializesPromptEmission(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	ctrl.EnableInteractiveApproval()

	askCtx, cancelAsk := context.WithCancel(context.Background())
	defer cancelAsk()
	taskDone := make(chan struct{})
	var sub <-chan []byte
	var cancelSub func()
	ctrl.ReplayPendingPromptsWith(func() event.Sink {
		sub, cancelSub = bc.Subscribe()
		go func() {
			_, _ = ctrl.Ask(askCtx, []event.AskQuestion{{
				ID: "q1", Prompt: "pick one", Options: []event.AskOption{{Label: "A"}, {Label: "B"}},
			}})
			close(taskDone)
		}()
		return event.FuncSink(func(e event.Event) { bc.EmitTo(sub, e) })
	})
	defer cancelSub()

	select {
	case data := <-sub:
		if !strings.Contains(string(data), `"kind":"ask_request"`) {
			t.Fatalf("handoff subscriber got %s, want ask_request", data)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handoff subscriber never received ask_request")
	}
	select {
	case data := <-sub:
		t.Fatalf("handoff subscriber got duplicate ask_request: %s", data)
	default:
	}

	cancelAsk()
	select {
	case <-taskDone:
	case <-time.After(2 * time.Second):
		t.Fatal("handoff ask did not exit after cancellation")
	}
}

// TestServeEventsReplaysPendingApprovalOnAttach covers the actual approval
// surface from #7643: a late browser must receive a parked ApprovalRequest and
// be able to answer it through the serve HTTP endpoint.
func TestServeEventsReplaysPendingApprovalOnAttach(t *testing.T) {
	reg := tool.NewRegistry()
	reg.Add(serveApprovalWriter{})
	ag := agent.New(&serveApprovalProvider{}, reg, agent.NewSession(""), agent.Options{}, event.Discard)
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{
		Runner:   ag,
		Executor: ag,
		Sink:     bc,
		Policy:   permission.New("ask", nil, nil, nil),
	})
	ctrl.EnableInteractiveApproval()
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	runDone := make(chan error, 1)
	go func() { runDone <- ctrl.Executor().Run(context.Background(), "write a file") }()

	deadline := time.After(2 * time.Second)
	for !ctrl.PendingPrompt() {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for parked approval")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}

	resp, err := http.Get(srv.URL + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/events status = %d", resp.StatusCode)
	}

	replayed := make(chan eventwire.Event, 1)
	go func() {
		buf := make([]byte, 0, 4096)
		tmp := make([]byte, 512)
		for {
			n, readErr := resp.Body.Read(tmp)
			if n > 0 {
				buf = append(buf, tmp[:n]...)
				if strings.Contains(string(buf), `"kind":"approval_request"`) {
					frame := string(buf)
					start := strings.Index(frame, "data: ")
					if start < 0 {
						return
					}
					end := strings.IndexByte(frame[start:], '\n')
					if end < 0 {
						end = len(frame) - start
					}
					var wire eventwire.Event
					if json.Unmarshal([]byte(strings.TrimSpace(frame[start+len("data: "):start+end])), &wire) == nil {
						replayed <- wire
					}
					return
				}
			}
			if readErr != nil {
				return
			}
		}
	}()

	var approval eventwire.Event
	select {
	case approval = <-replayed:
	case <-time.After(2 * time.Second):
		t.Fatal("late SSE attach never received replayed approval_request")
	}
	if approval.Kind != "approval_request" || approval.Approval == nil || approval.Approval.Tool != "serve_write" {
		t.Fatalf("replayed approval = %+v, want serve_write approval_request", approval)
	}

	payload, err := json.Marshal(map[string]any{"id": approval.Approval.ID, "allow": true})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/approve", strings.NewReader(string(payload)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	answer, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	answer.Body.Close()
	if answer.StatusCode != http.StatusNoContent {
		t.Fatalf("/approve status = %d", answer.StatusCode)
	}

	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("executor run after approval: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("executor did not finish after approval")
	}
}
