package agent

import (
	"context"
	"encoding/json"
	"testing"

	"reasonix/internal/effectscope"
	"reasonix/internal/event"
	"reasonix/internal/provider"
	"reasonix/internal/tool"
)

type scratchBash struct{}

func (scratchBash) Name() string        { return "bash" }
func (scratchBash) Description() string { return "scratch bash" }
func (scratchBash) ReadOnly() bool      { return false }
func (scratchBash) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"command":{"type":"string"},"stdin":{"type":"string"},"execution_scope":{"type":"string"}}}`)
}
func (scratchBash) Execute(context.Context, json.RawMessage) (string, error) {
	return `{"rows":3,"total":42}`, nil
}
func (scratchBash) ExecutionDescriptor(json.RawMessage) *tool.ShellExecution {
	return &tool.ShellExecution{Kind: "shell"}
}
func (scratchBash) ExecuteDetailed(context.Context, json.RawMessage) (tool.DetailedResult, error) {
	return tool.DetailedResult{
		Output: `{"rows":3,"total":42}`,
		Execution: &tool.ShellExecution{
			Kind:         "shell",
			State:        tool.ShellStateCompleted,
			MutationRisk: tool.ShellMutationNone,
			EffectScope:  effectscope.Scratch,
		},
	}, nil
}

func TestE2EPureScratchAnalysisCompletesWithoutWorkspaceDeliveryDuties(t *testing.T) {
	reg := tool.NewRegistry()
	reg.Add(scratchBash{})
	prov := &scriptedProvider{name: "scratch-analysis", turns: [][]provider.Chunk{
		{toolCallChunk("analyze", "bash", `{"command":"python -B -E analyze_dataset.py --request -","stdin":"{\"action\":\"inspect\"}","execution_scope":"scratch"}`), {Type: provider.ChunkDone}},
		{{Type: provider.ChunkText, Text: "分析完成，共 3 行。"}, {Type: provider.ChunkDone}},
	}}
	a := New(prov, reg, NewSession(""), Options{}, event.Discard)
	if err := a.Run(withClosedLoopContext(context.Background()), "分析页面中的表格，不导出文件"); err != nil {
		t.Fatalf("pure scratch analysis should finish directly: %v", err)
	}
}

func TestE2EArtifactDeliveryCompletesAfterManifestReadAndTrustedVerification(t *testing.T) {
	reg := evidenceRegistry()
	reg.Add(fakeReadFileTool{})
	verify := "python -B -E -m unittest discover -s .reasonix/skills/_shared -p test_delivery_artifacts.py"
	fullVerify := "python -m pytest"
	prov := &scriptedProvider{name: "artifact-delivery", turns: [][]provider.Chunk{
		{toolCallChunk("criteria", "todo_write", `{"todos":[{"content":"生成报告","status":"in_progress"}]}`), {Type: provider.ChunkDone}},
		{toolCallChunk("spec", "write_file", `{"path":"reports/demo/report-spec.json"}`), {Type: provider.ChunkDone}},
		{toolCallChunk("render", "bash", `{"command":"python render_report.py"}`), {Type: provider.ChunkDone}},
		{toolCallChunk("read-spec", "read_file", `{"path":"reports/demo/report-spec.json"}`), {Type: provider.ChunkDone}},
		{toolCallChunk("read-manifest", "read_file", `{"path":"reports/demo/artifact-manifest.json"}`), {Type: provider.ChunkDone}},
		{toolCallChunk("verify", "bash", `{"command":"`+verify+`","stdin":"{\"manifest\":\"reports/demo/artifact-manifest.json\"}","execution_scope":"scratch"}`), {Type: provider.ChunkDone}},
		{toolCallChunk("full-verify", "bash", `{"command":"`+fullVerify+`","execution_scope":"scratch"}`), {Type: provider.ChunkDone}},
		{toolCallChunk("signoff", "complete_step", `{"step":"生成报告","result":"报告已复读并验证","evidence":[{"kind":"verification","summary":"manifest 与实际产物一致","command":"`+verify+`"},{"kind":"verification","summary":"完整验证通过","command":"`+fullVerify+`"},{"kind":"files","summary":"已复读最终规范和 manifest","paths":["reports/demo/report-spec.json","reports/demo/artifact-manifest.json"]}]}`), {Type: provider.ChunkDone}},
		{{Type: provider.ChunkText, Text: "报告已生成并验证。"}, {Type: provider.ChunkDone}},
	}}
	a := New(prov, reg, NewSession(""), Options{}, event.Discard)
	if err := a.Run(withClosedLoopContext(context.Background()), "生成并交付报告"); err != nil {
		t.Fatalf("artifact delivery should pass after readback and verification: %v", err)
	}
}
