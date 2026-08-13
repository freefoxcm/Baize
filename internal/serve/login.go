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
	Lang                 string
	Title                string
	Tagline              string
	TaglineSecondary     string
	TaglineSecondaryLang string
	PortalLabel          string
	Welcome              string
	WelcomeSecondary     string
	WelcomeSecondaryLang string
	Instruction          string
	PasswordLabel        string
	SubmitLabel          string
	PrivacyNote          string
	Errors               map[loginErrorKey]string
}

var englishLoginCopy = loginCopy{
	Lang:                 "en",
	Title:                "Baize — Login",
	Tagline:              "Bring clarity to complex work.",
	TaglineSecondary:     "让复杂任务，回归清晰秩序。",
	TaglineSecondaryLang: "zh-CN",
	PortalLabel:          "ACCESS PORTAL",
	Welcome:              "Welcome back",
	WelcomeSecondary:     "欢迎回来",
	WelcomeSecondaryLang: "zh-CN",
	Instruction:          "Enter your access password to continue",
	PasswordLabel:        "Access password / 访问密码",
	SubmitLabel:          "Enter Baize / 进入 Baize",
	PrivacyNote:          "Your password is used only for authentication.",
	Errors: map[loginErrorKey]string{
		loginErrorRequired: "Password is required.",
		loginErrorInvalid:  "Invalid password.",
		loginErrorConfig:   "Server is not configured for password authentication.",
		loginErrorRate:     "Too many attempts. Please wait a minute.",
	},
}

var chineseLoginCopy = loginCopy{
	Lang:                 "zh-CN",
	Title:                "Baize — 登录",
	Tagline:              "让复杂任务，回归清晰秩序。",
	TaglineSecondary:     "Bring clarity to complex work.",
	TaglineSecondaryLang: "en",
	PortalLabel:          "访问入口",
	Welcome:              "欢迎回来",
	WelcomeSecondary:     "Welcome back",
	WelcomeSecondaryLang: "en",
	Instruction:          "输入访问密码以继续",
	PasswordLabel:        "访问密码 / Access password",
	SubmitLabel:          "进入 Baize / Enter Baize",
	PrivacyNote:          "密码仅用于本次身份验证",
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
