package serve

import (
	"strings"
	"testing"
)

func TestServeFooterComposerGlassContract(t *testing.T) {
	css := string(baizeCSS)
	for _, want := range []string{
		`--composer-glass-filter:none;--composer-glass-sheen:linear-gradient(transparent,transparent);--composer-glass-toolbar:var(--composer-bg);--composer-glass-control:var(--panel-2);`,
		`--composer-overlay-height:156px`,
		`--transcript-scrollbar-width:0px`,
		`--glass-composer:rgba(48,44,40,.54)`,
		`--composer-glass-filter:blur(20px) saturate(125%)`,
		`--composer-glass-toolbar:rgba(48,44,40,.24);--composer-glass-control:rgba(52,48,44,.56)`,
		`--glass-composer:rgba(255,250,242,.64)`,
		`--composer-glass-filter:blur(20px) saturate(116%)`,
		`--composer-glass-toolbar:rgba(255,250,242,.28);--composer-glass-control:rgba(235,227,216,.62)`,
		`.app{display:grid;grid-template-columns:var(--rail-width) var(--context-width) minmax(0,1fr);grid-template-rows:minmax(0,1fr);`,
		`.transcript{grid-column:3;grid-row:1;overflow-y:auto;padding:24px 28px calc(var(--composer-overlay-height,156px) + 20px);scroll-padding-bottom:calc(var(--composer-overlay-height,156px) + 20px);`,
		`.footer{--footer-bottom-guard:16px;--footer-card-offset:12px;--footer-content-fade:28px;grid-column:3;grid-row:1;align-self:end;z-index:20;border:0;background:transparent;padding:12px calc(28px + var(--transcript-scrollbar-width,0px)) 16px 28px;`,
		`.footer::before{content:'';position:absolute;z-index:0;inset:calc(-1 * var(--footer-content-fade)) 0 var(--footer-bottom-guard);background:linear-gradient(to bottom,transparent 0,var(--bg) calc(var(--footer-content-fade) + var(--footer-card-offset)));pointer-events:none}`,
		`.footer::after{content:'';position:absolute;z-index:0;inset:auto 0 0;height:var(--footer-bottom-guard);background:var(--bg);pointer-events:none}`,
		`.footer>*{position:relative;z-index:1;pointer-events:auto}`,
		`.composer-card{position:relative;width:100%;max-width:var(--chat-maxw);margin:0 auto;background:var(--composer-glass-sheen),var(--glass-composer);`,
		`-webkit-backdrop-filter:var(--composer-glass-filter);backdrop-filter:var(--composer-glass-filter);`,
		`.composer-card:focus-within{border-color:color-mix(in srgb,var(--accent) 62%,var(--composer-glass-border));box-shadow:var(--composer-glass-shadow),0 0 0 2px color-mix(in srgb,var(--accent) 18%,transparent)}`,
		`.composer-meta{display:flex;align-items:center;gap:10px;padding:8px 12px;border-top:1px solid var(--border);background:var(--composer-glass-toolbar);`,
		`:root{--composer-overlay-height:126px}`,
		`.footer{--footer-bottom-guard:calc(8px + env(safe-area-inset-bottom));--footer-card-offset:8px;--footer-content-fade:24px;grid-column:1;grid-row:1;padding:8px calc(max(12px,env(safe-area-inset-right)) + var(--transcript-scrollbar-width,0px)) calc(8px + env(safe-area-inset-bottom))`,
		`@media(prefers-reduced-transparency:reduce)`,
		`--composer-glass-filter:none;--composer-glass-sheen:linear-gradient(transparent,transparent);--composer-glass-toolbar:var(--composer-bg);--composer-glass-control:var(--panel-2);`,
	} {
		if !strings.Contains(css, want) {
			t.Errorf("Serve footer/composer glass CSS missing %q", want)
		}
	}
	for _, unwanted := range []string{`--glass-footer`, `.footer,.activity-rail`, `.footer{grid-column:3;grid-row:2`, `color-mix(in srgb,var(--glass-composer) 86%,transparent)`} {
		if strings.Contains(css, unwanted) {
			t.Errorf("Serve footer still carries obsolete glass layer %q", unwanted)
		}
	}
}

func TestServeComposerOverlayHeightContract(t *testing.T) {
	js := string(baizeJS)
	for _, want := range []string{
		`const keepPinned=pinnedToBottom`,
		`Math.ceil(composerFooter.getBoundingClientRect().height)`,
		`document.documentElement.style.setProperty('--composer-overlay-height',value)`,
		`Math.ceil(log.getBoundingClientRect().width-log.clientWidth)`,
		`document.documentElement.style.setProperty('--transcript-scrollbar-width',scrollbarValue)`,
		`if(keepPinned&&pinnedToBottom)requestAnimationFrame(()=>scrollDown(false))`,
		`const composerOverlayResizeObserver=new ResizeObserver(scheduleComposerOverlaySync)`,
		`composerOverlayResizeObserver.observe(composerFooter)`,
		`composerOverlayResizeObserver.observe(log)`,
		`const composerOverlayMutationObserver=new MutationObserver(scheduleComposerOverlaySync)`,
		`composerOverlayMutationObserver.observe(composerFooter,{attributes:true,childList:true,subtree:true,characterData:true})`,
		`const transcriptLayoutMutationObserver=new MutationObserver(scheduleComposerOverlaySync)`,
		`transcriptLayoutMutationObserver.observe(log,{childList:true,subtree:true,characterData:true})`,
		`scheduleComposerOverlaySync();`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("Serve composer overlay JavaScript missing %q", want)
		}
	}
}
