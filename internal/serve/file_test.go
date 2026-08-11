package serve

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"reasonix/internal/config"
	"reasonix/internal/control"
)

// testCtrlWithWorkspace builds a controller whose session lives in a temp
// workspace so /file and /attach have a root to confine against.
func testCtrlWithWorkspace(t *testing.T) (*control.Controller, string) {
	t.Helper()
	home := t.TempDir()
	ws := filepath.Join(home, "ws")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatal(err)
	}
	sessDir := filepath.Join(home, "sessions")
	if err := os.MkdirAll(sessDir, 0o755); err != nil {
		t.Fatal(err)
	}
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{
		Sink:          bc,
		SessionDir:    sessDir,
		WorkspaceRoot: ws,
	})
	// WorkspaceRoot must return the configured workspace.
	if got := ctrl.WorkspaceRoot(); got != ws {
		t.Fatalf("WorkspaceRoot() = %q, want %q", got, ws)
	}
	return ctrl, ws
}

func TestServeFileServesWorkspaceImages(t *testing.T) {
	ctrl, ws := testCtrlWithWorkspace(t)
	srv := httptest.NewServer(New(ctrl, NewBroadcaster(), config.ServeConfig{}).Handler())
	defer srv.Close()

	// PNG magic bytes in a file inside the workspace.
	png := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 1, 2, 3}
	if err := os.WriteFile(filepath.Join(ws, "pic.png"), png, 0o644); err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(srv.URL + "/file?path=pic.png")
	if err != nil {
		t.Fatal(err)
	}
	body := make([]byte, len(png))
	n, _ := resp.Body.Read(body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("file status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", ct)
	}
	if n != len(png) || body[1] != 'P' {
		t.Errorf("body mismatch: %q", body[:n])
	}

	// Absolute path inside workspace works too.
	if resp, _ := http.Get(srv.URL + "/file?path=" + filepath.ToSlash(filepath.Join(ws, "pic.png"))); resp.StatusCode != 200 {
		t.Errorf("absolute path status = %d, want 200", resp.StatusCode)
	}
}

func TestServeFileRejectsEscapes(t *testing.T) {
	ctrl, ws := testCtrlWithWorkspace(t)
	srv := httptest.NewServer(New(ctrl, NewBroadcaster(), config.ServeConfig{}).Handler())
	defer srv.Close()

	// A file outside the workspace (sibling dir).
	outside := filepath.Dir(ws)
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("s3cret"), 0o644); err != nil {
		t.Fatal(err)
	}

	for name, path := range map[string]string{
		"traversal":   "../secret.txt",
		"abs outside": filepath.ToSlash(filepath.Join(outside, "secret.txt")),
		"missing":     "nope.png",
		"no path":     "",
	} {
		resp, err := http.Get(srv.URL + "/file?path=" + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s: status = %d, want 400/404", name, resp.StatusCode)
		}
	}

	// Non-image types are refused.
	if err := os.WriteFile(filepath.Join(ws, "doc.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if resp, _ := http.Get(srv.URL + "/file?path=doc.txt"); resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("txt status = %d, want 415", resp.StatusCode)
	}
}

