package serve

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestServeRunFeedbackContract(t *testing.T) {
	js := string(baizeJS)
	for _, marker := range []string{
		`function showTurnFeedback(e)`,
		`if(e.cancelled===true)`,
		`showAppToast(__('turn_stopped'),'info',3000,'turn-feedback')`,
		`showAppToast(__('turn_failed'),'danger',0,'turn-feedback',String(e.err))`,
		`details.appendChild(el('summary','',__('error_details')))`,
		`details.appendChild(el('pre','',detail))`,
		`case 'turn_started': clearAppToast('turn-feedback')`,
		`function resetItems() { clearAppToast('turn-feedback')`,
	} {
		if !strings.Contains(js, marker) {
			t.Errorf("run feedback missing %q", marker)
		}
	}
	if strings.Contains(js, `log.appendChild(el('div','msg--error'`) {
		t.Fatal("turn failures must not be appended to the transcript")
	}
	for _, marker := range []string{`.app-toast__details summary`, `max-height:min(160px,30dvh)`, `overflow-wrap:anywhere`} {
		if !strings.Contains(string(baizeCSS), marker) {
			t.Errorf("bounded diagnostic details missing %q", marker)
		}
	}
}

func TestServeRunFeedbackBehavior(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node.js is required for WebUI behavior tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, node, "--test", "testdata/run_feedback.test.cjs")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("WebUI run feedback behavior: %v\n%s", err, output)
	}
}
