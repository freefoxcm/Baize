package serve

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"reasonix/internal/agent"
	"reasonix/internal/config"
	"reasonix/internal/control"
	"reasonix/internal/event"
	"reasonix/internal/provider"
)

type attachmentSessionRunner struct {
	session *agent.Session
	got     chan string
	release chan struct{}
}

func (r attachmentSessionRunner) Run(ctx context.Context, input string) error {
	r.session.Add(provider.Message{Role: provider.RoleUser, Content: input, RawContent: agent.RawUserInput(ctx, input)})
	r.got <- input
	<-r.release
	return nil
}

func attachmentSubmitServer(t *testing.T) (*control.Controller, *httptest.Server, control.AttachmentInfo, chan string, chan struct{}) {
	t.Helper()
	base, ws := testCtrlWithWorkspace(t)
	base.Close()
	bc := NewBroadcaster()
	session := agent.NewSession("system")
	got := make(chan string, 1)
	release := make(chan struct{}, 1)
	runner := attachmentSessionRunner{session: session, got: got, release: release}
	ctrl := control.New(control.Options{
		Runner:        runner,
		Executor:      agent.New(nil, nil, session, agent.Options{}, event.Discard),
		Sink:          bc,
		SessionDir:    t.TempDir(),
		WorkspaceRoot: ws,
	})
	t.Cleanup(ctrl.Close)
	info, err := ctrl.SaveAttachment(context.Background(), "notes.txt", strings.NewReader("trusted attachment body"), -1)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	t.Cleanup(srv.Close)
	return ctrl, srv, info, got, release
}

func TestPrepareAttachmentTurnUsesManagedNamesAndDeduplicates(t *testing.T) {
	ctrl, srv, info, _, _ := attachmentSubmitServer(t)
	server := New(ctrl, NewBroadcaster(), config.ServeConfig{})
	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	turn, err := server.prepareAttachmentTurn(req, "Review this", []string{info.Path, info.Path})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(turn.display, info.Path) != 1 || !strings.Contains(turn.display, "@[notes.txt](") {
		t.Fatalf("display = %q, want one trusted named reference", turn.display)
	}
	if strings.Count(turn.refLine, info.Path) != 1 || !strings.Contains(turn.refLine, "@"+info.Path) {
		t.Fatalf("refLine = %q, want one raw reference", turn.refLine)
	}
	if turn.input != "Review this" {
		t.Fatalf("input = %q, want visible user task", turn.input)
	}
	for _, bad := range []string{"README.md", ".reasonix/attachments/missing.txt", "../outside.txt"} {
		if _, err := server.prepareAttachmentTurn(req, "x", []string{bad}); err == nil {
			t.Errorf("prepareAttachmentTurn accepted %q", bad)
		}
	}
	_ = srv
}

