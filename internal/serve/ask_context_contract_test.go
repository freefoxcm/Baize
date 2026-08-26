package serve

import (
	"strings"
	"testing"
)

func TestServeAskReviewContextContract(t *testing.T) {
	js, css := string(baizeJS), string(baizeCSS)
	for _, marker := range []string{
		`const reviewContext = String(ask.context || '').trim()`, `contextMarkdown.innerHTML = renderMarkdown(reviewContext)`,
		`fixImageSrcs(contextMarkdown)`, `ask._contextCollapsed = !ask._contextCollapsed`, `ask._contextScrollTop = contextBody.scrollTop`,
		`requestAnimationFrame(() => { scheduleComposerLayoutSync(); scrollDown(true); })`, `'ask_context_title': '确认内容'`,
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("Ask review context JavaScript missing %q", marker)
		}
	}
	for _, marker := range []string{
		`.ask__context{overflow:hidden;flex-shrink:0;`, `.ask__context-body{max-height:min(32vh,300px);overflow:auto;overscroll-behavior:contain;`,
		`.ask__context-body[hidden]{display:none}`, `.ask__context-body{max-height:min(30dvh,240px);padding:9px 10px}`,
	} {
		if !strings.Contains(css, marker) {
			t.Errorf("Ask review context CSS missing %q", marker)
		}
	}
}
