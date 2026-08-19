package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"reasonix/internal/capability"
	"reasonix/internal/tool"
)

type argumentSchemaMCPTool struct {
	schema json.RawMessage
}

func (argumentSchemaMCPTool) Name() string        { return "mcp__ipap__get_case_detail" }
func (argumentSchemaMCPTool) Description() string { return "Get case detail" }
func (t argumentSchemaMCPTool) Schema() json.RawMessage {
	return t.schema
}
func (argumentSchemaMCPTool) Execute(context.Context, json.RawMessage) (string, error) {
	return "", nil
}
func (argumentSchemaMCPTool) ReadOnly() bool         { return true }
func (argumentSchemaMCPTool) MCPServerName() string  { return "ipap" }
func (argumentSchemaMCPTool) MCPRawToolName() string { return "get_case_detail" }

func TestUseCapabilityListSurfacesMCPArgumentKeys(t *testing.T) {
	reg := tool.NewRegistry()
	reg.Add(argumentSchemaMCPTool{schema: json.RawMessage(`{
		"type":"object",
		"properties":{"include":{"type":"array"},"case_no":{"type":"string"}},
		"required":["case_no"]
	}`)})
	proxy := NewUseCapabilityTool(context.Background(), nil, nil, reg, nil, nil, func() capability.Catalog {
		return capability.Catalog{Entries: []capability.Entry{{
			ID:          "mcp-tool:ipap/get_case_detail",
			Kind:        capability.KindMCPTool,
			Name:        "get_case_detail",
			Description: "Get case detail",
			Status:      capability.StatusReady,
			ReadOnly:    true,
		}}}
	})

	listed, err := proxy.Execute(context.Background(), json.RawMessage(`{"action":"list"}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"argument_keys": [`,
		`"case_no"`,
		`"include"`,
		`"required_argument_keys": [`,
		"Never infer argument names",
	} {
		if !strings.Contains(listed, want) {
			t.Fatalf("list result missing %q:\n%s", want, listed)
		}
	}
	if strings.Contains(listed, "case_number") {
		t.Fatalf("list result invented an argument alias:\n%s", listed)
	}
}
