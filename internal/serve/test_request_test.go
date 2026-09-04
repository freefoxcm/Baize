package serve

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"reasonix/internal/config"
)

// localTestRequest gives handler-level tests the loopback Host they would
// receive from an actual Serve listener. httptest.NewRequest defaults to
// example.com, which the DNS-rebinding guard intentionally rejects.
func localTestRequest(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.Host = "127.0.0.1"
	return req
}

// isolateServeConfig prevents tests of in-place session rotation from reading
// the developer's legacy user config (which may intentionally select another
// default model and therefore exercise the replacement-controller path).
func isolateServeConfig(t *testing.T) {
	t.Helper()
	isolateServeHome(t, t.TempDir())
	path := config.UserConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("default_model = \"test\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}
