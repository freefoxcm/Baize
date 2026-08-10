package serve

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"reasonix/internal/agent"
	"reasonix/internal/config"
	"reasonix/internal/control"
	"reasonix/internal/event"
	"reasonix/internal/eventwire"
	"reasonix/internal/provider"
	"reasonix/internal/store"
)

func makeApprovalTestSession(t *testing.T, path string) {
	t.Helper()
	sess := agent.NewSession("sys")
	sess.Add(provider.Message{Role: provider.RoleUser, Content: "hello"})
	if err := sess.Save(path); err != nil {
		t.Fatal(err)
	}
}

func newApprovalTestServer(t *testing.T, dir, path string) (*httptest.Server, *control.Controller) {
	t.Helper()
	makeApprovalTestSession(t, path)
	exec := agent.New(nil, nil, agent.NewSession("sys"), agent.Options{}, event.Discard)
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Executor: exec, SessionDir: dir, SessionPath: path, Label: "test"})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	t.Cleanup(srv.Close)
	return srv, ctrl
}

func TestApprovalModeSidecarRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "s1.jsonl")
	if got := readSessionApprovalMode(p); got != "" {
		t.Fatalf("missing sidecar read = %q, want empty", got)
	}
	if err := writeSessionApprovalMode(p, "auto"); err != nil {
		t.Fatal(err)
	}
	if got := readSessionApprovalMode(p); got != "auto" {
		t.Fatalf("read after write = %q, want auto", got)
	}
	if err := writeSessionApprovalMode(p, "yolo"); err != nil {
		t.Fatal(err)
	}
	if got := readSessionApprovalMode(p); got != "yolo" {
		t.Fatalf("read after rewrite = %q, want yolo", got)
	}
	if err := os.WriteFile(store.SessionApprovalMode(p), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := readSessionApprovalMode(p); got != "" {
		t.Fatalf("corrupt sidecar read = %q, want empty", got)
	}
}

