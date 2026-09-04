package provider

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const openCodeSessionHeader = "x-opencode-session"

type sessionIDContextKey struct{}

// WithSessionID binds the active conversation identity to provider calls.
func WithSessionID(ctx context.Context, id string) context.Context {
	id = strings.TrimSpace(id)
	if ctx == nil || id == "" {
		return ctx
	}
	return context.WithValue(ctx, sessionIDContextKey{}, id)
}

// SessionID returns the active provider conversation identity, when present.
func SessionID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(sessionIDContextKey{}).(string)
	return strings.TrimSpace(id)
}

// EnsureSessionID preserves a durable host session ID or mints an opaque ID
// for an ephemeral run. Callers retain the result for the whole conversation.
func EnsureSessionID(id string) string {
	if id = strings.TrimSpace(id); id != "" {
		return id
	}
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err == nil {
		return "ses_" + hex.EncodeToString(raw[:])
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d-%p", time.Now().UnixNano(), &raw)))
	return "ses_" + hex.EncodeToString(sum[:16])
}

// ApplyOpenCodeSessionHeader sends the stable conversation identity expected
// by OpenCode Go and OpenCode-compatible proxy provider entries.
func ApplyOpenCodeSessionHeader(ctx context.Context, header http.Header, name, baseURL string) {
	if header == nil || !usesOpenCodeSessionHeader(name, baseURL) {
		return
	}
	if id := SessionID(ctx); id != "" {
		header.Set(openCodeSessionHeader, id)
	}
}

// ApplyOpenCodeSessionRequest applies the OpenCode conversation identity and
// returns req so adapters can finalize a request without adding another step.
func ApplyOpenCodeSessionRequest(ctx context.Context, req *http.Request, name, baseURL string) *http.Request {
	if req != nil {
		ApplyOpenCodeSessionHeader(ctx, req.Header, name, baseURL)
	}
	return req
}

func usesOpenCodeSessionHeader(name, baseURL string) bool {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(name)), "opencode") {
		return true
	}
	path, ok := officialOpenCodeGoPath(baseURL)
	return ok && (path == "/zen" || strings.HasPrefix(path, "/zen/"))
}