func TestServeAttachmentOnlySubmissionSeparatesTranscriptAndModelText(t *testing.T) {
	_, srv, info, got, release := attachmentSubmitServer(t)
	body := `{"input":"","attachments":["` + info.Path + `","` + info.Path + `"]}`
	resp, err := http.Post(srv.URL+"/submit", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("submit status = %d, want 202", resp.StatusCode)
	}
	select {
	case input := <-got:
		if !strings.Contains(input, "trusted attachment body") || !strings.Contains(input, "Review the attached file or files") {
			t.Fatalf("model input = %q, want resolved attachment plus neutral task", input)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("attachment turn never reached runner")
	}
	release <- struct{}{}

	deadline := time.Now().Add(2 * time.Second)
	for {
		resp, err = http.Get(srv.URL + "/history")
		if err != nil {
			t.Fatal(err)
		}
		var history []historyMessage
		err = json.NewDecoder(resp.Body).Decode(&history)
		resp.Body.Close()
		if err == nil && len(history) > 0 {
			content := history[len(history)-1].Content
			if !strings.Contains(content, "@[notes.txt]("+info.Path+")") {
				t.Fatalf("history content = %q, want replayable display reference", content)
			}
			if strings.Contains(content, "Review the attached file or files") || strings.Contains(content, "trusted attachment body") {
				t.Fatalf("history leaked model-only text: %q", content)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("attachment message never appeared in history")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestServeAttachmentEditKeepsReplayableDisplayReference(t *testing.T) {
	ctrl, srv, info, got, release := attachmentSubmitServer(t)
	first := `{"input":"Original","attachments":["` + info.Path + `"]}`
	resp, err := http.Post(srv.URL+"/submit", "application/json", strings.NewReader(first))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	<-got
	release <- struct{}{}
	deadline := time.Now().Add(2 * time.Second)
	for ctrl.Running() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if ctrl.Running() {
		t.Fatal("initial attachment turn did not finish")
	}
	original := "Original\n\n@[notes.txt](" + info.Path + ")"
	editBody, err := json.Marshal(map[string]any{
		"input":       "Updated",
		"original":    original,
		"attachments": []string{info.Path},
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, err = http.Post(srv.URL+"/edit", "application/json", strings.NewReader(string(editBody)))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("edit status = %d, want 202", resp.StatusCode)
	}
	select {
	case input := <-got:
		if !strings.Contains(input, "Updated") || !strings.Contains(input, "trusted attachment body") {
			t.Fatalf("edited model input = %q", input)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("edited attachment turn never reached runner")
	}
	release <- struct{}{}
	deadline = time.Now().Add(2 * time.Second)
	for ctrl.Running() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	history := ctrl.History()
	if len(history) == 0 {
		t.Fatal("edited history is empty")
	}
	gotDisplay := agent.UserMessageText(history[len(history)-1])
	if !strings.Contains(gotDisplay, "Updated") || !strings.Contains(gotDisplay, "@[notes.txt]("+info.Path+")") {
		t.Fatalf("edited display = %q, want text plus replayable attachment", gotDisplay)
	}
	if strings.Contains(gotDisplay, "trusted attachment body") {
		t.Fatalf("edited display leaked resolved file content: %q", gotDisplay)
	}
}

func TestServeRejectsAttachmentCommandGoalRecoveryAndSteerMixes(t *testing.T) {
	ctrl, srv, info, _, _ := attachmentSubmitServer(t)
	cases := []struct {
		path string
		body string
	}{
		{"/submit", `{"input":"/model x","attachments":["` + info.Path + `"]}`},
		{"/submit", `{"input":"review","action":"approve_plan","attachments":["` + info.Path + `"]}`},
		{"/submit", `{"input":"review","attachments":[".reasonix/attachments/missing.txt"]}`},
		{"/delivery-recovery", `{"input":"continue","attachments":["` + info.Path + `"]}`},
		{"/steer", `{"text":"look closer","attachments":["` + info.Path + `"]}`},
		{"/goal", `{"goal":"ship it","attachments":["` + info.Path + `"]}`},
	}
	for _, tc := range cases {
		resp, err := http.Post(srv.URL+tc.path, "application/json", strings.NewReader(tc.body))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode < 400 || resp.StatusCode >= 500 {
			t.Errorf("POST %s status = %d, want 4xx", tc.path, resp.StatusCode)
		}
	}
	ctrl.SetGoal("active goal")
	resp, err := http.Post(srv.URL+"/submit", "application/json", strings.NewReader(`{"input":"review","attachments":["`+info.Path+`"]}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("Goal attachment status = %d, want 409", resp.StatusCode)
	}
}

type imageStatusController struct{ control.SessionAPI }

func (imageStatusController) ImageInputEnabled() bool { return true }

func TestServeStatusReportsImageInputCapability(t *testing.T) {
	bc := NewBroadcaster()
	base := control.New(control.Options{Sink: bc})
	t.Cleanup(base.Close)
	recorder := httptest.NewRecorder()
	New(imageStatusController{SessionAPI: base}, bc, config.ServeConfig{}).Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/status", nil))
	var body map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["imageInputEnabled"] != true {
		t.Fatalf("imageInputEnabled = %v, want true", body["imageInputEnabled"])
	}
}
