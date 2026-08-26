package serve

import (
	"encoding/json"
	"strings"
	"testing"

	"reasonix/internal/provider"
)

func TestHistoryMessagesCarriesResolvedCapabilityMetadata(t *testing.T) {
	msgs := []provider.Message{{Role: provider.RoleAssistant, ToolCalls: []provider.ToolCall{{
		ID: "call-1", Name: "use_capability",
		Arguments:    `{"action":"call","capability_id":"mcp-tool:ipap/get_case_detail"}`,
		ResolvedName: "mcp__ipap__get_case_detail", CapabilityID: "mcp-tool:ipap/get_case_detail",
	}}}}
	hm := historyMessages(msgs)
	if len(hm) != 1 || len(hm[0].ToolCalls) != 1 {
		t.Fatalf("history tool calls = %+v", hm)
	}
	got := hm[0].ToolCalls[0]
	if got.ResolvedName != "mcp__ipap__get_case_detail" || got.CapabilityID != "mcp-tool:ipap/get_case_detail" {
		t.Fatalf("resolved capability metadata = %+v", got)
	}
	b, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"resolvedName":"mcp__ipap__get_case_detail"`, `"capabilityId":"mcp-tool:ipap/get_case_detail"`} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("history JSON omits %s: %s", want, b)
		}
	}
}

func TestServeIndexFoldsExecutionAuditNoticesIntoToolCards(t *testing.T) {
	html := baizeWebSource()
	for _, want := range []string{
		"function attachAuditNotice(e)",
		"e.code==='capability_proxy'",
		"JSON.parse(String(tool.args||'{}')).capability_id",
		"e.code==='tool_output_truncated'",
		"String(tool.id||'')===toolCallId",
		"String(tool.id||'')===String(receipt.id||'')",
		"function auditEntryText(entry)",
		"function standaloneAuditNotice(e)",
		"if(receipt.kind==='ask'&&receipt.outcome==='answered')return null",
		"if(attachAuditNotice(e)){scrollDown();break;}",
		".card[data-open=\"false\"] .card-audit",
		"'tool_audit': 'Execution audit'",
		"'tool_audit': '运行审计'",
		"'tool_output_truncated': '输出过长，模型仅接收首尾内容；已省略 {elided} / 共 {total}'",
		"'decision_ask_answered': '问题已回答'",
		".notice--audit{",
		`window.matchMedia('(max-width:840px), (hover:none) and (pointer:coarse)')`,
		`let mobileWorkbenchOpen=false`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("serve index missing folded execution audit support %q", want)
		}
	}
}
