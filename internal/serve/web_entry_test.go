package serve

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"reasonix/internal/config"
	"reasonix/internal/control"
)

func TestServeIndexPageAndSessionDeepLink(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	srv := httptest.NewServer(New(ctrl, bc, config.ServeConfig{}).Handler())
	defer srv.Close()

	for _, path := range []string{"/", "/sessions/reserved-session"} {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s status = %d", path, resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
			t.Errorf("GET %s content-type = %q, want text/html", path, ct)
		}
	}
}

func TestServeWebPagesBootstrapFragmentTokenBeforeRequests(t *testing.T) {
	for name, html := range map[string]string{
		"index":          string(baizeJS),
		"provider setup": string(providerSetupHTML),
	} {
		t.Run(name, func(t *testing.T) {
			for _, want := range []string{
				"new URLSearchParams(window.location.hash.slice(1))",
				"'/auth/token'",
				"window.history.replaceState",
				"window.fetch",
			} {
				if !strings.Contains(html, want) {
					t.Fatalf("page missing fragment-token bootstrap %q", want)
				}
			}
		})
	}
	if !strings.Contains(string(baizeJS), "__authReady.then(connectEvents)") {
		t.Fatal("serve index must delay SSE until fragment authentication completes")
	}
}

func TestServeIndexLoadsExternalBaizeAssets(t *testing.T) {
	html := string(indexHTML)
	for _, want := range []string{
		`data-language="__LANG__"`,
		`<link rel="stylesheet" href="/assets/baize.css" />`,
		`<script src="/assets/vendor.min.js"></script>`,
		`<script src="/assets/baize.js"></script>`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("serve index missing external asset contract %q", want)
		}
	}
	if strings.Contains(html, "const { marked, DOMPurify, hljs }") || strings.Contains(html, "*,*::before,*::after") {
		t.Fatal("serve index still contains the extracted Baize application source")
	}
	if strings.Index(html, "/assets/vendor.min.js") > strings.Index(html, "/assets/baize.js") {
		t.Fatal("Baize application script must load after its vendor bundle")
	}
}

func TestServeBaizeAssetRoutes(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	t.Cleanup(ctrl.Close)
	handler := New(ctrl, bc, config.ServeConfig{}).Handler()

	for _, tc := range []struct {
		path        string
		contentType string
		body        string
	}{
		{path: "/assets/baize.css", contentType: "text/css; charset=utf-8", body: ".card-main"},
		{path: "/assets/baize.js", contentType: "application/javascript; charset=utf-8", body: "const __LANG_PREF"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want 200", tc.path, recorder.Code)
			}
			if got := recorder.Header().Get("Content-Type"); got != tc.contentType {
				t.Fatalf("GET %s Content-Type = %q, want %q", tc.path, got, tc.contentType)
			}
			if got := recorder.Header().Get("Cache-Control"); got != "no-cache" {
				t.Fatalf("GET %s Cache-Control = %q, want no-cache", tc.path, got)
			}
			if !strings.Contains(recorder.Body.String(), tc.body) {
				t.Fatalf("GET %s missing source marker %q", tc.path, tc.body)
			}
		})
	}
}

func TestServeClipboardHasInsecureContextFallback(t *testing.T) {
	source := string(baizeJS)
	for _, want := range []string{
		"navigator.clipboard?.writeText",
		"document.execCommand('copy')",
		"return fallbackCopyText(text)",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("baize.js missing clipboard behavior %q", want)
		}
	}
}

