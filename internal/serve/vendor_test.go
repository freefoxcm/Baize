package serve

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"reasonix/internal/config"
	"reasonix/internal/control"
)

// TestVendorBundleServed checks the embedded rendering libraries are reachable
// and load before the Baize application that consumes them.
func TestVendorBundleServed(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/assets/vendor.min.js")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("vendor status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/javascript") {
		t.Errorf("vendor Content-Type = %q, want application/javascript", ct)
	}
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		t.Fatal(err)
	}
	if buf.Len() < 10_000 {
		t.Fatalf("vendor bundle suspiciously small: %d bytes", buf.Len())
	}
	// Distinct markers from each bundled library survive minification.
	for _, marker := range []string{"DOMPurify", "highlight", "marked"} {
		if !bytes.Contains(buf.Bytes(), []byte(marker)) {
			t.Errorf("vendor bundle missing %q marker", marker)
		}
	}

	// index.html must load the bundle before the Baize application script.
	htmlResp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer htmlResp.Body.Close()
	html := new(bytes.Buffer)
	if _, err := html.ReadFrom(htmlResp.Body); err != nil {
		t.Fatal(err)
	}
	vendorIndex := bytes.Index(html.Bytes(), []byte(`<script src="/assets/vendor.min.js">`))
	if vendorIndex < 0 {
		t.Fatal("index.html does not reference /assets/vendor.min.js")
	}
	appIndex := bytes.Index(html.Bytes(), []byte(`<script src="/assets/baize.js">`))
	if appIndex < 0 || vendorIndex > appIndex {
		t.Error("vendor script must load before the Baize application script")
	}
	if !bytes.Contains(baizeJS, []byte(`const { marked, DOMPurify, hljs } = window.Vendor;`)) {
		t.Error("Baize application must consume the embedded vendor global")
	}
}
