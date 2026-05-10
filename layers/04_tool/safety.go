package tool

import "time"

// SideEffectLevel classifies what external state a tool touches.
type SideEffectLevel string

const (
	// SideEffectNone means pure computation, no I/O.
	SideEffectNone SideEffectLevel = "none"
	// SideEffectRead means reads external state, no mutations.
	SideEffectRead SideEffectLevel = "read"
	// SideEffectWrite means mutates local state.
	SideEffectWrite SideEffectLevel = "write"
	// SideEffectExternal means calls external services with side effects.
	SideEffectExternal SideEffectLevel = "external"
)

// ToolSafety declares safety metadata for a tool.
// Orchestration uses these fields for retry and isolation decisions.
type ToolSafety struct {
	SideEffects SideEffectLevel
	Idempotent  bool
	Requires    []string // required capabilities: "filesystem", "network", etc.
}

// ToolTimeoutPolicy configures per-tool deadline behavior.
type ToolTimeoutPolicy struct {
	Timeout        time.Duration
	RetryOnTimeout bool
}
