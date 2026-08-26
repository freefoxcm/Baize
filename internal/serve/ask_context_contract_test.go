package serve

import (
	"strings"
	"testing"
)

func TestServeAskReviewContextContract(t *testing.T) {
	js, css := string(baizeJS), string(baizeCSS)
	for _, marker := range []string{
		`const reviewContext = String(ask.context || '').trim()`, `contextMarkdown.innerHTML = renderMarkdown(reviewContext)`,
		`fixImageSrcs(contextMarkdown)`, `contextMarkdown.textContent.replace(/\s+/g, ' ').trim()`,
		`ask._contextExpanded = !ask._contextExpanded`, `ask._contextScrollTop = contextBody.scrollTop`,
		`requestAnimationFrame(() => { scheduleComposerLayoutSync(); scrollDown(true); })`, `'ask_context_title': '确认内容'`,
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("Ask review context JavaScript missing %q", marker)
		}
	}
	for _, marker := range []string{
		`.ask__context{overflow:hidden;flex-shrink:0;`, `.ask__context-summary{display:-webkit-box;overflow:hidden;`,
		`-webkit-line-clamp:3`, `.ask__context-summary[hidden]{display:none}`,
		`.ask__context-body{max-height:min(32vh,160px);overflow:auto;overscroll-behavior:contain;`,
		`.ask__context-body[hidden]{display:none}`, `.ask__context-body{max-height:min(30dvh,120px);padding:8px 10px}`,
		`.ask__rows{display:flex;flex:1 1 auto;min-height:0;`, `overflow-y:auto;overscroll-behavior:contain`,
	} {
		if !strings.Contains(css, marker) {
			t.Errorf("Ask review context CSS missing %q", marker)
		}
	}
}
