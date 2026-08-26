package serve

import (
	"strings"
	"testing"
)

func TestServeFooterComposerGlassContract(t *testing.T) {
	html := string(indexHTML)
	css := string(baizeCSS)
	for _, want := range []string{
		`--composer-glass-filter:none;--composer-glass-toolbar:var(--composer-bg);--composer-glass-control:var(--panel-2);`,
		`--transcript-scrollbar-width:0px;--composer-overlay-height:156px`,
		`--composer-window-left:max(28px,calc((100% - var(--chat-maxw))/2))`,
		`--composer-window-width:min(var(--chat-maxw),calc(100% - 56px))`,
		`--composer-window-radius:var(--radius-lg)`,
		`--glass-composer:rgba(48,44,40,.58)`,
		`--composer-glass-filter:blur(20px) saturate(125%)`,
		`--composer-glass-toolbar:rgba(48,44,40,.24);--composer-glass-control:rgba(52,48,44,.56)`,
		`--composer-glass-border:rgba(244,233,215,.22)`,
		`--glass-composer:rgba(255,250,242,.68)`,
		`--composer-glass-filter:blur(20px) saturate(116%)`,
		`--composer-glass-toolbar:rgba(255,250,242,.28);--composer-glass-control:rgba(235,227,216,.62)`,
		`.app{display:grid;grid-template-columns:var(--rail-width) var(--context-width) minmax(0,1fr);grid-template-rows:minmax(0,1fr);`,
		`.transcript{grid-column:3;grid-row:1;min-height:0;overflow-y:auto;padding:24px 28px calc(var(--composer-overlay-height) + 20px);scroll-padding-bottom:calc(var(--composer-overlay-height) + 20px);`,
		`.transcript>:not(.welcome){max-width:var(--chat-maxw);margin-inline:auto}`,
		`.composer-window-mask{grid-column:3;grid-row:1;align-self:end;height:var(--composer-overlay-height);position:relative;z-index:19;overflow:hidden;pointer-events:none}`,
		`.composer-window-mask__guard{position:absolute;left:var(--composer-window-left);top:var(--composer-window-top);width:var(--composer-window-width);height:var(--composer-window-height);background:radial-gradient`,
		`.composer-window-mask__guard::after{content:'';position:absolute;top:calc(100% - 1px);left:0;width:100%;height:var(--app-viewport-height,100dvh);background:var(--bg)}`,
		`.footer{grid-column:3;grid-row:1;align-self:end;z-index:20;border:0;background:transparent;padding:12px calc(28px + var(--transcript-scrollbar-width,0px)) 16px 28px;`,
		`.footer>*{pointer-events:auto}`,
		`.activity-rail{grid-column:1;grid-row:1/-1;`,
		`.context-panel{grid-column:2;grid-row:1/-1;`,
		`.composer-card{position:relative;width:100%;max-width:var(--chat-maxw);margin:0 auto;background:var(--glass-composer);`,
		`-webkit-backdrop-filter:var(--composer-glass-filter);backdrop-filter:var(--composer-glass-filter);`,
		`.composer-card:focus-within{border-color:color-mix(in srgb,var(--accent) 62%,var(--composer-glass-border));box-shadow:var(--composer-glass-shadow),0 0 0 2px color-mix(in srgb,var(--accent) 18%,transparent)}`,
		`.composer-meta{display:flex;align-items:center;gap:10px;padding:8px 12px;border-top:1px solid var(--border);background:var(--composer-glass-toolbar);`,
		`:root{--composer-overlay-height:126px;--composer-window-left:max(12px,env(safe-area-inset-left))`,
		`.transcript{grid-column:1;grid-row:1;padding:16px max(16px,env(safe-area-inset-right)) calc(var(--composer-overlay-height) + 20px) max(16px,env(safe-area-inset-left));scroll-padding-bottom:calc(var(--composer-overlay-height) + 20px)}`,
		`.composer-window-mask{grid-column:1;grid-row:1}`,
		`.footer{grid-column:1;grid-row:1;padding:8px calc(max(12px,env(safe-area-inset-right)) + var(--transcript-scrollbar-width,0px)) calc(8px + env(safe-area-inset-bottom))`,
		`@media(prefers-reduced-transparency:reduce)`,
		`--composer-glass-filter:none;--composer-glass-toolbar:var(--composer-bg);--composer-glass-control:var(--panel-2);`,
	} {
		if !strings.Contains(css, want) {
			t.Errorf("Serve footer/composer glass CSS missing %q", want)
		}
	}
	for _, unwanted := range []string{`--glass-footer`, `--composer-glass-sheen`, `--composer-glass-fade`, `--footer-bottom-guard`, `--footer-content-fade`, `100vmax`, `.composer-window-mask__aperture`, `.footer::before`, `.footer::after`, `isolation:isolate`, `.footer{grid-column:3;grid-row:2`, `color-mix(in srgb,var(--glass-composer) 86%,transparent)`} {
		if strings.Contains(css, unwanted) {
			t.Errorf("Serve footer still carries obsolete glass layer %q", unwanted)
		}
	}
	mask := strings.Index(html, `<div class="composer-window-mask" aria-hidden="true">`)
	footer := strings.Index(html, `<footer class="footer">`)
	if mask < 0 || footer < 0 || mask > footer {
		t.Fatal("composer aperture mask must be mounted immediately below the floating footer layer")
	}
}

