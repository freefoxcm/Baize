package serve

import (
	"embed"
	"mime"
	"net/http"
	"path"
	"strings"

	"reasonix/internal/config"
)

//go:embed index.html
var indexHTML []byte

//go:embed logo-wordmark.svg
var logoWordmarkSVG []byte

//go:embed logo-symbol.svg
var logoSymbolSVG []byte

//go:embed assets/vendor.min.js
var vendorJS []byte

//go:embed assets/baize.css
var baizeCSS []byte

//go:embed assets/baize.js
var baizeJS []byte

//go:embed assets/login.css
var loginCSS []byte

//go:embed assets/login-bg-desktop.webp
var loginBackgroundDesktop []byte

//go:embed assets/login-bg-mobile.webp
var loginBackgroundMobile []byte

//go:embed assets/pdfjs
var pdfJSAssets embed.FS

func (s *Server) registerWebAssetRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", s.index)
	mux.HandleFunc("GET /sessions/{id}", s.index)
	mux.HandleFunc("GET /assets/logo-wordmark.svg", s.logoWordmark)
	mux.HandleFunc("GET /assets/logo-symbol.svg", s.logoSymbol)
	mux.HandleFunc("GET /assets/login.css", s.loginCSSHandler)
	mux.HandleFunc("GET /assets/login-bg-desktop.webp", s.loginBackgroundDesktopHandler)
	mux.HandleFunc("GET /assets/login-bg-mobile.webp", s.loginBackgroundMobileHandler)
	mux.HandleFunc("GET /assets/vendor.min.js", s.vendorJSHandler)
	mux.HandleFunc("GET /assets/baize.css", s.baizeCSSHandler)
	mux.HandleFunc("GET /assets/baize.js", s.baizeJSHandler)
	mux.HandleFunc("GET /assets/pdfjs/", s.pdfJSAssetHandler)
	mux.HandleFunc("GET /workspace/entries", s.workspaceEntries)
	mux.HandleFunc("GET /workspace/search", s.workspaceSearch)
	mux.HandleFunc("GET /workspace/preview", s.workspacePreview)
	mux.HandleFunc("GET /workspace/content", s.workspaceContent)
	mux.HandleFunc("HEAD /workspace/content", s.workspaceContent)
}

func (s *Server) index(w http.ResponseWriter, _ *http.Request) {
	if setup, ok := s.providerSetupSnapshot(); ok && setup.Required {
		s.providerSetupIndex(w)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = config.MigrateLegacyIfNeeded()
	lang := "auto"
	if cfg, err := config.Load(); err == nil {
		if dl := cfg.DesktopLanguage(); dl != "" {
			lang = dl
		}
	}
	html := strings.ReplaceAll(string(indexHTML), "__LANG__", lang)
	_, _ = w.Write([]byte(html))
}

func (s *Server) logoSymbol(w http.ResponseWriter, _ *http.Request) {
	writeWebAsset(w, "image/svg+xml; charset=utf-8", "public, max-age=3600", logoSymbolSVG)
}

func (s *Server) logoWordmark(w http.ResponseWriter, _ *http.Request) {
	writeWebAsset(w, "image/svg+xml; charset=utf-8", "public, max-age=3600", logoWordmarkSVG)
}

func (s *Server) loginCSSHandler(w http.ResponseWriter, _ *http.Request) {
	writeWebAsset(w, "text/css; charset=utf-8", "no-cache", loginCSS)
}

func (s *Server) loginBackgroundDesktopHandler(w http.ResponseWriter, _ *http.Request) {
	writeWebAsset(w, "image/webp", "public, max-age=86400", loginBackgroundDesktop)
}

func (s *Server) loginBackgroundMobileHandler(w http.ResponseWriter, _ *http.Request) {
	writeWebAsset(w, "image/webp", "public, max-age=86400", loginBackgroundMobile)
}

func (s *Server) vendorJSHandler(w http.ResponseWriter, _ *http.Request) {
	writeWebAsset(w, "application/javascript; charset=utf-8", "public, max-age=86400", vendorJS)
}

func (s *Server) baizeCSSHandler(w http.ResponseWriter, _ *http.Request) {
	writeWebAsset(w, "text/css; charset=utf-8", "no-cache", baizeCSS)
}

func (s *Server) baizeJSHandler(w http.ResponseWriter, _ *http.Request) {
	writeWebAsset(w, "application/javascript; charset=utf-8", "no-cache", baizeJS)
}

func (s *Server) pdfJSAssetHandler(w http.ResponseWriter, r *http.Request) {
	name := path.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if !strings.HasPrefix(name, "assets/pdfjs/") {
		http.NotFound(w, r)
		return
	}
	body, err := pdfJSAssets.ReadFile(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	ext := path.Ext(name)
	contentType := mime.TypeByExtension(ext)
	if ext == ".mjs" {
		// Windows commonly registers .mjs as text/plain. PDF.js modules must
		// always be served as JavaScript regardless of the host MIME registry.
		contentType = "text/javascript; charset=utf-8"
	} else if contentType == "" {
		contentType = "application/octet-stream"
	}
	writeWebAsset(w, contentType, "public, max-age=86400", body)
}

func writeWebAsset(w http.ResponseWriter, contentType, cacheControl string, body []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", cacheControl)
	_, _ = w.Write(body)
}
