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

func TestServePDFJSAssetRoutes(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	t.Cleanup(ctrl.Close)
	handler := New(ctrl, bc, config.ServeConfig{}).Handler()

	for _, tc := range []struct {
		path        string
		contentType string
		body        string
	}{
		{path: "/assets/pdfjs/pdf.mjs", contentType: "text/javascript; charset=utf-8", body: "Mozilla Foundation"},
		{path: "/assets/pdfjs/pdf.worker.mjs", contentType: "text/javascript; charset=utf-8", body: "WorkerMessageHandler"},
		{path: "/assets/pdfjs/wasm/openjpeg.wasm", contentType: "application/wasm", body: ""},
		{path: "/assets/pdfjs/cmaps/Adobe-GB1-UCS2.bcmap", contentType: "application/octet-stream", body: ""},
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
			if got := recorder.Header().Get("Cache-Control"); got != "public, max-age=86400" {
				t.Fatalf("GET %s Cache-Control = %q", tc.path, got)
			}
			if tc.body != "" && !strings.Contains(recorder.Body.String(), tc.body) {
				t.Fatalf("GET %s missing source marker %q", tc.path, tc.body)
			}
		})
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/assets/pdfjs/missing.mjs", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("missing PDF.js asset status = %d, want 404", recorder.Code)
	}
}

func TestWorkspacePDFPreviewUsesPDFJS(t *testing.T) {
	js := string(baizeJS)
	for _, want := range []string{
		"import('/assets/pdfjs/pdf.mjs')",
		"pdf.worker.mjs",
		"renderWorkspacePDF(preview)",
		"workspace-pdf__pages",
		"updateWorkspacePDFPageFromScroll",
		"Math.abs(view.page-state.page)<=1",
		"releaseWorkspacePDFView(view)",
		"isEvalSupported:false",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("Baize PDF preview missing %q", want)
		}
	}
	if strings.Contains(js, "if(preview.kind==='pdf'){const frame=el('iframe'") {
		t.Fatal("PDF preview still delegates to the browser's iframe PDF viewer")
	}
	css := string(baizeCSS)
	for _, want := range []string{
		`.workspace-pdf__viewport{min-width:0;min-height:0;position:relative;flex:1;overflow:auto;`,
		`.workspace-pdf__pages{width:max-content;min-width:100%;min-height:100%;display:flex;flex-direction:column;`,
	} {
		if !strings.Contains(css, want) {
			t.Errorf("Baize PDF preview stylesheet missing %q", want)
		}
	}
}

func TestSessionListKeepsScrollableFixedRows(t *testing.T) {
	css := string(baizeCSS)
	for _, want := range []string{
		`.session-list{flex:1;min-height:0;overflow-y:auto;`,
		`.session-item{flex:0 0 auto;display:flex;`,
	} {
		if !strings.Contains(css, want) {
			t.Errorf("Baize stylesheet missing session list layout contract %q", want)
		}
	}
}

func TestReasoningSummaryAlwaysStartsBelowHeader(t *testing.T) {
	css := string(baizeCSS)
	const want = `.reasoning__summary{display:block;width:calc(100% - 8px);box-sizing:border-box;`
	if !strings.Contains(css, want) {
		t.Fatalf("Baize stylesheet missing fixed reasoning summary row %q", want)
	}
}

func TestBaizeSessionSubagentAndSettingsUIContracts(t *testing.T) {
	js := string(baizeJS)
	for _, want := range []string{
		"sessionsLoadSequence",
		"if(sequence!==sessionsLoadSequence)return",
		"SUBAGENT_PROGRESS_LIMITS={reasoning:8<<10,text:8<<10,notice:2<<10}",
		"reasonix.subagent.status",
		"subagentAutoCollapse()",
		"fetch('/settings'",
		"method:'PATCH'",
		"baize-theme-preference",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("Baize application missing %q", want)
		}
	}
	html := string(indexHTML)
	for _, want := range []string{`id="btn-settings"`, `id="settings-drawer"`, `name="defaultModel"`, `name="maxSubagentConcurrency"`} {
		if !strings.Contains(html, want) {
			t.Errorf("Baize settings UI missing %q", want)
		}
	}
}
