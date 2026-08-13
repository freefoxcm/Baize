package serve

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"reasonix/internal/config"
	"reasonix/internal/control"
)

func TestPasswordModeLoginPageLocalization(t *testing.T) {
	ag := newAuthGate(config.ServeConfig{AuthMode: "password", PasswordHash: mustHash("correct")})
	tests := []struct {
		name     string
		language string
		wantLang string
		want     []string
	}{
		{
			name: "english quality preference", language: "zh-CN;q=0.7,en-US;q=0.9", wantLang: "en",
			want: []string{"Welcome back", "Enter your access password to continue", "Bring clarity to complex work."},
		},
		{
			name: "chinese", language: "zh-CN,zh;q=0.9,en;q=0.7", wantLang: "zh-CN",
			want: []string{"欢迎回来", "输入访问密码以继续", "让复杂任务，回归清晰秩序。"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/login", nil)
			request.Header.Set("Accept-Language", test.language)
			ag.middleware(http.NotFoundHandler()).ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", recorder.Code)
			}
			if got := recorder.Header().Get("Content-Language"); got != test.wantLang {
				t.Fatalf("Content-Language = %q, want %q", got, test.wantLang)
			}
			for _, want := range test.want {
				if !strings.Contains(recorder.Body.String(), want) {
					t.Errorf("localized login page missing %q", want)
				}
			}
		})
	}
}

func TestPasswordModeLoginErrorLocalization(t *testing.T) {
	tests := []struct {
		name     string
		language string
		want     string
	}{
		{name: "english", language: "en", want: "Invalid password."},
		{name: "chinese", language: "zh-CN", want: "访问密码不正确。"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ag := newAuthGate(config.ServeConfig{AuthMode: "password", PasswordHash: mustHash("correct")})
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("password=wrong"))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("Accept-Language", test.language)
			ag.middleware(http.NotFoundHandler()).ServeHTTP(recorder, request)
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", recorder.Code)
			}
			if !strings.Contains(recorder.Body.String(), test.want) {
				t.Fatalf("localized error page missing %q", test.want)
			}
			if !strings.Contains(recorder.Body.String(), `role="alert"`) {
				t.Fatal("error page is missing an accessible alert")
			}
		})
	}
}

func TestPasswordModeRateLimitKeepsStatusAndLocalizedPage(t *testing.T) {
	ag := newAuthGate(config.ServeConfig{AuthMode: "password", PasswordHash: mustHash("correct")})
	handler := ag.middleware(http.NotFoundHandler())
	for attempt := 1; attempt <= rateLimitMax+1; attempt++ {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("password="))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("Accept-Language", "zh-CN")
		request.RemoteAddr = "192.0.2.44:4321"
		handler.ServeHTTP(recorder, request)
		if attempt <= rateLimitMax && recorder.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d, want 401", attempt, recorder.Code)
		}
		if attempt == rateLimitMax+1 {
			if recorder.Code != http.StatusTooManyRequests {
				t.Fatalf("rate-limited status = %d, want 429", recorder.Code)
			}
			if !strings.Contains(recorder.Body.String(), "尝试次数过多，请一分钟后再试。") {
				t.Fatal("rate-limited page is not localized")
			}
		}
	}
}

func TestLoginAssetsArePublicInAuthenticatedModes(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.ServeConfig
	}{
		{name: "token", cfg: config.ServeConfig{AuthMode: "token", Token: "secret"}},
		{name: "password", cfg: config.ServeConfig{AuthMode: "password", PasswordHash: mustHash("correct")}},
	}
	assets := []string{
		"/assets/logo-wordmark.svg", "/assets/logo-symbol.svg", "/assets/login.css",
		"/assets/login-bg-desktop.webp", "/assets/login-bg-mobile.webp",
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := newAuthGate(test.cfg).middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			for _, method := range []string{http.MethodGet, http.MethodHead} {
				for _, asset := range assets {
					recorder := httptest.NewRecorder()
					handler.ServeHTTP(recorder, httptest.NewRequest(method, asset, nil))
					if recorder.Code != http.StatusOK {
						t.Errorf("%s %s status = %d, want 200", method, asset, recorder.Code)
					}
				}
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/assets/vendor.min.js", nil))
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("protected asset status = %d, want 401", recorder.Code)
			}
		})
	}
}

func TestServePublicLoginAssetRoutes(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	t.Cleanup(func() { ctrl.Close() })
	handler := New(ctrl, bc, config.ServeConfig{}).Handler()
	tests := []struct {
		path        string
		contentType string
		cache       string
		body        string
	}{
		{path: "/assets/logo-wordmark.svg", contentType: "image/svg+xml; charset=utf-8", cache: "public, max-age=3600", body: `aria-label="Baize"`},
		{path: "/assets/logo-symbol.svg", contentType: "image/svg+xml; charset=utf-8", cache: "public, max-age=3600", body: `aria-label="Baize"`},
		{path: "/assets/login.css", contentType: "text/css; charset=utf-8", cache: "no-cache", body: ":root"},
		{path: "/assets/login-bg-desktop.webp", contentType: "image/webp", cache: "public, max-age=86400", body: "RIFF"},
		{path: "/assets/login-bg-mobile.webp", contentType: "image/webp", cache: "public, max-age=86400", body: "RIFF"},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			for _, method := range []string{http.MethodGet, http.MethodHead} {
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, httptest.NewRequest(method, test.path, nil))
				if recorder.Code != http.StatusOK {
					t.Fatalf("%s status = %d, want 200", method, recorder.Code)
				}
				if got := recorder.Header().Get("Content-Type"); got != test.contentType {
					t.Errorf("Content-Type = %q, want %q", got, test.contentType)
				}
				if got := recorder.Header().Get("Cache-Control"); got != test.cache {
					t.Errorf("Cache-Control = %q, want %q", got, test.cache)
				}
				if method == http.MethodGet && !strings.Contains(recorder.Body.String(), test.body) {
					t.Errorf("response body does not contain %q", test.body)
				}
			}
		})
	}
}
