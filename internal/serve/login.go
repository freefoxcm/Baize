package serve

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

//go:embed login.html
var loginHTML []byte

var loginTemplate = template.Must(template.New("login").Parse(string(loginHTML)))

type loginErrorKey string

const (
	loginErrorRequired loginErrorKey = "required"
	loginErrorInvalid  loginErrorKey = "invalid"
	loginErrorConfig   loginErrorKey = "config"
	loginErrorRate     loginErrorKey = "rate"
)

type loginCopy struct {
	Lang          string
	Title         string
	LoginLabel    string
	PasswordLabel string
	SubmitLabel   string
	Errors        map[loginErrorKey]string
}

var englishLoginCopy = loginCopy{
	Lang:          "en",
	Title:         "Baize — Login",
	LoginLabel:    "Baize login",
	PasswordLabel: "Access password",
	SubmitLabel:   "Continue",
	Errors: map[loginErrorKey]string{
		loginErrorRequired: "Password is required.",
		loginErrorInvalid:  "Invalid password.",
		loginErrorConfig:   "Server is not configured for password authentication.",
		loginErrorRate:     "Too many attempts. Please wait a minute.",
	},
}

var chineseLoginCopy = loginCopy{
	Lang:          "zh-CN",
	Title:         "Baize — 登录",
	LoginLabel:    "Baize 登录",
	PasswordLabel: "访问密码",
	SubmitLabel:   "继续",
	Errors: map[loginErrorKey]string{
		loginErrorRequired: "请输入访问密码。",
		loginErrorInvalid:  "访问密码不正确。",
		loginErrorConfig:   "服务器尚未配置密码验证。",
		loginErrorRate:     "尝试次数过多，请一分钟后再试。",
	},
}

type loginPageData struct {
	loginCopy
	Error string
}

func (ag *authGate) loginPage(w http.ResponseWriter, r *http.Request) {
	ag.renderLoginPage(w, r, http.StatusOK, "")
}

func (ag *authGate) loginPageWithError(w http.ResponseWriter, r *http.Request, status int, key loginErrorKey) {
	ag.renderLoginPage(w, r, status, key)
}

func (ag *authGate) renderLoginPage(w http.ResponseWriter, r *http.Request, status int, key loginErrorKey) {
	copy := loginCopyForRequest(r)
	data := loginPageData{loginCopy: copy, Error: copy.Errors[key]}
	var body bytes.Buffer
	if err := loginTemplate.Execute(&body, data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Language", copy.Lang)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Vary", "Accept-Language")
	w.WriteHeader(status)
	_, _ = w.Write(body.Bytes())
}

func loginCopyForRequest(r *http.Request) loginCopy {
	selected := englishLoginCopy
	bestQuality := -1.0
	for preference := range strings.SplitSeq(r.Header.Get("Accept-Language"), ",") {
		parts := strings.Split(preference, ";")
		language := strings.ToLower(strings.TrimSpace(parts[0]))
		quality := 1.0
		for _, parameter := range parts[1:] {
			name, value, ok := strings.Cut(strings.TrimSpace(parameter), "=")
			if ok && strings.EqualFold(name, "q") {
				if parsed, err := strconv.ParseFloat(value, 64); err == nil {
					quality = parsed
				}
			}
		}
		if quality <= 0 || quality <= bestQuality {
			continue
		}
		switch {
		case language == "zh", strings.HasPrefix(language, "zh-"):
			selected = chineseLoginCopy
			bestQuality = quality
		case language == "en", strings.HasPrefix(language, "en-"):
			selected = englishLoginCopy
			bestQuality = quality
		}
	}
	return selected
}
