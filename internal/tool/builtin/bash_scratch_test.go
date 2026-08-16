package builtin

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"reasonix/internal/effectscope"
	"reasonix/internal/sandbox"
	"reasonix/internal/sessiontemp"
)

func TestScratchBashSpecIsMinimalAndOffline(t *testing.T) {
	b := bash{sb: sandbox.Spec{
		Mode:            "off",
		WriteRoots:      []string{t.TempDir()},
		ForbidReadRoots: []string{"/secret"},
		Network:         true,
	}}
	spec := b.specForParams(context.Background(), bashParams{ExecutionScope: "scratch"})
	if !spec.Enforce() || !spec.MinimalWrites || spec.Network || len(spec.WriteRoots) != 0 {
		t.Fatalf("scratch spec = %+v", spec)
	}
	if len(spec.ForbidReadRoots) != 1 || spec.ForbidReadRoots[0] != "/secret" {
		t.Fatalf("forbid roots = %v", spec.ForbidReadRoots)
	}
}

func TestScratchBashRejectsExpandedLifecycle(t *testing.T) {
	tests := []bashParams{
		{Command: "echo ok", ExecutionScope: "scratch", RunInBackground: true},
		{Command: "echo ok", ExecutionScope: "scratch", PreserveBackgroundProcesses: true},
		{Command: "echo ok", ExecutionScope: "scratch", AdditionalWriteDirs: []string{"/outside"}, Justification: "test"},
	}
	for _, p := range tests {
		if err := validateBashParams(p); err == nil {
			t.Fatalf("params should fail: %+v", p)
		}
	}
}

func TestScratchBashFailsClosedWithoutSandbox(t *testing.T) {
	m := sessiontemp.NewWithRoot(t.TempDir())
	m.Retain()
	defer m.Release()
	b := bash{sessionTemp: m}
	oldCommand := bashSandboxCommand
	bashSandboxCommand = func(sandbox.Spec, sandbox.Shell, string) ([]string, bool) { return nil, false }
	defer func() { bashSandboxCommand = oldCommand }()

	_, _, err := b.prepareLaunch(context.Background(), b.resolved(), bashParams{Command: "echo ok", ExecutionScope: "scratch"}, json.RawMessage(`{}`))
	if err == nil || !strings.Contains(err.Error(), "use analyze_data") {
		t.Fatalf("error = %v", err)
	}
}

func TestScratchBashFailsClosedWithoutSessionTemp(t *testing.T) {
	b := bash{}
	_, _, err := b.prepareLaunch(context.Background(), b.resolved(), bashParams{Command: "echo ok", ExecutionScope: "scratch"}, json.RawMessage(`{}`))
	if err == nil || !strings.Contains(err.Error(), "session-private") || !strings.Contains(err.Error(), "analyze_data") {
		t.Fatalf("error = %v", err)
	}
}

func TestScratchBashRecordsRuntimeProof(t *testing.T) {
	m := sessiontemp.NewWithRoot(t.TempDir())
	m.Retain()
	defer m.Release()
	b := bash{sessionTemp: m}
	oldCommand := bashSandboxCommand
	bashSandboxCommand = func(_ sandbox.Spec, sh sandbox.Shell, command string) ([]string, bool) {
		return unconfinedShellArgv(sh, command), true
	}
	defer func() { bashSandboxCommand = oldCommand }()

	res, err := b.ExecuteDetailed(context.Background(), json.RawMessage(`{"command":"echo ok","execution_scope":"scratch"}`))
	if err != nil {
		t.Fatal(err)
	}
	if res.Execution == nil || res.Execution.EffectScope != effectscope.Scratch {
		t.Fatalf("execution = %+v", res.Execution)
	}
	if !strings.Contains(strings.ToLower(res.Output), "ok") {
		t.Fatalf("output = %q", res.Output)
	}
}
