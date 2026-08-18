package builtin

import (
	"strings"
	"testing"
)

func TestBashStdinValidation(t *testing.T) {
	input := "bounded input"
	valid := bashParams{Command: "consume", Stdin: &input}
	if err := validateBashParams(valid); err != nil {
		t.Fatalf("bounded foreground stdin: %v", err)
	}
	atLimit := strings.Repeat("x", bashStdinMaxBytes)
	if err := validateBashParams(bashParams{Command: "consume", Stdin: &atLimit}); err != nil {
		t.Fatalf("stdin at byte limit: %v", err)
	}
	empty := ""
	tooLarge := strings.Repeat("x", bashStdinMaxBytes+1)
	tooLargeUTF8 := strings.Repeat("界", bashStdinMaxBytes/3+1)
	for _, params := range []bashParams{
		{Command: "consume", Stdin: &input, RunInBackground: true},
		{Command: "consume", Stdin: &input, PreserveBackgroundProcesses: true},
		{Command: "consume", Stdin: &empty, RunInBackground: true},
		{Command: "consume", Stdin: &tooLarge},
		{Command: "consume", Stdin: &tooLargeUTF8},
	} {
		if err := validateBashParams(params); err == nil {
			t.Fatalf("params should fail: %+v", params)
		}
	}
}

func TestBashSchemaAdvertisesBoundedStdin(t *testing.T) {
	schema := string(bash{}.Schema())
	for _, want := range []string{`"stdin"`, "maximum 2 MiB", "foreground"} {
		if !strings.Contains(schema, want) {
			t.Fatalf("schema missing %q: %s", want, schema)
		}
	}
}
