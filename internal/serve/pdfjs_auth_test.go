package serve

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"reasonix/internal/config"
)

func TestPDFJSAssetsRequireServeAuthentication(t *testing.T) {
	for _, cfg := range []config.ServeConfig{
		{AuthMode: "token", Token: "secret"},
		{AuthMode: "password", PasswordHash: mustHash("test")},
	} {
		gate := newAuthGate(cfg)
		handler := gate.middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/assets/pdfjs/pdf.mjs", nil))
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("PDF.js asset status = %d, want 401", recorder.Code)
		}
	}
}
