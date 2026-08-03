package serve

import (
	"encoding/base64"
	"encoding/json"
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
