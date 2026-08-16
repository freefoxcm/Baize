package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"sort"
	"strings"
	"time"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"

	"reasonix/internal/tool"
)

const (
	analyzeDataMaxProgramBytes = 32 << 10
	analyzeDataMaxInputBytes   = 2 << 20
	analyzeDataMaxOutputBytes  = 1 << 20
	analyzeDataMaxSteps        = 1_000_000
	analyzeDataMaxDepth        = 256
	analyzeDataTimeout         = 5 * time.Second
)

func init() { tool.RegisterBuiltin(analyzeData{}) }

type analyzeData struct{}

type analyzeDataParams struct {
	Program string          `json:"program"`
	Input   json.RawMessage `json:"input"`
}

func (analyzeData) Name() string { return "analyze_data" }

func (analyzeData) Description() string {
	return "Run deterministic, side-effect-free data calculations over JSON. Use MCP/read tools to obtain facts first, then pass their JSON here for grouping, aggregation, intersections, rankings, or trends without creating temporary files. The Starlark program receives `input` and must assign its JSON-compatible answer to global `result`; put loops inside a function and call it when assigning result. It has no file, environment, clock, random, network, module-loading, or subprocess access. Prefer this tool over writing analysis scripts with bash."
}

func (analyzeData) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"program":{"type":"string","maxLength":32768,"description":"Starlark source. It receives global input and must assign a JSON-compatible value to global result."},"input":{"description":"JSON data to analyze (maximum encoded size 2 MiB)."}},"required":["program","input"]}`)
}

func (analyzeData) ReadOnly() bool { return true }

func (analyzeData) PlanModeSafe() bool { return true }

func (analyzeData) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p analyzeDataParams
	dec := json.NewDecoder(bytes.NewReader(args))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	if len(p.Program) > analyzeDataMaxProgramBytes {
		return "", fmt.Errorf("program exceeds %d bytes", analyzeDataMaxProgramBytes)
	}
	p.Program = strings.TrimSpace(p.Program)
	if p.Program == "" {
		return "", fmt.Errorf("program is required")
	}
	if len(p.Input) == 0 {
		return "", fmt.Errorf("input is required")
	}
	if len(p.Input) > analyzeDataMaxInputBytes {
		return "", fmt.Errorf("input exceeds %d bytes", analyzeDataMaxInputBytes)
	}

	input, err := decodeAnalyzeJSON(p.Input)
	if err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	runCtx, cancel := context.WithTimeout(ctx, analyzeDataTimeout)
	defer cancel()
	thread := &starlark.Thread{Name: "analyze_data"}
	thread.SetMaxExecutionSteps(analyzeDataMaxSteps)
	thread.Load = func(_ *starlark.Thread, module string) (starlark.StringDict, error) {
		return nil, fmt.Errorf("module loading is disabled: %s", module)
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-runCtx.Done():
			thread.Cancel(runCtx.Err().Error())
		case <-done:
		}
	}()
	globals, execErr := starlark.ExecFileOptions(&syntax.FileOptions{}, thread, "analyze_data.star", p.Program, starlark.StringDict{"input": input})
	close(done)
	if runCtx.Err() != nil {
		return "", fmt.Errorf("analysis stopped: %w", runCtx.Err())
	}
	if execErr != nil {
		return "", fmt.Errorf("analysis failed: %w", execErr)
	}
	result, ok := globals["result"]
	if !ok || result == starlark.None {
		return "", fmt.Errorf("program must assign a non-null value to global result")
	}
	if analyzeResultEmpty(result) {
		return "", fmt.Errorf("global result must not be empty")
	}
	plain, err := analyzeStarlarkToJSON(result)
	if err != nil {
		return "", fmt.Errorf("result is not JSON-compatible: %w", err)
	}
	out, err := json.Marshal(plain)
	if err != nil {
		return "", fmt.Errorf("encode result: %w", err)
	}
	if len(out) > analyzeDataMaxOutputBytes {
		return "", fmt.Errorf("result exceeds %d bytes", analyzeDataMaxOutputBytes)
	}
	return string(out), nil
}

func analyzeResultEmpty(value starlark.Value) bool {
	switch value := value.(type) {
	case starlark.String:
		return len(value) == 0
	case *starlark.List:
		return value.Len() == 0
	case starlark.Tuple:
		return len(value) == 0
	case *starlark.Dict:
		return value.Len() == 0
	default:
		return false
	}
}

func decodeAnalyzeJSON(raw json.RawMessage) (starlark.Value, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("multiple JSON values")
		}
		return nil, err
	}
	return analyzeJSONToStarlark(value)
}

func analyzeJSONToStarlark(value any) (starlark.Value, error) {
	return analyzeJSONToStarlarkDepth(value, 0)
}

func analyzeJSONToStarlarkDepth(value any, depth int) (starlark.Value, error) {
	if depth > analyzeDataMaxDepth {
		return nil, fmt.Errorf("JSON nesting exceeds %d levels", analyzeDataMaxDepth)
	}
	switch value := value.(type) {
	case nil:
		return starlark.None, nil
	case bool:
		return starlark.Bool(value), nil
	case string:
		return starlark.String(value), nil
	case json.Number:
		if strings.ContainsAny(value.String(), ".eE") {
			f, err := value.Float64()
			if err != nil || math.IsInf(f, 0) || math.IsNaN(f) {
				return nil, fmt.Errorf("invalid number %q", value)
			}
			return starlark.Float(f), nil
		}
		integer := new(big.Int)
		if _, ok := integer.SetString(value.String(), 10); !ok {
			return nil, fmt.Errorf("invalid integer %q", value)
		}
		return starlark.MakeBigInt(integer), nil
	case []any:
		items := make([]starlark.Value, 0, len(value))
		for _, item := range value {
			converted, err := analyzeJSONToStarlarkDepth(item, depth+1)
			if err != nil {
				return nil, err
			}
			items = append(items, converted)
		}
		return starlark.NewList(items), nil
	case map[string]any:
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		dict := starlark.NewDict(len(keys))
		for _, key := range keys {
			converted, err := analyzeJSONToStarlarkDepth(value[key], depth+1)
			if err != nil {
				return nil, err
			}
			if err := dict.SetKey(starlark.String(key), converted); err != nil {
				return nil, err
			}
		}
		return dict, nil
	default:
		return nil, fmt.Errorf("unsupported JSON value %T", value)
	}
}

func analyzeStarlarkToJSON(value starlark.Value) (any, error) {
	return (&analyzeConversionState{
		lists: make(map[*starlark.List]bool),
		dicts: make(map[*starlark.Dict]bool),
	}).convert(value, 0)
}

type analyzeConversionState struct {
	lists map[*starlark.List]bool
	dicts map[*starlark.Dict]bool
}

func (state *analyzeConversionState) convert(value starlark.Value, depth int) (any, error) {
	if depth > analyzeDataMaxDepth {
		return nil, fmt.Errorf("result nesting exceeds %d levels", analyzeDataMaxDepth)
	}
	switch value := value.(type) {
	case starlark.NoneType:
		return nil, nil
	case starlark.Bool:
		return bool(value), nil
	case starlark.String:
		return string(value), nil
	case starlark.Int:
		return json.Number(value.String()), nil
	case starlark.Float:
		f := float64(value)
		if math.IsInf(f, 0) || math.IsNaN(f) {
			return nil, fmt.Errorf("non-finite float")
		}
		return f, nil
	case *starlark.List:
		if state.lists[value] {
			return nil, fmt.Errorf("cyclic list")
		}
		state.lists[value] = true
		defer delete(state.lists, value)
		out := make([]any, 0, value.Len())
		iter := value.Iterate()
		defer iter.Done()
		var item starlark.Value
		for iter.Next(&item) {
			converted, err := state.convert(item, depth+1)
			if err != nil {
				return nil, err
			}
			out = append(out, converted)
		}
		return out, nil
	case starlark.Tuple:
		out := make([]any, 0, len(value))
		for _, item := range value {
			converted, err := state.convert(item, depth+1)
			if err != nil {
				return nil, err
			}
			out = append(out, converted)
		}
		return out, nil
	case *starlark.Dict:
		if state.dicts[value] {
			return nil, fmt.Errorf("cyclic dictionary")
		}
		state.dicts[value] = true
		defer delete(state.dicts, value)
		out := make(map[string]any, value.Len())
		for _, item := range value.Items() {
			key, ok := starlark.AsString(item[0])
			if !ok {
				return nil, fmt.Errorf("object key %s is not a string", item[0].Type())
			}
			converted, err := state.convert(item[1], depth+1)
			if err != nil {
				return nil, err
			}
			out[key] = converted
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported Starlark value %s", value.Type())
	}
}
