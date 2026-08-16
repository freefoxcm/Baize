package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"

	"reasonix/internal/i18n"
	"reasonix/internal/sandbox"
	"reasonix/internal/sessiontemp"
)

// prepareLaunch acquires a session-temp lease, builds the sandboxed argv, and
// applies sandbox-escape approval. The caller releases the returned lease.
func (b bash) prepareLaunch(ctx context.Context, sh sandbox.Shell, p bashParams, rawArgs json.RawMessage) (sandbox.Prepared, *sessiontemp.Lease, error) {
	var lease *sessiontemp.Lease
	sessionDir := ""
	if m := b.sessionTempManager(ctx); m != nil {
		l, err := m.Acquire()
		if err != nil {
			return sandbox.Prepared{}, nil, fmt.Errorf("session temporary directory: %w", err)
		}
		lease = l
		sessionDir = l.Dir()
	}
	if p.ExecutionScope == "scratch" && sessionDir == "" {
		return sandbox.Prepared{}, nil, fmt.Errorf("scratch analysis requires a session-private temporary directory; use analyze_data instead")
	}

	spec := b.specForParams(ctx, p)
	spec.SessionTemp = sessionDir
	argv, wrapped := bashSandboxCommand(spec, sh, p.Command)
	linuxSB := wrapped && sessionDir != "" && runtime.GOOS == "linux"
	prepared := sandbox.Prepared{
		Argv:           argv,
		Wrapped:        wrapped,
		SessionTemp:    sessionDir,
		EnvOverrides:   sandbox.SessionTempEnv(sessionDir, linuxSB),
		LinuxSandboxed: linuxSB,
	}

	if spec.Enforce() && p.ExecutionScope != "scratch" && bashSandboxEscapeSessionAllowed(ctx, p.Command, rawArgs) {
		prepared.Argv = unconfinedShellArgv(sh, p.Command)
		prepared.Wrapped = false
		prepared.LinuxSandboxed = false
		prepared.EnvOverrides = sandbox.SessionTempEnv(sessionDir, false)
	} else if spec.Enforce() && !prepared.Wrapped {
		if p.ExecutionScope == "scratch" {
			lease.Release()
			return sandbox.Prepared{}, nil, fmt.Errorf("scratch analysis requires an available OS sandbox and never runs unconfined; use analyze_data instead (%s)", sandbox.BackendUnavailableReason())
		}
		allow, reason, err := approveBashSandboxEscape(ctx, p.Command, rawArgs, i18n.M.SandboxEscapeWrapReason)
		if err != nil {
			if lease != nil {
				lease.Release()
			}
			return sandbox.Prepared{}, nil, err
		}
		if !allow {
			if lease != nil {
				lease.Release()
			}
			if reason != "" {
				return sandbox.Prepared{}, nil, fmt.Errorf("%s", reason)
			}
			return sandbox.Prepared{}, nil, fmt.Errorf("%s", sandbox.UnavailableMessage())
		}
		prepared.Argv = unconfinedShellArgv(sh, p.Command)
		prepared.Wrapped = false
		prepared.LinuxSandboxed = false
		prepared.EnvOverrides = sandbox.SessionTempEnv(sessionDir, false)
	}
	return prepared, lease, nil
}

func (b bash) sessionTempManager(ctx context.Context) *sessiontemp.Manager {
	if m := sessiontemp.FromContext(ctx); m != nil {
		return m
	}
	return b.sessionTemp
}
