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

// RerankCandidate represents a scored item for plan-level reranking.
// Mirrors rerank.Candidate for structural typing compatibility.
type RerankCandidate struct {
	Content  string
	Score    float64
	Source   string
	Metadata map[string]any
}

// RerankInput groups candidates with original query for reranking.
type RerankInput struct {
	Query      string
	Candidates []RerankCandidate
}

// CandidateReranker scores and filters candidates. The rerank.Reranker type
// satisfies this interface when wrapped with an adapter function that converts
// between orchestrate and rerank types. This follows the same structural
// boundary pattern as ContextProvider.
type CandidateReranker interface {
	Rerank(ctx context.Context, input RerankInput) ([]RerankCandidate, error)
}
