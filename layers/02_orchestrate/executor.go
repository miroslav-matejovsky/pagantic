package orchestrate

import (
	"context"
	"fmt"
)

// StepHandler processes one step type. Receives step with Input set,
// returns step with Output set.
type StepHandler func(ctx context.Context, step Step) (Step, error)

// PlanExecutor runs ExecutionPlan steps in order using registered handlers.
// Output of step N is available as Input of step N+1.
type PlanExecutor struct {
	handlers map[StepType]StepHandler
}

// NewPlanExecutor creates executor with given step handlers.
func NewPlanExecutor(handlers map[StepType]StepHandler) *PlanExecutor {
	copied := make(map[StepType]StepHandler, len(handlers))
	for k, v := range handlers {
		copied[k] = v
	}
	return &PlanExecutor{handlers: copied}
}

// Execute runs all steps in plan order. Each step's Output becomes the next
// step's Input. Returns completed steps or error on first failure.
func (pe *PlanExecutor) Execute(ctx context.Context, plan ExecutionPlan) ([]Step, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if pe == nil {
		return nil, fmt.Errorf("plan executor: nil executor")
	}
	if len(plan.Steps) == 0 {
		return []Step{}, nil
	}

	completed := make([]Step, 0, len(plan.Steps))
	for i, step := range plan.Steps {
		if err := ctx.Err(); err != nil {
			return completed, err
		}

		handler, ok := pe.handlers[step.Type]
		if !ok {
			return completed, fmt.Errorf("plan executor: no handler for step type %q (step %d: %s)", step.Type, i, step.Name)
		}

		// Chain: previous step output feeds current step input.
		if i > 0 && step.Input == nil {
			step.Input = completed[i-1].Output
		}

		result, err := handler(ctx, step)
		if err != nil {
			return completed, fmt.Errorf("plan executor: step %d (%s) failed: %w", i, step.Name, err)
		}

		completed = append(completed, result)
	}

	return completed, nil
}
