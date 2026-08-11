package serve

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"reasonix/internal/config"
	"reasonix/internal/control"
)

func TestServeIndexPageAndSessionDeepLink(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	for _, path := range []string{"/", "/sessions/reserved-session"} {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s status = %d", path, resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
			t.Errorf("GET %s content-type = %q, want text/html", path, ct)
		}
	}
}

func TestServeWebPagesBootstrapFragmentTokenBeforeRequests(t *testing.T) {
	for name, html := range map[string]string{
		"index":          string(baizeJS),
		"provider setup": string(providerSetupHTML),
	} {
		t.Run(name, func(t *testing.T) {
			for _, want := range []string{
				"new URLSearchParams(window.location.hash.slice(1))",
				"'/auth/token'",
				"window.history.replaceState",
				"window.fetch",
			} {
				if !strings.Contains(html, want) {
					t.Fatalf("page missing fragment-token bootstrap %q", want)
				}
			}
		})
	}
	if !strings.Contains(string(baizeJS), "__authReady.then(connectEvents)") {
		t.Fatal("serve index must delay SSE until fragment authentication completes")
	}
}

func TestServeIndexLoadsExternalBaizeAssets(t *testing.T) {
	html := string(indexHTML)
	for _, want := range []string{
		`data-language="__LANG__"`,
		`<link rel="stylesheet" href="/assets/baize.css" />`,
		`<script src="/assets/vendor.min.js"></script>`,
		`<script src="/assets/baize.js"></script>`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("serve index missing external asset contract %q", want)
		}
	}
	if strings.Contains(html, "const { marked, DOMPurify, hljs }") || strings.Contains(html, "*,*::before,*::after") {
		t.Fatal("serve index still contains the extracted Baize application source")
	}
	if strings.Index(html, "/assets/vendor.min.js") > strings.Index(html, "/assets/baize.js") {
		t.Fatal("Baize application script must load after its vendor bundle")
	}
}

func TestServeBaizeAssetRoutes(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	t.Cleanup(ctrl.Close)
	handler := New(ctrl, bc, config.ServeConfig{}).Handler()

	for _, tc := range []struct {
		path        string
		contentType string
		body        string
	}{
		{path: "/assets/baize.css", contentType: "text/css; charset=utf-8", body: ".card-main"},
		{path: "/assets/baize.js", contentType: "application/javascript; charset=utf-8", body: "const __LANG_PREF"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want 200", tc.path, recorder.Code)
			}
			if got := recorder.Header().Get("Content-Type"); got != tc.contentType {
				t.Fatalf("GET %s Content-Type = %q, want %q", tc.path, got, tc.contentType)
			}
			if got := recorder.Header().Get("Cache-Control"); got != "no-cache" {
				t.Fatalf("GET %s Cache-Control = %q, want no-cache", tc.path, got)
			}
			if !strings.Contains(recorder.Body.String(), tc.body) {
				t.Fatalf("GET %s missing source marker %q", tc.path, tc.body)
			}
		})
	}
}
