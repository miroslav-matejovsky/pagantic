package orchestrate

import (
	"context"
	"time"
)

// Planner creates an ExecutionPlan from a SystemRequest.
// PlanExecutor executes plans - Planner creates them. This separation
// is an explicit architectural boundary.
//
// Implementations range from static plan builders (return hardcoded plans)
// to LLM-based planners that generate plans from request analysis.
type Planner interface {
	CreatePlan(ctx context.Context, req SystemRequest) (ExecutionPlan, error)
}

// PlanPolicy constrains plan construction.
type PlanPolicy struct {
	MaxSteps       int
	AllowedTypes   []StepType
	RequireLinear  bool
	TimeoutPerStep time.Duration
}

// PlanTrace records metadata about how a plan was created.
type PlanTrace struct {
	PlannerID string
	Rationale string
	CreatedAt time.Time
	StepCount int
}
