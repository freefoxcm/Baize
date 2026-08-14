package serve

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	"reasonix/internal/config"
	fileenc "reasonix/internal/fileutil/encoding"
)

func workspaceTestServer(t *testing.T, cfg config.ServeConfig) (*httptest.Server, string) {
	t.Helper()
	ctrl, root := testCtrlWithWorkspace(t)
	t.Cleanup(ctrl.Close)
	return httptest.NewServer(New(ctrl, NewBroadcaster(), cfg).Handler()), root
}

func writeWorkspaceTestFile(t *testing.T, root, name string, data []byte) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestWorkspaceEntriesSortsAndProtectsFiles(t *testing.T) {
	srv, root := workspaceTestServer(t, config.ServeConfig{})
	defer srv.Close()
	writeWorkspaceTestFile(t, root, "z.txt", []byte("z"))
	writeWorkspaceTestFile(t, root, "A.txt", []byte("a"))
	writeWorkspaceTestFile(t, root, ".env", []byte("SECRET=x"))
	writeWorkspaceTestFile(t, root, ".env.example", []byte("SECRET="))
	writeWorkspaceTestFile(t, root, ".mcp.json", []byte("{}"))
	writeWorkspaceTestFile(t, root, "folder/nested.txt", []byte("nested"))
	if err := os.MkdirAll(filepath.Join(root, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(srv.URL + "/workspace/entries")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var result struct {
		Path    string           `json:"path"`
		Entries []workspaceEntry `json:"entries"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Path != "" {
		t.Fatalf("root path = %q, want empty", result.Path)
	}
	got := make([]string, 0, len(result.Entries))
	for _, entry := range result.Entries {
		got = append(got, entry.Path)
	}
	if strings.Join(got, ",") != "folder,.env.example,A.txt,z.txt" {
		t.Fatalf("entries = %v", got)
	}
	if !result.Entries[0].IsDir || result.Entries[2].Size != 1 || result.Entries[2].ModifiedAt == "" {
		t.Fatalf("entry metadata incomplete: %+v", result.Entries)
	}
}

func TestWorkspaceSearchFiltersProtectedFiles(t *testing.T) {
	srv, root := workspaceTestServer(t, config.ServeConfig{})
	defer srv.Close()
	writeWorkspaceTestFile(t, root, "reports/report.html", []byte("ok"))
	writeWorkspaceTestFile(t, root, "reports/report.key", []byte("secret"))

	resp, err := http.Get(srv.URL + "/workspace/search?q=report")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var result struct {
		Entries []workspaceEntry `json:"entries"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	foundReport := false
	for _, entry := range result.Entries {
		if entry.Path == "reports/report.key" {
			t.Fatalf("protected search entry leaked: %+v", result.Entries)
		}
		foundReport = foundReport || entry.Path == "reports/report.html"
	}
	if !foundReport {
		t.Fatalf("report.html missing from search entries: %+v", result.Entries)
	}
	if bad, _ := http.Get(srv.URL + "/workspace/search?q=x"); bad.StatusCode != http.StatusBadRequest {
		bad.Body.Close()
		t.Fatalf("short query status = %d, want 400", bad.StatusCode)
	}
}

func TestWorkspacePreviewKindsAndEncodings(t *testing.T) {
	srv, root := workspaceTestServer(t, config.ServeConfig{})
	defer srv.Close()
	writeWorkspaceTestFile(t, root, "README.md", []byte("# 标题"))
	writeWorkspaceTestFile(t, root, "report.html", []byte("<!doctype html><h1>报告</h1>"))
	writeWorkspaceTestFile(t, root, "gb.txt", fileenc.Encode("警务数据", fileenc.GB18030))
	utf16Bytes := []byte{0xff, 0xfe}
	for _, unit := range utf16.Encode([]rune("笔录分析")) {
		utf16Bytes = append(utf16Bytes, byte(unit), byte(unit>>8))
	}
	writeWorkspaceTestFile(t, root, "utf16.txt", utf16Bytes)
	writeWorkspaceTestFile(t, root, "raw.bin", []byte{0, 1, 2, 3})
	writeWorkspaceTestFile(t, root, "chart.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`))

	for _, tc := range []struct {
		path string
		kind string
		body string
		url  bool
	}{
		{path: "README.md", kind: "markdown", body: "标题"},
		{path: "report.html", kind: "html", body: "报告"},
		{path: "gb.txt", kind: "text", body: "警务数据"},
		{path: "utf16.txt", kind: "text", body: "笔录分析"},
		{path: "raw.bin", kind: "binary"},
		{path: "chart.svg", kind: "image", url: true},
	} {
		t.Run(tc.path, func(t *testing.T) {
			resp, err := http.Get(srv.URL + "/workspace/preview?path=" + tc.path)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d", resp.StatusCode)
			}
			var preview workspacePreview
			if err := json.NewDecoder(resp.Body).Decode(&preview); err != nil {
				t.Fatal(err)
			}
			if preview.Kind != tc.kind || tc.body != "" && !strings.Contains(preview.Body, tc.body) {
				t.Fatalf("preview = %+v", preview)
			}
			if tc.url != (preview.ContentURL != "") {
				t.Fatalf("content URL = %q, want present=%v", preview.ContentURL, tc.url)
			}
		})
	}
}

func TestWorkspacePreviewTruncatesAtUTF8Boundary(t *testing.T) {
	srv, root := workspaceTestServer(t, config.ServeConfig{})
	defer srv.Close()
	data := bytesRepeat([]byte("界"), workspacePreviewLimit/3+2)
	writeWorkspaceTestFile(t, root, "large.txt", data)
	resp, err := http.Get(srv.URL + "/workspace/preview?path=large.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var preview workspacePreview
	if err := json.NewDecoder(resp.Body).Decode(&preview); err != nil {
		t.Fatal(err)
	}
	if !preview.Truncated || preview.Kind != "text" || strings.ContainsRune(preview.Body, '\ufffd') {
		t.Fatalf("bad truncated preview: kind=%s truncated=%v tail=%q", preview.Kind, preview.Truncated, preview.Body[len(preview.Body)-12:])
	}
}

func bytesRepeat(value []byte, count int) []byte {
	out := make([]byte, 0, len(value)*count)
	for range count {
		out = append(out, value...)
	}
	return out
}

func TestWorkspaceEndpointsRejectEscapesAndSecrets(t *testing.T) {
	srv, root := workspaceTestServer(t, config.ServeConfig{})
	defer srv.Close()
	writeWorkspaceTestFile(t, root, ".env", []byte("SECRET=x"))
	writeWorkspaceTestFile(t, filepath.Dir(root), "outside.txt", []byte("outside"))
	for _, path := range []string{".env", "../outside.txt", filepath.ToSlash(filepath.Join(filepath.Dir(root), "outside.txt"))} {
		resp, err := http.Get(srv.URL + "/workspace/preview?path=" + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("path %q status = %d, want 400/403", path, resp.StatusCode)
		}
	}
	link := filepath.Join(root, "secret.txt")
	if err := os.Symlink(filepath.Join(root, ".env"), link); err == nil {
		resp, err := http.Get(srv.URL + "/workspace/preview?path=secret.txt")
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("protected symlink status = %d, want 403", resp.StatusCode)
		}
	}
}

func TestProtectedWorkspacePathNormalizesWindowsAliases(t *testing.T) {
	for _, path := range []string{
		`.ENV`,
		`folder/id_rsa.`,
		`folder/.env:stream`,
		`folder/report.html:payload`,
		`C:\project\secret.PEM`,
	} {
		if !protectedWorkspacePath(path) {
			t.Errorf("protectedWorkspacePath(%q) = false", path)
		}
	}
	for _, path := range []string{`.env.example`, `C:\project\.env.example`, `reports/report.html`} {
		if protectedWorkspacePath(path) {
			t.Errorf("protectedWorkspacePath(%q) = true", path)
		}
	}
}

func TestWorkspaceContentHeadersAndRange(t *testing.T) {
	srv, root := workspaceTestServer(t, config.ServeConfig{})
	defer srv.Close()
	writeWorkspaceTestFile(t, root, "chart.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`))
	writeWorkspaceTestFile(t, root, "report.pdf", []byte("%PDF-1.7\n0123456789"))
	writeWorkspaceTestFile(t, root, "report.html", []byte("<h1>not executable</h1>"))

	svgResp, err := http.Get(srv.URL + "/workspace/content?path=chart.svg")
	if err != nil {
		t.Fatal(err)
	}
	svgResp.Body.Close()
	if svgResp.Header.Get("Content-Type") != "image/svg+xml" || svgResp.Header.Get("Cache-Control") != "no-store" || !strings.Contains(svgResp.Header.Get("Content-Security-Policy"), "script-src 'none'") {
		t.Fatalf("unsafe SVG headers: %+v", svgResp.Header)
	}
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/workspace/content?path=report.pdf", nil)
	req.Header.Set("Range", "bytes=0-3")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if response.StatusCode != http.StatusPartialContent || string(body) != "%PDF" {
		t.Fatalf("range response status=%d body=%q", response.StatusCode, body)
	}
	if htmlResp, _ := http.Get(srv.URL + "/workspace/content?path=report.html"); htmlResp.StatusCode != http.StatusUnsupportedMediaType {
		htmlResp.Body.Close()
		t.Fatalf("HTML content status = %d, want 415", htmlResp.StatusCode)
	}
}

func TestWorkspaceEndpointsUseServeAuthentication(t *testing.T) {
	srv, root := workspaceTestServer(t, config.ServeConfig{AuthMode: "token", Token: "workspace-token"})
	defer srv.Close()
	writeWorkspaceTestFile(t, root, "report.txt", []byte("ok"))
	unauthorized, err := http.Get(srv.URL + "/workspace/preview?path=report.txt")
	if err != nil {
		t.Fatal(err)
	}
	unauthorized.Body.Close()
	if unauthorized.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want 401", unauthorized.StatusCode)
	}
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/workspace/preview?path=report.txt", nil)
	req.AddCookie(&http.Cookie{Name: cookieToken, Value: "workspace-token"})
	authorized, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	authorized.Body.Close()
	if authorized.StatusCode != http.StatusOK {
		t.Fatalf("authorized status = %d, want 200", authorized.StatusCode)
	}
}

func TestWorkspaceWebUIContract(t *testing.T) {
	html := string(indexHTML)
	for _, marker := range []string{
		`id="btn-sidebar-collapse"`,
		`id="btn-workspace"`,
		`id="workspace-panel"`,
		`id="workspace-resizer" role="separator"`,
		`aria-valuemin="480" aria-valuemax="1200" aria-valuenow="720" tabindex="0"`,
		`id="workspace-html-toggle"`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("workspace UI missing %s", marker)
		}
	}
	css := string(baizeCSS)
	for _, marker := range []string{
		`.app--sidebar-collapsed`,
		`.app--workspace-open`,
		`.workspace-panel`,
		`.workspace-resize-guide`,
		`position:fixed;z-index:180`,
		`width:min(var(--workspace-width),80vw,1200px)`,
		`.decision-pending .footer`,
		`@media(max-width:768px)`,
	} {
		if !strings.Contains(css, marker) {
			t.Fatalf("workspace CSS missing %s", marker)
		}
	}
	js := string(baizeJS)
	for _, marker := range []string{
		`baize-sidebar-collapsed`,
		`baize-workspace-width`,
		`/workspace/entries?path=`,
		`/workspace/preview?path=`,
		`WHOLE_DOCUMENT:true`,
		`preview.kind==='pdf'`,
		`workspaceChartScriptURL`,
		`rendered.interactive?'allow-scripts':''`,
		`connect-src 'none'`,
		`requestAnimationFrame(paint)`,
		`fixWorkspaceLinks(root)`,
		`workspacePathCandidate(code.textContent,false)`,
		`openWorkspacePath(path)`,
		`if(error?.status===404)`,
		`WORKSPACE_WIDTH_DEFAULT=720`,
		`window.innerWidth*.8`,
		`workspaceResizer.addEventListener('keydown'`,
		`decisionInteractionLocked`,
		`closeSettings({preserveDraft:true,restoreFocus:false})`,
		`usageSelectedCost`,
		`s.sessionCostQuote`,
		`__('multi_currency')`,
	} {
		if !strings.Contains(js, marker) {
			t.Fatalf("workspace JavaScript missing %s", marker)
		}
	}
	pdfBranch := strings.SplitN(js, `if(preview.kind==='pdf')`, 2)
	if len(pdfBranch) != 2 || strings.Contains(strings.SplitN(pdfBranch[1], "return;", 2)[0], "sandbox") {
		t.Fatal("PDF preview must not sandbox Chrome's built-in PDF viewer")
	}
}
