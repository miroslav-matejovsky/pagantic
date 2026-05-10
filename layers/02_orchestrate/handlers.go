package orchestrate

import (
	"context"
	"fmt"

	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
)

// InferHandler builds a StepHandler that runs inference.
// Step.Input should be an inference.Request; Output is *inference.Result.
func InferHandler(engine inference.Engine) StepHandler {
	return func(ctx context.Context, step Step) (Step, error) {
		req, ok := step.Input.(inference.Request)
		if !ok {
			return step, fmt.Errorf("infer handler: input must be inference.Request, got %T", step.Input)
		}

		result, err := engine.Infer(ctx, req)
		if err != nil {
			return step, fmt.Errorf("infer handler: %w", err)
		}

		step.Output = result
		return step, nil
	}
}

// RerankHandler builds a StepHandler that reranks candidates.
// Step.Input should be RerankInput; Output is []RerankCandidate.
func RerankHandler(reranker CandidateReranker) StepHandler {
	return func(ctx context.Context, step Step) (Step, error) {
		input, ok := step.Input.(RerankInput)
		if !ok {
			return step, fmt.Errorf("rerank handler: input must be RerankInput, got %T", step.Input)
		}

		result, err := reranker.Rerank(ctx, input)
		if err != nil {
			return step, fmt.Errorf("rerank handler: %w", err)
		}

		step.Output = result
		return step, nil
	}
}

// RetrieveHandler builds a StepHandler that retrieves context.
// Step.Input should be a query string; Output is []core.Message.
func RetrieveHandler(provider ContextProvider) StepHandler {
	return func(ctx context.Context, step Step) (Step, error) {
		query, ok := step.Input.(string)
		if !ok {
			return step, fmt.Errorf("retrieve handler: input must be string, got %T", step.Input)
		}

		msgs, err := provider.Build(ctx, query)
		if err != nil {
			return step, fmt.Errorf("retrieve handler: %w", err)
		}

		step.Output = msgs
		return step, nil
	}
}

// ValidateHandler builds a StepHandler that validates output using a check
// function. Step.Input should be a string; Output is the same string if valid.
func ValidateHandler(check func(output string) error) StepHandler {
	return func(_ context.Context, step Step) (Step, error) {
		output, ok := step.Input.(string)
		if !ok {
			return step, fmt.Errorf("validate handler: input must be string, got %T", step.Input)
		}

		if err := check(output); err != nil {
			return step, fmt.Errorf("validate handler: %w", err)
		}

		step.Output = output
		return step, nil
	}
}
