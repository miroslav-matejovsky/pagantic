package orchestrate

import "time"

// LifecycleState represents a position in the execution state machine.
type LifecycleState string

const (
	// StateInit means request received and envelope validated.
	StateInit LifecycleState = "init"
	// StatePlan means creating ExecutionPlan (mode=plan only).
	StatePlan LifecycleState = "plan"
	// StatePrepare means gathering context, tools, schema, grammar.
	StatePrepare LifecycleState = "prepare"
	// StateExecute means running loop, steps, or inference.
	StateExecute LifecycleState = "execute"
	// StateValidate means applying constraints, rules, semantic checks.
	StateValidate LifecycleState = "validate"
	// StateComplete means final response assembled.
	StateComplete LifecycleState = "complete"
	// StateError means terminal failure.
	StateError LifecycleState = "error"
	// StateCancelled means context cancelled or timeout expired.
	StateCancelled LifecycleState = "cancelled"
)

// ExecutionState tracks lifecycle position during request processing.
type ExecutionState struct {
	Current    LifecycleState
	EnteredAt  time.Time
	Attempt    int
	TotalSteps int
	LastError  *SystemError
	LastStep   string
}

// ExecutionContext carries per-request data through the lifecycle.
// Immutable fields are set at init. Mutable fields update during execution.
type ExecutionContext struct {
	// Immutable - set at StateInit.
	RequestID string
	SessionID string
	TraceID   string
	Mode      Mode
	StartedAt time.Time

	// Mutable - updated during execution.
	State         ExecutionState
	SelectedTools []string
	Metadata      map[string]any
}