func TestServeFileRejectsSymlinkEscape(t *testing.T) {
	ctrl, ws := testCtrlWithWorkspace(t)
	srv := httptest.NewServer(New(ctrl, NewBroadcaster(), config.ServeConfig{}).Handler())
	defer srv.Close()

	outside := filepath.Dir(ws)
	target := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(target, []byte("s3cret"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(ws, "link.png")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	resp, err := http.Get(srv.URL + "/file?path=link.png")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		t.Errorf("symlink escape status = %d, want 400/404", resp.StatusCode)
	}
}

func TestServeAttachSavesBase64Image(t *testing.T) {
	ctrl, ws := testCtrlWithWorkspace(t)
	srv := httptest.NewServer(New(ctrl, NewBroadcaster(), config.ServeConfig{}).Handler())
	defer srv.Close()

	b64 := base64.StdEncoding.EncodeToString([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a})
	resp, err := http.Post(srv.URL+"/attach", "application/json",
		strings.NewReader(`{"name":"../evil.png","data":"`+b64+`"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("attach status = %d, want 200", resp.StatusCode)
	}
	var out struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Path == "" || strings.Contains(out.Path, "..") {
		t.Fatalf("bad returned path: %q", out.Path)
	}
	abs := filepath.Join(ws, filepath.FromSlash(out.Path))
	if _, err := os.Stat(abs); err != nil {
		t.Fatalf("saved file missing: %v", err)
	}
	if !strings.HasPrefix(filepath.ToSlash(out.Path), ".reasonix/attachments/") {
		t.Errorf("attachment path %q should live under .reasonix/attachments/", out.Path)
	}

	// The saved file is now servable through /file.
	if resp, _ := http.Get(srv.URL + "/file?path=" + out.Path); resp.StatusCode != 200 {
		t.Errorf("attached file not servable, status = %d", resp.StatusCode)
	}

	// Rejects garbage.
	resp, err = http.Post(srv.URL+"/attach", "application/json", strings.NewReader(`{"name":"x.png","data":"@@not-base64@@base64::"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("bad base64 status = %d, want 400", resp.StatusCode)
	}
	if resp, _ := http.Post(srv.URL+"/attach", "application/json", strings.NewReader(`{"name":"x.png"}`)); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("missing data status = %d, want 400", resp.StatusCode)
	}
}

// TestServeEffortEndpoints checks the effort capability shape and validation.
func TestServeEffortEndpoints(t *testing.T) {
	ctrl, _ := testCtrlWithWorkspace(t)
	srv := httptest.NewServer(New(ctrl, NewBroadcaster(), config.ServeConfig{}).Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/effort")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /effort status = %d, want 200", resp.StatusCode)
	}
	var body struct {
		Supported bool     `json:"supported"`
		Levels    []string `json:"levels"`
		Current   string   `json:"current"`
		Default   string   `json:"default"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode /effort: %v", err)
	}
	if body.Supported && len(body.Levels) == 0 {
		t.Error("supported=true but no levels")
	}
	if body.Supported && body.Current == "" {
		t.Error("supported=true but empty current")
	}

	bad, err := http.Post(srv.URL+"/effort", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	bad.Body.Close()
	if bad.StatusCode != http.StatusBadRequest {
		t.Errorf("POST /effort {} status = %d, want 400", bad.StatusCode)
	}
}

// TestServeDefaultToolApprovalMode checks desktop parity: a freshly built
// serve controller defaults to the config desktop.default_tool_approval_mode
// (auto under the default config), and an explicit ask config keeps ask.
func TestServeDefaultToolApprovalMode(t *testing.T) {
	// The serve CLI entrypoint applies the configured desktop default; New()
	// itself must not touch the controller posture (tests and embedded use
	// construct their own policy). Isolate via REASONIX_HOME so the operator's
	// machine config cannot leak in.
	t.Setenv("REASONIX_HOME", t.TempDir())

	mode := func(t *testing.T) string {
		t.Helper()
		ctrl, _ := testCtrlWithWorkspace(t)
		ApplyDesktopDefaultApprovalMode(ctrl)
		return ctrl.ToolApprovalMode()
	}

	// No explicit desktop default → the config default (auto, desktop parity).
	if got := mode(t); got != "auto" {
		t.Fatalf("default mode = %q, want auto", got)
	}

	// Explicit auto → auto.
	editPath := config.UserConfigPath()
	if editPath == "" {
		t.Skip("no user config path in this test environment")
	}
	edit := config.LoadForEdit(editPath)
	if err := edit.SetDesktopDefaultToolApprovalMode("auto"); err != nil {
		t.Fatal(err)
	}
	if err := edit.SaveTo(editPath); err != nil {
		t.Fatal(err)
	}
	if got := mode(t); got != "auto" {
		t.Fatalf("configured auto mode = %q, want auto", got)
	}

	// Explicit ask → ask.
	edit2 := config.LoadForEdit(editPath)
	if err := edit2.SetDesktopDefaultToolApprovalMode("ask"); err != nil {
		t.Fatal(err)
	}
	if err := edit2.SaveTo(editPath); err != nil {
		t.Fatal(err)
	}
	if got := mode(t); got != "ask" {
		t.Fatalf("configured ask mode = %q, want ask", got)
	}
}

// TestServeSubmitEffortBare checks that a bare /effort submit is intercepted
// (reporting the current effort capability) instead of falling through to the
// controller's "unknown command" notice.
func TestServeSubmitEffortBare(t *testing.T) {
	ctrl, _ := testCtrlWithWorkspace(t)
	srv := httptest.NewServer(New(ctrl, NewBroadcaster(), config.ServeConfig{}).Handler())
	defer srv.Close()

	// Bare command (with and without trailing whitespace) must return the
	// same capability payload as GET /effort — 200 JSON, not unknown command.
	for _, input := range []string{`{"input":"/effort"}`, `{"input":"/effort "}`} {
		resp, err := http.Post(srv.URL+"/submit", "application/json", strings.NewReader(input))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			t.Fatalf("POST /submit %s status = %d, want 200 (%s)", input, resp.StatusCode, body)
		}
		var body struct {
			Supported bool     `json:"supported"`
			Levels    []string `json:"levels"`
			Current   string   `json:"current"`
			Default   string   `json:"default"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			resp.Body.Close()
			t.Fatalf("decode /submit %s: %v", input, err)
		}
		resp.Body.Close()
		if body.Supported && (len(body.Levels) == 0 || body.Current == "") {
			t.Errorf("bare /effort payload inconsistent: %+v", body)
		}
	}

	// An invalid level must not fall through to the controller either — the
	// switch rejects it before any turn is submitted.
	bad, err := http.Post(srv.URL+"/submit", "application/json", strings.NewReader(`{"input":"/effort bogus"}`))
	if err != nil {
		t.Fatal(err)
	}
	bad.Body.Close()
	if bad.StatusCode != http.StatusInternalServerError {
		t.Errorf("POST /submit /effort bogus status = %d, want 500", bad.StatusCode)
	}
}

