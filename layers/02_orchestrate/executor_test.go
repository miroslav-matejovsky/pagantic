package orchestrate

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlanExecutor_EmptyPlan(t *testing.T) {
	pe := NewPlanExecutor(nil)
	results, err := pe.Execute(context.Background(), ExecutionPlan{})
	require.NoError(t, err)
	require.Empty(t, results)
}

func TestPlanExecutor_NilExecutor(t *testing.T) {
	var pe *PlanExecutor
	_, err := pe.Execute(context.Background(), ExecutionPlan{Steps: []Step{{Name: "x"}}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "nil executor")
}

func TestPlanExecutor_MissingHandler(t *testing.T) {
	pe := NewPlanExecutor(nil)
	plan := ExecutionPlan{Steps: []Step{{Name: "step1", Type: StepInfer}}}
	_, err := pe.Execute(context.Background(), plan)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no handler")
}

func TestPlanExecutor_SingleStep(t *testing.T) {
	handler := func(_ context.Context, step Step) (Step, error) {
		step.Output = "result"
		return step, nil
	}

	pe := NewPlanExecutor(map[StepType]StepHandler{
		StepInfer: handler,
	})

	plan := ExecutionPlan{Steps: []Step{
		{Name: "infer1", Type: StepInfer, Input: "prompt"},
	}}

	results, err := pe.Execute(context.Background(), plan)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "result", results[0].Output)
}

func TestPlanExecutor_ChainedSteps(t *testing.T) {
	retrieve := func(_ context.Context, step Step) (Step, error) {
		step.Output = []string{"doc1", "doc2"}
		return step, nil
	}
	rerank := func(_ context.Context, step Step) (Step, error) {
		docs, ok := step.Input.([]string)
		require.True(t, ok)
		step.Output = docs[:1] // keep best
		return step, nil
	}

	pe := NewPlanExecutor(map[StepType]StepHandler{
		StepRetrieve: retrieve,
		StepRerank:   rerank,
	})

	plan := ExecutionPlan{Steps: []Step{
		{Name: "fetch-docs", Type: StepRetrieve, Input: "query"},
		{Name: "rerank-docs", Type: StepRerank},
	}}

	results, err := pe.Execute(context.Background(), plan)
	require.NoError(t, err)
	require.Len(t, results, 2)
	require.Equal(t, []string{"doc1"}, results[1].Output)
}

func TestPlanExecutor_StepFailure(t *testing.T) {
	ok := func(_ context.Context, step Step) (Step, error) {
		step.Output = "fine"
		return step, nil
	}
	fail := func(_ context.Context, _ Step) (Step, error) {
		return Step{}, fmt.Errorf("boom")
	}

	pe := NewPlanExecutor(map[StepType]StepHandler{
		StepRetrieve: ok,
		StepRerank:   fail,
	})

	plan := ExecutionPlan{Steps: []Step{
		{Name: "retrieve", Type: StepRetrieve},
		{Name: "rerank", Type: StepRerank},
	}}

	results, err := pe.Execute(context.Background(), plan)
	require.Error(t, err)
	require.Contains(t, err.Error(), "boom")
	require.Len(t, results, 1) // first step completed
}

func TestPlanExecutor_CancelledContext(t *testing.T) {
	handler := func(_ context.Context, step Step) (Step, error) {
		step.Output = "done"
		return step, nil
	}

	pe := NewPlanExecutor(map[StepType]StepHandler{
		StepInfer: handler,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	plan := ExecutionPlan{Steps: []Step{{Name: "infer", Type: StepInfer}}}
	_, err := pe.Execute(ctx, plan)
	require.Error(t, err)
}

func TestPlanExecutor_ExplicitInput(t *testing.T) {
	handler := func(_ context.Context, step Step) (Step, error) {
		step.Output = step.Input
		return step, nil
	}

	pe := NewPlanExecutor(map[StepType]StepHandler{
		StepInfer:    handler,
		StepValidate: handler,
	})

	plan := ExecutionPlan{Steps: []Step{
		{Name: "infer", Type: StepInfer, Input: "first"},
		{Name: "validate", Type: StepValidate, Input: "explicit"}, // explicit overrides chain
	}}

	results, err := pe.Execute(context.Background(), plan)
	require.NoError(t, err)
	require.Len(t, results, 2)
	require.Equal(t, "explicit", results[1].Output) // used explicit Input, not chained
}
