package serve

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"reasonix/internal/agent"
	"reasonix/internal/config"
	"reasonix/internal/control"
	"reasonix/internal/event"
)

func postQualityFloor(t *testing.T, url, floor string) *http.Response {
	t.Helper()
	resp, err := http.Post(url+"/quality-floor", "application/json", strings.NewReader(`{"floor":`+strconv.Quote(floor)+`}`))
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestQualityFloorEndpointPersistsAndStatusReports(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "quality.jsonl")
	srv, ctrl := newApprovalTestServer(t, dir, path)

	resp := postQualityFloor(t, srv.URL, control.QualityFloorDelivery)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("POST status = %d, want 204", resp.StatusCode)
	}
	if got := ctrl.QualityFloor(); got != control.QualityFloorDelivery {
		t.Fatalf("controller floor = %q, want delivery", got)
	}
	meta, ok, err := agent.LoadBranchMeta(path)
	if err != nil || !ok || meta.QualityFloor != control.QualityFloorDelivery {
		t.Fatalf("persisted meta = %+v, ok=%v err=%v", meta, ok, err)
	}

	statusResp, err := http.Get(srv.URL + "/status")
	if err != nil {
		t.Fatal(err)
	}
	defer statusResp.Body.Close()
	var status struct {
		QualityFloor string `json:"qualityFloor"`
	}
	if err := json.NewDecoder(statusResp.Body).Decode(&status); err != nil {
		t.Fatal(err)
	}
	if status.QualityFloor != control.QualityFloorDelivery {
		t.Fatalf("status qualityFloor = %q, want delivery", status.QualityFloor)
	}

	resp = postQualityFloor(t, srv.URL, control.QualityFloorStandard)
	resp.Body.Close()
	meta, ok, err = agent.LoadBranchMeta(path)
	if err != nil || !ok || meta.QualityFloor != "" {
		t.Fatalf("standard meta = %+v, ok=%v err=%v", meta, ok, err)
	}
}

func TestQualityFloorEndpointRejectsInvalidAndRunningSwitch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "quality.jsonl")
	srv, _ := newApprovalTestServer(t, dir, path)

	resp := postQualityFloor(t, srv.URL, "turbo")
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid floor status = %d, want 400", resp.StatusCode)
	}

	bc := NewBroadcaster()
	runningCtrl := control.New(control.Options{Sink: bc, Runner: blockingRunner{}, SessionDir: t.TempDir()})
	runningServer := httptest.NewServer(New(runningCtrl, bc, config.ServeConfig{}).Handler())
	t.Cleanup(runningServer.Close)
	runningCtrl.Submit("keep the turn active")
	waitRunning(t, runningCtrl)
	resp = postQualityFloor(t, runningServer.URL, control.QualityFloorDelivery)
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("running floor status = %d, want 409", resp.StatusCode)
	}
	runningCtrl.Cancel()
}

func TestResumeRestoresSessionQualityFloor(t *testing.T) {
	dir := t.TempDir()
	pathA := filepath.Join(dir, "a.jsonl")
	pathB := filepath.Join(dir, "b.jsonl")
	srv, ctrl := newApprovalTestServer(t, dir, pathA)
	makeApprovalTestSession(t, pathB)
	if err := agent.UpdateBranchMeta(pathB, false, func(meta *agent.BranchMeta) error {
		meta.QualityFloor = control.QualityFloorDelivery
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(srv.URL+"/resume", "application/json", strings.NewReader(`{"path":`+strconv.Quote(pathB)+`}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("resume status = %d, want 204", resp.StatusCode)
	}
	if got := ctrl.QualityFloor(); got != control.QualityFloorDelivery {
		t.Fatalf("resumed floor = %q, want delivery", got)
	}
}

func TestNewSessionInheritsQualityFloor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "source.jsonl")
	srv, ctrl := newApprovalTestServer(t, dir, path)
	resp := postQualityFloor(t, srv.URL, control.QualityFloorDelivery)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("set floor status = %d", resp.StatusCode)
	}

	resp, err := http.Post(srv.URL+"/new", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("new status = %d, want 204", resp.StatusCode)
	}
	if got := ctrl.QualityFloor(); got != control.QualityFloorDelivery {
		t.Fatalf("new session floor = %q, want delivery", got)
	}
	if got := qualityFloorFromMeta(ctrl.SessionPath()); got != control.QualityFloorDelivery {
		t.Fatalf("new session persisted floor = %q, want delivery", got)
	}
}

func TestSwitchModelPreservesQualityFloor(t *testing.T) {
	bc := NewBroadcaster()
	old := control.New(control.Options{
		Executor: agent.New(nil, nil, agent.NewSession("sys"), agent.Options{}, event.Discard),
		Sink:     bc,
		Label:    "old",
	})
	if err := old.SetQualityFloor(control.QualityFloorDelivery); err != nil {
		t.Fatal(err)
	}
	s := &Server{ctrl: old, bc: bc}
	var built *control.Controller
	s.buildController = func(_ context.Context, _ string) (*control.Controller, error) {
		built = control.New(control.Options{Sink: bc, Label: "new"})
		return built, nil
	}

	if err := s.switchModel(context.Background(), "next-model"); err != nil {
		t.Fatalf("switchModel: %v", err)
	}
	defer built.Close()
	if got := s.ctl().QualityFloor(); got != control.QualityFloorDelivery {
		t.Fatalf("quality floor after switch = %q, want delivery", got)
	}
}
