package orchestrate

import "context"

// StepType says what kind of work step does.
type StepType string

const (
	// StepInfer runs model inference.
	StepInfer StepType = "infer"
	// StepTool runs tool call.
	StepTool StepType = "tool"
	// StepValidate runs validation.
	StepValidate StepType = "validate"
	// StepRetrieve runs retrieval.
	StepRetrieve StepType = "retrieve"
	// StepRerank runs rerank.
	StepRerank StepType = "rerank"
)

// Step is one unit of work in execution plan.
type Step struct {
	Name   string
	Type   StepType
	Input  any
	Output any
}

// ExecutionPlan is ordered step list.
type ExecutionPlan struct {
	Steps []Step
}

// StepExecutor runs one step.
type StepExecutor interface {
	Execute(ctx context.Context, step Step) (Step, error)
}

// RoutingStrategy chooses how to route or reorder steps.
type RoutingStrategy interface {
	Route(ctx context.Context, plan ExecutionPlan) (ExecutionPlan, error)
}