func TestServeWebUIStreamsAndManagesAttachments(t *testing.T) {
	html := string(indexHTML)
	js := string(baizeJS)
	for _, want := range []string{
		`id="btn-attach"`,
		`aria-haspopup="menu"`,
		`aria-expanded="false"`,
		`id="attachment-menu"`,
		`role="menu"`,
		`id="attachment-image-option"`,
		`role="menuitem"`,
		`id="attachment-image-input" accept="image/png,image/jpeg,image/gif,image/webp" multiple`,
		`id="attachment-file-input"`,
		`id="composer-attachments"`,
		`id="composer-attachment-warning"`,
		`id="attachment-upload-cancel"`,
		`id="attachments-list"`,
		`id="attachments-clear"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("WebUI missing attachment element %q", want)
		}
	}
	metaIndex := strings.Index(html, `id="composer-meta"`)
	attachIndex := strings.Index(html, `id="btn-attach"`)
	taskModeIndex := strings.Index(html, `id="btn-task-mode"`)
	if metaIndex < 0 || attachIndex < metaIndex || taskModeIndex < attachIndex {
		t.Fatalf("attachment button must be in the composer meta bar before the task mode control")
	}
	for _, want := range []string{
		"new FormData()",
		"xhr.upload.onprogress",
		"setRequestHeader('X-Reasonix-Request','attachment-v1')",
		"attachmentXHR?.abort()",
		"function parseAttachmentRefsForDisplay(value)",
		"function renderAttachmentCards(attachments, options={})",
		"function renderDraftAttachments()",
		"imageInputEnabled === false",
		"attachments:attachments.map(a=>a.path)",
		"attachments: editAttachments.map(a => a.path)",
		"closeAttachmentMenu({restoreFocus:true})",
		"e.key==='ArrowDown'||e.key==='ArrowUp'",
		"fetch('/attachments')",
		"post('/attachments/delete'",
		"post('/attachments/clear'",
		"'delivery_checks': 'Delivery checks: {count} total (initial check + {automatic} automatic follow-up checks).'",
		"'delivery_checks': '已完成 {count} 次交付检查（初始检查 + {automatic} 次自动补查）。'",
		"const attempts=Number(e&&e.readiness&&e.readiness.attempts)||0;",
	} {
		if !strings.Contains(js, want) {
			t.Fatalf("WebUI missing streaming attachment behavior %q", want)
		}
	}
}

func TestServeSettingsLayoutAndSaveContract(t *testing.T) {
	html := string(indexHTML)
	js := string(baizeJS)
	css := string(baizeCSS)
	settings := strings.Index(html, `id="settings-form"`)
	visionModel := strings.Index(html, `id="setting-vision-model"`)
	execution := strings.Index(html, `__('settings_execution')`)
	floor := strings.Index(html, `id="quality-floor-select"`)
	appearance := strings.Index(html, `__('settings_appearance')`)
	taskErrors := strings.Index(html, `id="setting-show-task-errors"`)
	attachments := strings.Index(html, `class="settings-group attachment-storage"`)
	if settings < 0 || visionModel < settings || execution < visionModel || floor < execution || appearance < floor || taskErrors < appearance || attachments < taskErrors {
		t.Fatalf("settings controls are not ordered as execution floor, appearance error visibility, attachment storage")
	}
	if strings.Count(html, `id="quality-floor-select"`) != 1 {
		t.Fatal("quality floor must have exactly one settings-only entry")
	}
	actions := strings.Index(html, `class="attachment-storage__actions"`)
	actionsEnd := -1
	if actions >= 0 {
		actionsEnd = strings.Index(html[actions:], `</div>`)
	}
	if actions < attachments || actionsEnd < 0 || !strings.Contains(html[actions:actions+actionsEnd], `id="attachments-refresh"`) || !strings.Contains(html[actions:actions+actionsEnd], `id="attachments-clear"`) {
		t.Fatal("attachment refresh and clear actions must share the storage header")
	}
	for _, marker := range []string{
		`fillVisionModelSetting($('#setting-vision-model'),value.visionModels||[],value.visionModel)`,
		`['defaultModel','visionModel','plannerModel','subagentModel'`,
		`qualityFloorDraft=qualityFloorSelect.value==='delivery'?'delivery':'standard'`,
		`qualityFloorSelect.disabled=floorDisabled;`,
		`if(qualityFloorDraft!==qualityFloor){await requestQualityFloor(qualityFloorDraft)`,
		`await rollbackFloor();`,
		`settingsState(__('settings_conflict'),'warn');return;`,
		`showAppToast(message,tone,pending?6000:3000)`,
		`collapseWorkbench({restoreFocus:true})`,
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("settings save flow missing %q", marker)
		}
	}
	for _, marker := range []string{
		`.attachment-storage__actions{display:flex;`,
		`.settings-check--task-errors strong{white-space:nowrap}`,
		`.app-toast-region{position:fixed;`,
		`.app-toast--success{`,
		`.app-toast--warn{`,
		`.app-toast--danger{`,
	} {
		if !strings.Contains(css, marker) {
			t.Errorf("settings stylesheet missing %q", marker)
		}
	}
	if !strings.Contains(html, `id="app-toast-region" role="status" aria-live="polite" aria-atomic="true"`) {
		t.Fatal("settings save toast region must be a polite atomic live region")
	}
	if strings.Contains(js, `settingsState(message,tone);showNotice(message,tone)`) {
		t.Fatal("settings save result must not be appended to the transcript")
	}
}

func TestServeOperationalFeedbackDoesNotEnterTranscript(t *testing.T) {
	js := string(baizeJS)
	css := string(baizeCSS)
	for _, marker := range []string{
		`APP_TOAST_DURATIONS={info:3000,success:3000,warn:6000,danger:8000}`,
		`toast.setAttribute('role','alert')`,
		`app-toast__close`,
		`showAppToast(error instanceof Error?error.message:__('auth_failed'),'danger',0,'connection')`,
		`es.onopen=()=>{setConnState('connected');clearAppToast('connection')`,
		`focusDecisionPrompt(__('workspace_decision_pending'))`,
		`setDeliveryCardError(card,String(err&&err.message||err))`,
		`showAppToast(__('no_checkpoints'),'warn')`,
		`showAppToast(__('extensions_reloading'),'info')`,
		`showAppToast(__('extensions_reloaded'),'success')`,
		`showAppToast((await response.text()).trim()||__('submit_failed'),'danger')`,
		`showAppToast(__('cannot_delete_active'),'warn')`,
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("operational feedback contract missing %q", marker)
		}
	}
	if strings.Contains(js, `showNotice(`) {
		t.Fatal("page operations must not use the legacy transcript notice helper")
	}
	for _, marker := range []string{
		`appendTranscriptNotice(`,
		`case 'notice': { if(attachAuditNotice(e))`,
		`log.appendChild(el('div','msg--error'`,
		`const it = { id: genItemId(), kind: 'notice', text:`,
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("runtime transcript feedback contract missing %q", marker)
		}
	}
	for _, marker := range []string{
		`.app-toast{display:flex;`,
		`pointer-events:auto`,
		`.app-toast__close{`,
		`.delivery-card__error,.decision-inline-notice{`,
	} {
		if !strings.Contains(css, marker) {
			t.Errorf("operational feedback stylesheet missing %q", marker)
		}
	}
}

func TestServeModelSwitchRefreshesProviderIdentity(t *testing.T) {
	source := string(baizeJS)
	for _, want := range []string{
		"function fetchModelsState()",
		"modelsCache=Array.isArray(data?.models)?data.models:[];",
		"updateModelTrigger();",
		"function refreshModelSelection(){fetchModelsState().catch(()=>{});fetchStatus();fetchEffort();}",
		"function submitModelSwitch(ref)",
		"if(isModelSwitch)refreshModelSelection()",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("Baize model switch missing provider refresh contract %q", want)
		}
	}
	if got := strings.Count(source, "submitModelSwitch("); got != 3 {
		t.Fatalf("submitModelSwitch references = %d, want function plus both model pickers", got)
	}
}

func TestServePDFJSAssetRoutes(t *testing.T) {
	bc := NewBroadcaster()
	ctrl := control.New(control.Options{Sink: bc})
	t.Cleanup(ctrl.Close)
	handler := New(ctrl, bc, config.ServeConfig{}).Handler()

	for _, tc := range []struct {
		path        string
		contentType string
		body        string
	}{
		{path: "/assets/pdfjs/pdf.mjs", contentType: "text/javascript; charset=utf-8", body: "Mozilla Foundation"},
		{path: "/assets/pdfjs/pdf.worker.mjs", contentType: "text/javascript; charset=utf-8", body: "WorkerMessageHandler"},
		{path: "/assets/pdfjs/wasm/openjpeg.wasm", contentType: "application/wasm", body: ""},
		{path: "/assets/pdfjs/cmaps/Adobe-GB1-UCS2.bcmap", contentType: "application/octet-stream", body: ""},
	} {
		t.Run(tc.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want 200", tc.path, recorder.Code)
			}
			if got := recorder.Header().Get("Content-Type"); got != tc.contentType {
				t.Fatalf("GET %s Content-Type = %q, want %q", tc.path, got, tc.contentType)
			}
			if got := recorder.Header().Get("Cache-Control"); got != "public, max-age=86400" {
				t.Fatalf("GET %s Cache-Control = %q", tc.path, got)
			}
			if tc.body != "" && !strings.Contains(recorder.Body.String(), tc.body) {
				t.Fatalf("GET %s missing source marker %q", tc.path, tc.body)
			}
		})
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/assets/pdfjs/missing.mjs", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("missing PDF.js asset status = %d, want 404", recorder.Code)
	}
}

func TestWorkspacePDFPreviewUsesPDFJS(t *testing.T) {
	js := string(baizeJS)
	for _, want := range []string{
		"import('/assets/pdfjs/pdf.mjs')",
		"pdf.worker.mjs",
		"renderWorkspacePDF(preview)",
		"workspace-pdf__pages",
		"updateWorkspacePDFPageFromScroll",
		"Math.abs(view.page-state.page)<=1",
		"releaseWorkspacePDFView(view)",
		"isEvalSupported:false",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("Baize PDF preview missing %q", want)
		}
	}
	if strings.Contains(js, "if(preview.kind==='pdf'){const frame=el('iframe'") {
		t.Fatal("PDF preview still delegates to the browser's iframe PDF viewer")
	}
	css := string(baizeCSS)
	for _, want := range []string{
		`.workspace-pdf__viewport{min-width:0;min-height:0;position:relative;flex:1;overflow:auto;`,
		`.workspace-pdf__pages{width:max-content;min-width:100%;min-height:100%;display:flex;flex-direction:column;`,
	} {
		if !strings.Contains(css, want) {
			t.Errorf("Baize PDF preview stylesheet missing %q", want)
		}
	}
}

func TestSessionListOnlyEnablesOverflowWhenNeeded(t *testing.T) {
	css := string(baizeCSS)
	for _, want := range []string{
		`.timeline-shell{position:relative;min-height:0;flex:1;overflow-y:hidden;`,
		`.timeline-shell--scrollable{overflow-y:auto}`,
		`.session-list{position:relative;z-index:1;display:flex;min-height:0;flex-direction:column;`,
		`.session-item{position:relative;display:grid;`,
	} {
		if !strings.Contains(css, want) {
			t.Errorf("Baize stylesheet missing session list layout contract %q", want)
		}
	}
	js := string(baizeJS)
	for _, want := range []string{
		`contentHeight>shell.clientHeight+1`,
		`shell.classList.toggle('timeline-shell--scrollable',scrollable)`,
		`if(!scrollable&&shell.scrollTop)shell.scrollTop=0`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("Baize session list overflow behavior missing %q", want)
		}
	}
}

func TestReasoningSummaryAlwaysStartsBelowHeader(t *testing.T) {
	css := string(baizeCSS)
	const want = `.reasoning__summary{display:block;width:calc(100% - 8px);box-sizing:border-box;`
	if !strings.Contains(css, want) {
		t.Fatalf("Baize stylesheet missing fixed reasoning summary row %q", want)
	}
}

func TestReasoningDefaultsToCollapsed(t *testing.T) {
	js := string(baizeJS)
	for _, want := range []string{
		"let display='closed';try{display=localStorage.getItem('baize-reasoning-display')||'closed';}",
		"storageValue('baize-reasoning-display','closed')",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("Baize application missing collapsed reasoning default %q", want)
		}
	}
	const closedFirst = `<select id="setting-reasoning-display"><option value="closed">`
	if !strings.Contains(string(indexHTML), closedFirst) {
		t.Errorf("Baize settings UI missing collapsed-first reasoning option %q", closedFirst)
	}
}

func TestToolProgressPreservesCollapsedState(t *testing.T) {
	js := string(baizeJS)
	const want = "body.style.display=card.dataset.open==='true'?'':'none';"
	if !strings.Contains(js, want) {
		t.Fatalf("Baize application missing collapsed tool-progress contract %q", want)
	}
	const unwanted = "body.style.display=''; card.dataset.open='true';"
	if strings.Contains(js, unwanted) {
		t.Fatalf("Baize application still forces tool progress open with %q", unwanted)
	}
}

func TestNestedSubagentToolPreservesParentCollapse(t *testing.T) {
	js := string(baizeJS)
	if strings.Contains(js, "parentCard.dataset.open='true'") {
		t.Fatal("nested tool still forces its parent subagent card open")
	}
	for _, want := range []string{
		"card.dataset.open=isSubagentTool(tool)&&tool.status!=='done'?'true':'false';",
		"parentRoot=nest;",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("Baize application missing subagent nesting contract %q", want)
		}
	}
}

func TestBaizeSessionSubagentAndSettingsUIContracts(t *testing.T) {
	js := string(baizeJS)
	for _, want := range []string{
		"sessionsLoadSequence",
		"if(sequence!==sessionsLoadSequence)return",
		"SUBAGENT_PROGRESS_LIMITS={reasoning:8<<10,text:8<<10,notice:2<<10}",
		"reasonix.subagent.status",
		"subagentAutoCollapse()",
		"fetch('/settings'",
		"method:'PATCH'",
		"baize-theme-preference",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("Baize application missing %q", want)
		}
	}
	html := string(indexHTML)
	for _, want := range []string{`id="btn-settings"`, `id="settings-drawer"`, `name="defaultModel"`, `name="maxSubagentConcurrency"`} {
		if !strings.Contains(html, want) {
			t.Errorf("Baize settings UI missing %q", want)
		}
	}
}

func TestBaizeThemePaletteContracts(t *testing.T) {
	html := string(indexHTML)
	css := string(baizeCSS)
	js := string(baizeJS)
	themes := []string{"charcoal-copper", "ivory-morning"}
	for _, theme := range themes {
		if !strings.Contains(html, `name="appearanceTheme" value="`+theme+`"`) {
			t.Errorf("Baize settings UI missing theme %q", theme)
		}
		if !strings.Contains(css, `:root[data-theme="`+theme+`"]`) {
			t.Errorf("Baize stylesheet missing theme tokens for %q", theme)
		}
	}
	for _, removed := range []string{"night-paper", "paper-workbench", "mist-stone", "sand-apricot"} {
		if strings.Contains(html, `name="appearanceTheme" value="`+removed+`"`) {
			t.Errorf("removed theme %q still appears in the settings UI", removed)
		}
		if strings.Contains(css, `:root[data-theme="`+removed+`"]`) {
			t.Errorf("removed theme %q still has CSS tokens", removed)
		}
		if !strings.Contains(html+js, `'`+removed+`':'`) {
			t.Errorf("removed theme %q is missing its storage migration", removed)
		}
	}
	for _, want := range []string{
		`name="appearanceTheme" value="auto"`,
		`id="setting-theme-picker"`,
		`baize-theme-dark-palette`,
		`baize-theme-light-palette`,
		`preference==='dark'`,
		`preference==='light'`,
		`document.documentElement.dataset.themeFamily`,
		`window.matchMedia('(prefers-color-scheme: light)')`,
	} {
		if !strings.Contains(html+js, want) {
			t.Errorf("Baize theme implementation missing %q", want)
		}
	}
	for _, want := range []string{
		`if(preference==='dark')preference='charcoal-copper'`,
		`family==='light'?'ivory-morning':'charcoal-copper'`,
	} {
		if !strings.Contains(html+js, want) {
			t.Errorf("Baize default theme implementation missing %q", want)
		}
	}
	footer := strings.Index(html, `<footer class="footer">`)
	model := strings.Index(html, `id="btn-composer-model"`)
	footerEnd := -1
	if footer >= 0 {
		footerEnd = strings.Index(html[footer:], `</footer>`)
	}
	if footer < 0 || model < footer || footerEnd < 0 || model > footer+footerEnd {
		t.Fatal("composer model selector moved out of the composer footer")
	}
}

func TestServeMobileGlassComposerContract(t *testing.T) {
	html := string(indexHTML)
	css := string(baizeCSS)
	js := string(baizeJS)

	for _, want := range []string{
		`viewport-fit=cover`,
		`interactive-widget=resizes-content`,
		`id="btn-mobile-controls"`,
		`id="mobile-controls-dialog"`,
		`id="mobile-task-slot"`,
		`id="mobile-approval-slot"`,
		`id="mobile-effort-slot"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("mobile composer HTML missing %q", want)
		}
	}

	for _, want := range []string{
		`--glass-filter:none`,
		`@supports ((-webkit-backdrop-filter:blur(1px)) or (backdrop-filter:blur(1px)))`,
		`height:var(--app-viewport-height,100dvh)`,
		`calc(8px + env(safe-area-inset-bottom))`,
		`.composer-meta{height:44px;min-height:44px;max-height:44px;flex-wrap:nowrap`,
		`.composer-mobile-trigger{display:inline-flex`,
		`.modelsw__menu{position:absolute;inset:auto 0 calc(100% + 8px) auto;width:min(320px,calc(100vw - 24px));max-width:calc(100vw - 68px);max-height:min(42dvh,320px);overscroll-behavior:contain}`,
		`.mobile-controls-dialog::backdrop`,
		`.mobile-controls-dialog--active .task-mode__menu,.mobile-controls-dialog--active .effortsw__menu{position:static;display:grid!important`,
		`.mobile-controls-dialog--active .task-mode__item.is-active,.mobile-controls-dialog--active .effortsw__item.is-active{border-color:`,
		`.mobile-controls-dialog--active .task-mode__item.is-active::after,.mobile-controls-dialog--active .effortsw__item.is-active::after{content:'\2713'`,
		`.mobile-controls-dialog--active .composer-modebar__thumb{box-shadow:`,
		`.mobile-controls-dialog--active .composer-modebar__item.is-active{font-weight:750`,
		`@media(prefers-reduced-transparency:reduce)`,
	} {
		if !strings.Contains(css, want) {
			t.Errorf("mobile composer CSS missing %q", want)
		}
	}

	for _, want := range []string{
		`window.visualViewport?.addEventListener('resize',syncAppViewportHeight`,
		`document.documentElement.style.setProperty('--app-viewport-height'`,
		`slot.appendChild(node)`,
		`marker.parentNode?.insertBefore(node,marker.nextSibling)`,
		`mobileControlsDialog.showModal()`,
		`mobileControlsDialog.addEventListener('keydown',event=>{if(event.key==='Escape')`,
		`mobileControlsDialog.addEventListener('cancel'`,
		`if(!event.matches)closeMobileControls({restoreFocus:false})`,
		`const secondaryDisabled=running||waitingPrompt!==null`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("mobile composer JavaScript missing %q", want)
		}
	}
	if strings.Contains(js, `cloneNode`) {
		t.Fatal("mobile composer controls must move the existing nodes instead of cloning stateful controls")
	}
	if strings.Contains(html+js, `btn-mobile-approval`) || strings.Contains(js, `syncMobileApprovalTrigger`) {
		t.Fatal("mobile composer must expose only the single chat-controls trigger")
	}
	if strings.Contains(css, `.modelsw__menu{position:fixed;`) {
		t.Fatal("mobile model menu must not use fixed positioning inside the glass composer")
	}
}

func TestServeLeftWorkbenchContract(t *testing.T) {
	html := string(indexHTML)
	css := string(baizeCSS)
	js := string(baizeJS)
	for _, want := range []string{
		`class="activity-rail"`, `id="btn-history"`, `data-workbench-mode="files"`,
		`data-workbench-mode="settings"`, `class="context-panel"`, `id="timeline-lines"`,
		`id="session-load-more"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("Serve workbench HTML missing %q", want)
		}
	}
	for _, want := range []string{
		`grid-template-columns:var(--rail-width) var(--context-width) minmax(0,1fr)`,
		`.activity-rail__item--active::before`, `.timeline-lines__travel.is-travelling`,
		`.context-panel[data-view="files"] .workspace-panel`,
		`.context-panel[data-view="settings"] .settings-drawer--open`,
	} {
		if !strings.Contains(css, want) {
			t.Errorf("Serve workbench CSS missing %q", want)
		}
	}
	for _, want := range []string{
		`fetch('/sessions/timeline?'`, `setTimeout(()=>loadSessions(),180)`,
		`function redrawTimeline()`, `loadSessions({append:true})`,
		`contextPanel.append($('#workspace-panel'),$('#settings-drawer'))`,
	} {
		if !strings.Contains(js, want) {
			t.Errorf("Serve workbench JavaScript missing %q", want)
		}
	}
}

func TestServeWorkbenchFollowupLayoutContract(t *testing.T) {
	html := string(indexHTML)
	navStart := strings.Index(html, `<nav class="activity-rail__nav">`)
	navEnd := strings.Index(html, `</nav>`)
	theme := strings.Index(html, `id="btn-theme"`)
	settings := strings.Index(html, `id="btn-settings"`)
	status := strings.Index(html, `id="sidebar-status"`)
	bottom := strings.Index(html, `class="activity-rail__bottom"`)
	if navStart < 0 || navEnd < navStart || strings.Contains(html[navStart:navEnd], `id="btn-settings"`) {
		t.Fatal("settings button must not remain in the primary activity navigation")
	}
	if bottom < 0 || theme < bottom || settings < theme || status < settings {
		t.Fatal("activity rail bottom order must be theme, settings, then status")
	}
	if got := strings.Count(html, ` data-workbench-collapse`); got != 3 {
		t.Fatalf("workbench collapse controls = %d, want 3", got)
	}
	for _, marker := range []string{`id="history-collapse"`, `id="workspace-collapse"`, `id="settings-collapse"`} {
		if !strings.Contains(html, marker) {
			t.Errorf("missing unified workbench collapse control %s", marker)
		}
	}
	for _, removed := range []string{`id="btn-sidebar-collapse"`, `id="workspace-close"`, `id="settings-close"`} {
		if strings.Contains(html, removed) {
			t.Errorf("obsolete close control remains: %s", removed)
		}
	}

	paletteStart := strings.Index(html, `class="theme-picker__palette-grid"`)
	if paletteStart < 0 {
		t.Fatal("settings theme palette grid is missing")
	}
	paletteEnd := strings.Index(html[paletteStart:], `</div>`)
	if paletteEnd < 0 {
		t.Fatal("settings theme palette grid is incomplete")
	}
	palette := html[paletteStart : paletteStart+paletteEnd]
	for _, themeName := range []string{"charcoal-copper", "ivory-morning"} {
		if !strings.Contains(palette, `value="`+themeName+`"`) {
			t.Errorf("theme palette grid missing %s", themeName)
		}
	}
	if auto := strings.Index(html, `value="auto"`); auto < 0 || auto > paletteStart {
		t.Fatal("follow-system theme must remain above the palette grid")
	}
	css := string(baizeCSS)
	for _, marker := range []string{
		`.theme-picker__palette-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr))`,
		`.theme-picker__palette-grid{grid-template-columns:1fr}`,
	} {
		if !strings.Contains(css, marker) {
			t.Errorf("responsive theme palette CSS missing %q", marker)
		}
	}
	js := string(baizeJS)
	for _, marker := range []string{
		`workbenchMode==='files'&&workspaceOpen&&!sidebarCollapsed`,
		`workbenchMode==='settings'&&settingsDrawer.classList.contains('settings-drawer--open')&&!sidebarCollapsed`,
		`element.inert=hidden`,
		`setSurfaceHidden(contextPanel,mobile?!mobileOpen:sidebarCollapsed)`,
		`setSurfaceHidden($('#log'),mobileOpen)`,
		`setSurfaceHidden($('.footer'),mobileOpen)`,
		`menuBtn.setAttribute('aria-expanded',mobileOpen?'true':'false')`,
		`closeSidebar({restoreFocus:true})`,
		`workbenchMode==='files'&&!workspaceOpen||workbenchMode==='settings'&&!settingsDrawer.classList.contains('settings-drawer--open')`,
		`setSurfaceHidden(workspacePanel,true)`,
		`setSurfaceHidden(settingsDrawer,true)`,
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("workbench mode-aware toggle missing %q", marker)
		}
	}
	if strings.Contains(js, `function openSidebar(){workbenchMode='history'`) {
		t.Fatal("mobile sidebar opener must preserve the active files/settings view")
	}
}

func TestServeTimelineFollowupContract(t *testing.T) {
	html := string(indexHTML)
	css := string(baizeCSS)
	js := string(baizeJS)
	for _, marker := range []string{
		`id="timeline-travel-gradient"`,
		`TIMELINE_TRAVEL_MS=520`,
		`TIMELINE_TRAVEL_SPAN=24`,
		`for(let index=0;index<24;index++)`,
		`1-Math.pow(1-progress,3)`,
		`window.matchMedia('(prefers-reduced-motion: reduce)')`,
		`data-selected-pulse`,
		`timelineTravelResolve`,
		`resolve(true)`,
		`setTimeout(reload,TIMELINE_TRAVEL_MS+40)`,
		`if(!response.ok)throw new Error`,
		`baize-timeline-collapsed-dates`,
		`baize-timeline-date-overrides`,
		`JSON.parse(raw)`,
		`entries.every(([,value])=>typeof value==='boolean')`,
		`new Map(legacy.map(value=>[value,true]))`,
		`Object.fromEntries(Array.from(timelineDateOverrides.entries())`,
		`Date.UTC(today.getFullYear(),today.getMonth(),today.getDate())`,
		`return delta>=0&&delta<3`,
		`if(sessionFilter.trim())return false`,
		`if(timelineDateOverrides.has(key))return timelineDateOverrides.get(key)`,
		`return!timelineDateDefaultExpanded(key)`,
		`collapse=!timelineDateCollapsed(key)`,
		`date.type='button'`,
		`date.setAttribute('aria-expanded'`,
		`timeline-group__sessions`,
		`timelineResizeObserver.observe($('#session-timeline'))`,
		`timelineResizeObserver.observe($('#session-list'))`,
		`app.addEventListener('transitionend',settleWorkbenchTransition)`,
		`event.propertyName==='grid-template-columns'`,
	} {
		if !strings.Contains(html+css+js, marker) {
			t.Errorf("timeline follow-up implementation missing %q", marker)
		}
	}
	for _, removed := range []string{
		`for(let i=0;i<=length;i+=3)`,
		`addEventListener('scroll',redrawTimeline`,
		`.session-item__meta`,
	} {
		if strings.Contains(css+js, removed) {
			t.Errorf("obsolete timeline implementation remains: %q", removed)
		}
	}
	renderStart := strings.Index(js, `function renderSessions(){`)
	renderEnd := strings.Index(js, `let sessionsLoadSequence=0;`)
	if renderStart < 0 || renderEnd < renderStart {
		t.Fatal("could not isolate session renderer")
	}
	renderer := js[renderStart:renderEnd]
	for _, marker := range []string{`session-item__title`, `session-item__time`, `session-del`} {
		if !strings.Contains(renderer, marker) {
			t.Errorf("single-line session renderer missing %q", marker)
		}
	}
	if strings.Contains(renderer, `.turns`) || strings.Contains(renderer, `session-item__meta`) {
		t.Fatal("session renderer must not display turns or a second metadata line")
	}
	for _, marker := range []string{
		`.session-item{position:relative;display:grid;grid-template-columns:minmax(0,1fr) auto auto`,
		`.session-item__title{display:block;min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap`,
		`.timeline-lines__travel.is-travelling{filter:drop-shadow`,
		`.timeline-lines__travel.is-positioned{opacity:1}`,
		`.timeline-group__sessions[hidden]{display:none}`,
	} {
		if !strings.Contains(css, marker) {
			t.Errorf("timeline visual contract missing %q", marker)
		}
	}
}
