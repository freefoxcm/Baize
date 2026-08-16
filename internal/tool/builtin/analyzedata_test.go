package builtin

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestAnalyzeDataAggregatesJSON(t *testing.T) {
	args := json.RawMessage(`{
		"input":{"rows":[{"area":"A","count":2},{"area":"B","count":3},{"area":"A","count":5}]},
		"program":"def analyze(data):\n    totals = {}\n    for row in data[\"rows\"]:\n        area = row[\"area\"]\n        totals[area] = totals.get(area, 0) + row[\"count\"]\n    return totals\nresult = analyze(input)"
	}`)
	out, err := (analyzeData{}).Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if out != `{"A":7,"B":3}` {
		t.Fatalf("output = %s", out)
	}
}

func TestAnalyzeDataRejectsCapabilitiesAndInvalidResults(t *testing.T) {
	tests := []struct {
		name    string
		program string
		want    string
	}{
		{name: "load", program: "load(\"os.star\", \"open\")\nresult = 1", want: "module loading is disabled"},
		{name: "file builtin", program: `result = open("secret.txt")`, want: "undefined: open"},
		{name: "missing result", program: `x = input`, want: "must assign"},
		{name: "null result", program: `result = None`, want: "non-null"},
		{name: "empty result", program: `result = {}`, want: "must not be empty"},
		{name: "cyclic result", program: "result = [1]\nresult.append(result)", want: "cyclic list"},
		{name: "function result", program: "def f():\n    return 1\nresult = f", want: "not JSON-compatible"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args, err := json.Marshal(map[string]any{"input": map[string]any{}, "program": tc.program})
			if err != nil {
				t.Fatal(err)
			}
			_, err = (analyzeData{}).Execute(context.Background(), args)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestAnalyzeDataLimitsExecution(t *testing.T) {
	args := json.RawMessage(`{"input":{},"program":"def analyze():\n    x = 0\n    for i in range(2000000):\n        x += i\n    return x\nresult = analyze()"}`)
	_, err := (analyzeData{}).Execute(context.Background(), args)
	if err == nil || !strings.Contains(err.Error(), "too many steps") {
		t.Fatalf("error = %v", err)
	}
}

func TestAnalyzeDataPreservesLargeIntegers(t *testing.T) {
	args := json.RawMessage(`{"input":{"n":9007199254740993},"program":"result = input[\"n\"] + 1"}`)
	out, err := (analyzeData{}).Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if out != `9007199254740994` {
		t.Fatalf("output = %s", out)
	}
}

func TestAnalyzeDataEnforcesEncodedSizeLimits(t *testing.T) {
	t.Run("program", func(t *testing.T) {
		args, err := json.Marshal(map[string]any{
			"input":   map[string]any{},
			"program": strings.Repeat("x", analyzeDataMaxProgramBytes+1),
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = (analyzeData{}).Execute(context.Background(), args)
		if err == nil || !strings.Contains(err.Error(), "program exceeds") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("input", func(t *testing.T) {
		args, err := json.Marshal(map[string]any{
			"input":   strings.Repeat("x", analyzeDataMaxInputBytes),
			"program": "result = 1",
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = (analyzeData{}).Execute(context.Background(), args)
		if err == nil || !strings.Contains(err.Error(), "input exceeds") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("output", func(t *testing.T) {
		args, err := json.Marshal(map[string]any{
			"input":   strings.Repeat("x", analyzeDataMaxOutputBytes),
			"program": "result = input",
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = (analyzeData{}).Execute(context.Background(), args)
		if err == nil || !strings.Contains(err.Error(), "result exceeds") {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestAnalyzeDataHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	args := json.RawMessage(`{"input":{},"program":"result = 1"}`)
	_, err := (analyzeData{}).Execute(ctx, args)
	if err == nil || !strings.Contains(err.Error(), "analysis stopped") {
		t.Fatalf("error = %v", err)
	}
}
