// Package effectscope names the durable-state boundary proven for one tool run.
package effectscope

// Scope describes whether a successful tool call only observed state, wrote
// disposable host-managed scratch data, or may have changed durable state.
type Scope string

const (
	Observation Scope = "observation"
	Scratch     Scope = "scratch"
	Durable     Scope = "durable"
	Unknown     Scope = "unknown"
)
