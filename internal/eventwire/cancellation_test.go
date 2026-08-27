package eventwire

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"reasonix/internal/event"
)

func TestTurnDoneCancellationPreservesErrorAndOmitsFalse(t *testing.T) {
	for _, tc := range []struct {
		name      string
		cancelled bool
		err       error
	}{
		{name: "success"},
		{name: "stopped", cancelled: true},
		{name: "stopped request", cancelled: true, err: errors.New("provider: context canceled")},
		{name: "provider cancellation", err: errors.New("provider: context canceled")},
		{name: "provider failure", err: errors.New("provider: HTTP 500")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wire := ToWire(event.Event{Kind: event.TurnDone, Cancelled: tc.cancelled, Err: tc.err})
			if wire.Cancelled != tc.cancelled {
				t.Fatalf("cancelled = %v, want %v", wire.Cancelled, tc.cancelled)
			}
			if tc.err != nil && wire.Err != tc.err.Error() {
				t.Fatalf("original error was lost: %q", wire.Err)
			}
			data, err := json.Marshal(wire)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(data), `"cancelled":true`) != tc.cancelled || strings.Contains(string(data), `"cancelled":false`) {
				t.Fatalf("unexpected cancellation JSON: %s", data)
			}
		})
	}
}