func TestServeActivityRailTooltipContract(t *testing.T) {
	css := string(baizeCSS)
	for _, want := range []string{
		`.activity-rail__tooltip{position:absolute;z-index:250;top:50%;left:calc(100% + 10px);`,
		`transform:translateY(-50%)`,
		`@media(hover:none) and (pointer:coarse) and (any-hover:none){.activity-rail__tooltip{display:none!important}}`,
	} {
		if !strings.Contains(css, want) {
			t.Errorf("Serve activity rail tooltip CSS missing %q", want)
		}
	}
	if strings.Contains(css, `.activity-rail__tooltip{position:fixed`) {
		t.Error("Serve activity rail tooltip must not use fixed positioning inside the filtered rail")
	}
}

func TestServeComposerLayoutSyncContract(t *testing.T) {
	js := string(baizeJS)
	for _, want := range []string{
		`composerFooter = $('.footer'), composerCard = $('.composer-card'), composerWindowMask = $('.composer-window-mask')`,
		`const keepPinned=pinnedToBottom`,
		`const footerRect=composerFooter.getBoundingClientRect(),cardRect=composerCard.getBoundingClientRect(),maskRect=composerWindowMask.getBoundingClientRect()`,
		`Math.ceil(log.getBoundingClientRect().width-log.clientWidth)`,
		`'--composer-overlay-height':Math.ceil(footerRect.height)+'px'`,
		`'--composer-window-left':Math.max(0,cardRect.left-maskRect.left)+'px'`,
		`'--composer-window-top':Math.max(0,cardRect.top-footerRect.top)+'px'`,
		`'--composer-window-width':Math.max(0,cardRect.width)+'px'`,
		`'--composer-window-radius':cardRadius+'px'`,
		`if(layoutChanged)requestAnimationFrame(scheduleComposerLayoutSync)`,
		`if(keepPinned&&pinnedToBottom)requestAnimationFrame(()=>scrollDown(false))`,
		`const composerLayoutResizeObserver=new ResizeObserver(scheduleComposerLayoutSync)`,
		`composerLayoutResizeObserver.observe(composerFooter)`,
		`composerLayoutResizeObserver.observe(composerCard)`,
		`composerLayoutResizeObserver.observe(composerWindowMask)`,
		`composerLayoutResizeObserver.observe(log)`,
		`const composerLayoutMutationObserver=new MutationObserver(scheduleComposerLayoutSync)`,
		`composerLayoutMutationObserver.observe(composerFooter,{attributes:true,childList:true,subtree:true,characterData:true})`,
		`const transcriptLayoutMutationObserver=new MutationObserver(scheduleComposerLayoutSync)`,
		`transcriptLayoutMutationObserver.observe(log,{childList:true,subtree:true,characterData:true})`,
		`window.addEventListener('resize',scheduleComposerLayoutSync)`,
		`scheduleComposerLayoutSync();`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("Serve composer overlay JavaScript missing %q", want)
		}
	}
	for _, unwanted := range []string{`scheduleComposerOverlaySync`, `composerOverlaySyncFrame`} {
		if strings.Contains(js, unwanted) {
			t.Errorf("Serve composer layout JavaScript still carries overlay logic %q", unwanted)
		}
	}
}

func TestServeQuestionJumpBarLeftEdgeContract(t *testing.T) {
	css := string(baizeCSS)
	js := string(baizeJS)
	for _, want := range []string{
		`.app > .jump-bar{position:absolute;top:50%;left:calc(var(--rail-width) + var(--context-width));right:auto;`,
		`.jump-bar{z-index:30;display:flex;flex-direction:column;align-items:flex-start;`,
		`justify-content:flex-start`,
		`.jump-preview{position:absolute;left:100%;right:auto;`,
		`A thin rail on the transcript's left edge marks each user question.`,
	} {
		if !strings.Contains(css+js, want) {
			t.Errorf("Serve left question jump bar missing %q", want)
		}
	}
}
