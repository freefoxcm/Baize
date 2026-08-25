package serve

import (
	"strings"
	"testing"
)

func TestServeFooterComposerGlassContract(t *testing.T) {
	css := string(baizeCSS)
	for _, want := range []string{
		`--composer-glass-border:var(--border-strong);--composer-glass-shadow:var(--shadow-md)`,
		`--glass-composer:rgba(48,44,40,.66)`,
		`--composer-glass-border:rgba(244,233,215,.24);--composer-glass-shadow:0 14px 34px rgba(0,0,0,.28),inset 0 1px 0 rgba(255,255,255,.08)`,
		`--glass-composer:rgba(255,250,242,.72)`,
		`--composer-glass-border:rgba(49,43,36,.18);--composer-glass-shadow:0 14px 34px rgba(73,60,43,.14),inset 0 1px 0 rgba(255,255,255,.72)`,
		`.footer{grid-column:3;grid-row:2;border-top:1px solid color-mix(in srgb,var(--border) 70%,transparent);background:transparent;`,
		`.composer-card{position:relative;width:100%;max-width:var(--chat-maxw);margin:0 auto;background:var(--glass-composer);color:var(--composer-fg);border:1px solid var(--composer-glass-border);`,
		`.composer-card:focus-within{border-color:var(--accent);box-shadow:var(--composer-glass-shadow),0 0 0 3px var(--accent-soft)}`,
		`@media(prefers-reduced-transparency:reduce)`,
		`:root,:root[data-theme="charcoal-copper"]{--glass-panel:var(--panel);--glass-sidebar:var(--sidebar-bg);--glass-composer:var(--composer-bg);--glass-popover:var(--card);--glass-filter:none;--composer-glass-border:var(--border-strong);--composer-glass-shadow:var(--shadow-md)}`,
	} {
		if !strings.Contains(css, want) {
			t.Errorf("Serve footer/composer glass CSS missing %q", want)
		}
	}
	for _, unwanted := range []string{`--glass-footer`, `.footer::before`, `.footer,.activity-rail`} {
		if strings.Contains(css, unwanted) {
			t.Errorf("Serve footer still carries obsolete glass layer %q", unwanted)
		}
	}
}