func TestToolApprovalModePersistsToSession(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s1.jsonl")
	srv, ctrl := newApprovalTestServer(t, dir, path)

	resp, err := http.Post(srv.URL+"/tool-approval-mode", "application/json", strings.NewReader(`{"mode":"yolo"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	if got := readSessionApprovalMode(path); got != "yolo" {
		t.Fatalf("sidecar = %q, want yolo", got)
	}
	if got := ctrl.ToolApprovalMode(); got != control.ToolApprovalYolo {
		t.Fatalf("controller mode = %q, want yolo", got)
	}
}

func TestNewSessionResetsApprovalModeToDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s1.jsonl")
	srv, ctrl := newApprovalTestServer(t, dir, path)
	if _, err := http.Post(srv.URL+"/tool-approval-mode", "application/json", strings.NewReader(`{"mode":"yolo"}`)); err != nil {
		t.Fatal(err)
	}
	want := defaultApprovalMode()

	resp, err := http.Post(srv.URL+"/new", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	if got := ctrl.ToolApprovalMode(); got != want {
		t.Fatalf("controller mode after /new = %q, want default %q", got, want)
	}
}

func TestResumeRestoresSessionApprovalMode(t *testing.T) {
	dir := t.TempDir()
	pathA := filepath.Join(dir, "a.jsonl")
	pathB := filepath.Join(dir, "b.jsonl")
	srv, ctrl := newApprovalTestServer(t, dir, pathA)
	makeApprovalTestSession(t, pathB)
	if err := writeSessionApprovalMode(pathA, "yolo"); err != nil {
		t.Fatal(err)
	}
	if err := writeSessionApprovalMode(pathB, "ask"); err != nil {
		t.Fatal(err)
	}
	// Apply A's posture so the controller starts where a resumed A would be.
	ctrl.SetToolApprovalMode("yolo")

	for _, tc := range []struct {
		target string
		want   string
	}{
		{pathB, "ask"},
		{pathA, "yolo"},
		{pathB, "ask"},
	} {
		resp, err := http.Post(srv.URL+"/resume", "application/json", strings.NewReader(`{"path":`+strconv.Quote(tc.target)+`}`))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNoContent {
			body := respBody(resp)
			t.Fatalf("resume %s status = %d, want 204 (%s)", tc.target, resp.StatusCode, body)
		}
		resp.Body.Close()
		if got := ctrl.ToolApprovalMode(); got != tc.want {
			t.Fatalf("mode after resume %s = %q, want %q", tc.target, got, tc.want)
		}
	}
}

func TestSwitchRestoresSessionApprovalMode(t *testing.T) {
	dir := t.TempDir()
	pathA := filepath.Join(dir, "a.jsonl")
	pathB := filepath.Join(dir, "b.jsonl")
	srv, ctrl := newApprovalTestServer(t, dir, pathA)
	makeApprovalTestSession(t, pathB)
	if err := writeSessionApprovalMode(pathA, "yolo"); err != nil {
		t.Fatal(err)
	}
	if err := writeSessionApprovalMode(pathB, "auto"); err != nil {
		t.Fatal(err)
	}
	ctrl.SetToolApprovalMode("yolo")

	resp, err := http.Post(srv.URL+"/submit", "application/json", strings.NewReader(`{"input":"/switch `+agent.BranchID(pathB)+`"}`))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("switch status = %d, want 204 (body: %s)", resp.StatusCode, respBody(resp))
	}
	resp.Body.Close()
	if got := ctrl.SessionPath(); got != pathB {
		t.Fatalf("controller session after /switch = %q, want %q", got, pathB)
	}
	if got := ctrl.ToolApprovalMode(); got != control.ToolApprovalAuto {
		t.Fatalf("mode after switch = %q, want auto", got)
	}
}

func TestForkInheritsSourceApprovalMode(t *testing.T) {
	dir := t.TempDir()
	pathA := filepath.Join(dir, "a.jsonl")
	makeApprovalTestSession(t, pathA)
	bc := NewBroadcaster()
	got := make(chan string, 1)
	ctrl := control.New(control.Options{
		Executor:   agent.New(nil, nil, agent.NewSession("sys"), agent.Options{}, event.Discard),
		SessionDir: dir, SessionPath: pathA, Label: "test",
		Runner: fakeRunner{got: got},
		Sink:   bc,
	})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()
	sub, cancel := bc.Subscribe()
	defer cancel()

	// A real turn seeds the checkpoint boundary fork needs.
	if resp, err := http.Post(srv.URL+"/submit", "application/json", strings.NewReader(`{"input":"hi"}`)); err != nil {
		t.Fatal(err)
	} else {
		resp.Body.Close()
	}
	select {
	case <-got:
	case <-time.After(2 * time.Second):
		t.Fatal("runner never ran")
	}
	deadline := time.After(5 * time.Second)
	for {
		select {
		case data := <-sub:
			var w eventwire.Event
			if err := json.Unmarshal(data, &w); err == nil && w.Kind == "turn_done" {
				goto turned
			}
		case <-deadline:
			t.Fatal("never saw turn_done on the stream")
		}
	}
turned:
	if _, err := http.Post(srv.URL+"/tool-approval-mode", "application/json", strings.NewReader(`{"mode":"yolo"}`)); err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(srv.URL+"/fork", "application/json", strings.NewReader(`{"turn":0}`))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		body := respBody(resp)
		t.Fatalf("fork status = %d, want 200 (%s)", resp.StatusCode, body)
	}
	resp.Body.Close()
	if got := ctrl.ToolApprovalMode(); got != control.ToolApprovalYolo {
		t.Fatalf("controller mode after fork = %q, want yolo", got)
	}
	p := ctrl.SessionPath()
	if p == pathA {
		t.Fatal("fork did not switch to a new session path")
	}
	if got := readSessionApprovalMode(p); got != control.ToolApprovalYolo {
		t.Fatalf("fork sidecar mode = %q, want yolo", got)
	}
}

func TestSwitchUnknownRefFails(t *testing.T) {
	dir := t.TempDir()
	pathA := filepath.Join(dir, "a.jsonl")
	srv, _ := newApprovalTestServer(t, dir, pathA)

	resp, err := http.Post(srv.URL+"/submit", "application/json", strings.NewReader(`{"input":"/switch does-not-exist-xyz"}`))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("switch unknown ref status = %d, want 500", resp.StatusCode)
	}
	resp.Body.Close()
}

func respBody(resp *http.Response) string {
	if resp == nil || resp.Body == nil {
		return ""
	}
	defer resp.Body.Close()
	var sb strings.Builder
	buf := make([]byte, 512)
	for {
		n, err := resp.Body.Read(buf)
		sb.Write(buf[:n])
		if err != nil {
			break
		}
	}
	return sb.String()
}
