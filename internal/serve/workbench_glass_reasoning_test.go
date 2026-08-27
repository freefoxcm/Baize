package serve

import (
	"strings"
	"testing"
)

func TestReasoningClosedModeSettlesToHeaderOnly(t *testing.T) {
	css := string(baizeCSS)
	js := string(baizeJS)
	for _, want := range []string{
		`.reasoning[data-running="false"][data-display="closed"] .reasoning__summary{display:none!important}`,
		`function reasoningDisplayMode()`,
		`w.dataset.display=display`,
		`setReasoningOpen(w,display==='open'||(display==='auto'&&!it.done),!it.done||display!=='closed')`,
		`const open=it.reasoningUserOverride?!!body&&body.style.display!=='none':display==='open'`,
		`setReasoningOpen(r,open,display!=='closed')`,
	} {
		if !strings.Contains(css+js, want) {
			t.Errorf("Baize reasoning presentation missing %q", want)
		}
	}
}

func TestServeLeftWorkbenchGlassContract(t *testing.T) {
	css := string(baizeCSS)
	for _, want := range []string{
		`--workbench-glass:rgba(24,23,22,.68)`,
		`--workbench-glass:rgba(238,231,220,.74)`,
		`--workbench-glass-filter:blur(18px) saturate(120%)`,
		`--workbench-glass-filter:blur(18px) saturate(114%)`,
		`.app::before{content:'';grid-column:1/3;grid-row:1;z-index:0;min-width:0;background:var(--workbench-ambient);pointer-events:none}`,
		`.activity-rail,.context-panel,.workspace-panel,.settings-drawer{-webkit-backdrop-filter:var(--workbench-glass-filter);backdrop-filter:var(--workbench-glass-filter)}`,
		`.workspace-panel{position:relative;width:100%;height:100%;min-width:0;display:none;flex-direction:column;background:var(--workbench-glass-surface);overflow:hidden}`,
		`.settings-drawer{position:relative;width:100%;height:100%;min-width:0;display:none;flex-direction:column;background:var(--workbench-glass-surface);overflow:hidden}`,
		`.app--mobile-workbench-open::before{display:block}`,
		`.activity-rail{position:fixed;inset:0 auto auto -56px`,
		`.context-panel{position:fixed;inset:0 auto auto calc(-100vw + 56px)`,
		`.app--mobile-workbench-open .activity-rail{left:0}`,
		`.app--mobile-workbench-open .context-panel{left:56px}`,
		`--workbench-glass-filter:none`,
		`.app::before{display:none!important}`,
	} {
		if !strings.Contains(css, want) {
			t.Errorf("Serve workbench glass contract missing %q", want)
		}
	}
}
