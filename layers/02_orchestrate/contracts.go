package orchestrate

import (
	"time"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
)

// Mode selects execution strategy.
type Mode string

const (
	// ModeChat selects AgentLoop for multi-turn conversation.
	ModeChat Mode = "chat"
	// ModeStructured selects SpecializedLoop for schema-constrained output.
	ModeStructured Mode = "structured"
	// ModePlan selects PlanExecutor for multi-step pipeline.
	ModePlan Mode = "plan"
	// ModeRedundant selects RedundantLoop for N-version voting.
	ModeRedundant Mode = "redundant"
)

// FailureCategory classifies errors for recovery and observability.
type FailureCategory string

const (
	// InferenceFailure covers engine errors and model unavailability.
	InferenceFailure FailureCategory = "inference_failure"
	// ToolFailure covers tool missing, execution error, or timeout.
	ToolFailure FailureCategory = "tool_failure"
	// ConstraintFailure covers grammar, schema, or JSON validity failures.
	ConstraintFailure FailureCategory = "constraint_failure"
	// ValidationFailure covers rule or semantic validation failures.
	ValidationFailure FailureCategory = "validation_failure"
	// OrchestrationFailure covers plan step mismatch or iteration limit.
	OrchestrationFailure FailureCategory = "orchestration_failure"
	// ConfigurationFailure covers missing engine or required schema.
	ConfigurationFailure FailureCategory = "configuration_failure"
	// CancellationFailure covers timeout or context cancellation.
	CancellationFailure FailureCategory = "cancellation"
)

// SystemRequest is input from adapters into the system.
// Adapters construct SystemRequest from external input and pass it to orchestration.
// No adapter may bypass SystemRequest to access internal layers directly.
type SystemRequest struct {
	// Correlation identifiers.
	RequestID string
	SessionID string
	TraceID   string

	// Intent.
	Messages []core.Message
	Mode     Mode

	// Execution tuning.
	Hints ExecutionHints

	// Output expectations.
	Output OutputContract

	// Allowed tool names. Empty means all registered tools.
	Tools []string
}

// SystemResponse is output returned to adapters.
type SystemResponse struct {
	// Correlation (echoed from request).
	RequestID string
	SessionID string

	// Output.
	Content          string
	StructuredOutput any
	ToolCallsMade    []core.ToolCall

	// Quality signals.
	Confidence       *float64
	ConfidenceSource string

	// Usage.
	TokenUsage core.TokenUsage

	// Error is non-nil on failure.
	Error *SystemError
}

// SystemError is the canonical error model.
// All errors map to a FailureCategory from the failure taxonomy.
type SystemError struct {
	Code      string
	Category  FailureCategory
	Retryable bool
	Message   string
	Details   map[string]any
	CausedBy  error
}

// Error implements the error interface.
func (e *SystemError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

// Unwrap returns the wrapped cause for errors.Is/As support.
func (e *SystemError) Unwrap() error {
	return e.CausedBy
}

// ExecutionHints carries optional tuning knobs from adapters.
type ExecutionHints struct {
	MaxTokens   int
	Timeout     time.Duration
	Temperature float64
	TopP        float64
	Options     map[string]any
}

// OutputContract declares schema and grammar expectations for the response.
type OutputContract struct {
	Schema           *core.Schema
	Grammar          string
	RepairAllowed    bool
	StrictValidation bool
}

// PromptProvider builds system prompts for orchestration patterns.
// Implementations live in the prompt layer and satisfy this interface
// without importing orchestrate (structural typing - same pattern as
// ContextProvider and CandidateReranker). Prompt types must implement
// BuildSystemPrompt to participate in this contract.
type PromptProvider interface {
	BuildSystemPrompt() (core.Message, error)
}
