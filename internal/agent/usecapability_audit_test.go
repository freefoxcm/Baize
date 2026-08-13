package agent

import (
	"testing"

	"reasonix/internal/event"
	"reasonix/internal/tool"
)

func TestEmitProxyAuditCarriesStableCodeAndCapability(t *testing.T) {
	var got event.Event
	EmitProxyAudit(event.FuncSink(func(e event.Event) { got = e }), tool.ResolvedCall{
		DisplayName:  "use_capability",
		TargetName:   "mcp__ipap__get_case_detail",
		CapabilityID: "mcp-tool:ipap/get_case_detail",
	})
	if got.Kind != event.Notice || got.Code != event.NoticeCodeCapabilityProxy || got.Detail != "mcp-tool:ipap/get_case_detail" {
		t.Fatalf("proxy audit = %+v", got)
	}
}