// TestServeProfileEndpoints checks the work-mode (runtime profile) endpoints.
func TestServeProfileEndpoints(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{
		Sink:       bc,
		Label:      "m",
		SessionDir: t.TempDir(),
	})
	server := New(ctrl, bc, config.ServeConfig{})
	var rebuilt bool
	server.buildController = func(_ context.Context, _ string) (*control.Controller, error) {
		rebuilt = true
		return control.New(control.Options{
			Sink:       bc,
			Label:      "m",
			SessionDir: t.TempDir(),
		}), nil
	}
	srv := httptest.NewServer(server.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/profile")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["mode"] != "full" {
		t.Fatalf("initial mode = %v, want full", body["mode"])
	}

	// Missing / invalid mode are rejected.
	for _, payload := range []string{`{}`, `{"mode":"bogus"}`} {
		bad, err := http.Post(srv.URL+"/profile", "application/json", strings.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		bad.Body.Close()
		if bad.StatusCode != http.StatusBadRequest {
			t.Errorf("POST /profile %s status = %d, want 400", payload, bad.StatusCode)
		}
	}

	// Switching to economy rebuilds the controller and persists the mode.
	ok, err := http.Post(srv.URL+"/profile", "application/json", strings.NewReader(`{"mode":"economy"}`))
	if err != nil {
		t.Fatal(err)
	}
	ok.Body.Close()
	if ok.StatusCode != http.StatusNoContent {
		t.Fatalf("POST /profile economy status = %d, want 204", ok.StatusCode)
	}
	if !rebuilt {
		t.Error("work-mode switch did not rebuild the controller")
	}
	resp2, err := http.Get(srv.URL + "/profile")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if err := json.NewDecoder(resp2.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["mode"] != "economy" {
		t.Fatalf("mode after switch = %v, want economy", body["mode"])
	}
}

// TestSecurePathJoinNormalizesRootSymlinks locks the macOS /var →
// /private/var fix: t.TempDir() resolves under a symlinked prefix on macOS,
// so an unresolved root used to fail every containment check (400 on /file).
// On Windows/Linux the prefix is not a symlink and the test passes trivially.
func TestSecurePathJoinNormalizesRootSymlinks(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "ws")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	img := filepath.Join(root, "pic.png")
	if err := os.WriteFile(img, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Relative path inside the workspace must resolve.
	got, err := securePathJoin(root, "pic.png")
	if err != nil {
		t.Fatalf("relative path: %v", err)
	}
	if fi, err := os.Stat(got); err != nil || fi.IsDir() {
		t.Fatalf("resolved path %q not servable: %v", got, err)
	}

	// Absolute path (possibly unresolved form) must resolve too.
	if _, err := securePathJoin(root, img); err != nil {
		t.Fatalf("absolute path: %v", err)
	}

	// Escapes must still be rejected.
	outside := filepath.Join(base, "secret.txt")
	if err := os.WriteFile(outside, []byte("s"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := securePathJoin(root, outside); err == nil {
		t.Fatal("absolute path outside workspace was accepted")
	}
	if _, err := securePathJoin(root, filepath.Join(root, "..", "secret.txt")); err == nil {
		t.Fatal(".. escape was accepted")
	}
}
