package serve

import (
	"strings"
	"testing"
)

func TestServeInteractionWidthContract(t *testing.T) {
	css := string(baizeCSS)
	for _, want := range []string{
		`--chat-maxw:960px;--todos-collapsed-maxw:760px`,
		`.transcript>:not(.welcome){max-width:var(--chat-maxw);margin-inline:auto}`,
		`.approval-slot{width:100%;max-width:var(--chat-maxw);margin:0 auto}`,
		`.approval{width:100%;max-width:var(--chat-maxw);margin:0 auto;background:var(--interaction-glass);`,
		`.ask-slot{width:100%;max-width:var(--chat-maxw);margin:0 auto}`,
		`.todos{width:100%;max-width:var(--chat-maxw);margin:0 auto;`,
		`.todos--collapsed{max-width:var(--todos-collapsed-maxw)}`,
		`.composer-card{position:relative;width:100%;max-width:var(--chat-maxw);`,
		`.footer{grid-column:3;grid-row:1;align-self:end;`, `gap:6px;position:relative;pointer-events:none`,
		`--interaction-glass:rgba(44,41,38,.68)`, `--interaction-glass:rgba(255,250,242,.74)`,
		`--interaction-glass-filter:blur(16px)`, `--interaction-glass-filter:none`,
		`.todos--collapsed .todos__head{min-height:32px;padding:4px 10px}`,
	} {
		if !strings.Contains(css, want) {
			t.Errorf("Serve interaction width CSS missing %q", want)
		}
	}
}

func TestServeTodoStatusIconContract(t *testing.T) {
	css := string(baizeCSS)
	js := string(baizeJS)
	for _, want := range []string{
		`.todos__item{display:grid;grid-template-columns:18px minmax(0,1fr);`,
		`.todos__status{display:grid;place-items:center;width:18px;height:18px;`,
		`border-radius:50%;background:color-mix(in srgb,var(--panel-2) 58%,transparent)`,
		`.todos__status svg{display:block;width:12px;height:12px;`,
		`.todos__status--in_progress{border-color:color-mix(in srgb,var(--accent) 24%,transparent);background:var(--accent-soft);color:var(--accent)}`,
		`.todos__status--completed{border-color:color-mix(in srgb,var(--success) 24%,transparent);background:var(--success-soft);color:var(--success)}`,
		`function todoStatusIcon(s)`,
		`class="todos__status-ring"`,
		`class="todos__status-play"`,
		`class="todos__status-check"`,
		`role="img" aria-label="'+escHtml(statusText)+'" title="'+escHtml(statusText)+'"`,
		`'todo_pending': '待办'`,
		`'todo_completed': '已完成'`,
	} {
		if !strings.Contains(css+js, want) {
			t.Errorf("Serve todo status icon contract missing %q", want)
		}
	}
	for _, unwanted := range []string{`--composer-maxw`, `default:return '○'`, `.todos__status-dot`, `.todos__status{display:inline-flex;align-items:center;justify-content:center;padding:1px 7px`, `.todos__status-disc`} {
		if strings.Contains(css+js, unwanted) {
			t.Errorf("Serve todo status still contains legacy pill marker %q", unwanted)
		}
	}
}
