// CLI example: pipeline query using built-in StepHandler constructors.
//
// Demonstrates the constructor helpers in layer 2 orchestration:
//
//   - RetrieveHandler wires a ContextProvider into an ExecutionPlan step.
//     It expects a query string and returns []core.Message.
//
//   - RerankHandler wires a CandidateReranker into a plan step. It expects
//     orchestrate.RerankInput and returns []orchestrate.RerankCandidate.
//
//   - InferHandler wires an inference.Engine into a plan step. It expects an
//     inference.Request and returns *inference.Result.
//
//   - ValidateHandler wires a plain Go check function into a plan step. It
//     expects a string and returns the same string when validation passes.
//
// Built-in handlers are intentionally small and strongly typed. Real pipelines
// often need bridge functions between stages because one step's natural output
// is not the exact input type required by the next constructor. This example
// shows that pattern explicitly:
//
//   - retrieve returns []core.Message
//   - rerank needs orchestrate.RerankInput
//   - infer needs inference.Request
//   - validate needs string
//
// Each bridge in this file creates a new StepHandler that converts the step
// input, delegates to the built-in handler constructor, and then passes the
// built-in output forward. This keeps orchestration logic small while still
// using the reusable constructors from layers/02_orchestrate/handlers.go.
//
// Four-stage pipeline:
//
//  1. Retrieve: ContextBuilder pulls relevant Go knowledge into messages.
//  2. Rerank: rerank.Reranker scores retrieved candidates and keeps best ones.
//  3. Infer: the model answers using reranked context and is told to emit JSON.
//  4. Validate: a JSON check ensures the final answer is valid JSON text.
//
// Error handling also shows the canonical orchestrate.SystemError wrapper by
// printing both SystemError.Error() and the unwrapped cause when execution
// fails.
//
// Compare with rerank-query:
//
//   - rerank-query builds custom rerank and infer handlers inline.
//   - pipeline-query uses RetrieveHandler, RerankHandler, InferHandler, and
//     ValidateHandler, with small bridge adapters around them.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/miroslav-matejovsky/pagantic/adapters/cli"
	"github.com/miroslav-matejovsky/pagantic/kronk"
	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	orchestrate "github.com/miroslav-matejovsky/pagantic/layers/02_orchestrate"
	pctx "github.com/miroslav-matejovsky/pagantic/layers/03_context"
	rerank "github.com/miroslav-matejovsky/pagantic/layers/06_rerank"
)

const llmModel = "unsloth/gemma-4-E4B-it"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	prompt, err := cli.ReadPrompt(os.Args[1:], os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	krn, cleanup, err := kronk.Load(ctx, kronk.Config{ModelSource: llmModel})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load engine: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	engine := kronk.NewAdapter(krn, nil)

	retriever := pctx.NewInMemoryRetriever(
		pctx.Document{Content: "Go interfaces are satisfied implicitly. A type implements an interface by having the required methods. No implements keyword is used.", Source: "go-interfaces"},
		pctx.Document{Content: "Go channels are typed conduits for communication between goroutines. Use send and receive operations with the <- operator.", Source: "go-channels"},
		pctx.Document{Content: "Go slices are dynamic views over arrays. append may grow a slice and may allocate a new backing array.", Source: "go-slices"},
		pctx.Document{Content: "Go error handling uses explicit return values. Errors are ordinary values, often returned as the last result.", Source: "go-errors"},
		pctx.Document{Content: "Go structs group fields together. Composition and embedding are preferred over inheritance.", Source: "go-structs"},
		pctx.Document{Content: "Go goroutines are lightweight concurrent functions managed by the Go runtime. Start one with the go keyword.", Source: "go-goroutines"},
		pctx.Document{Content: "Go context carries deadlines, cancellation signals, and request-scoped values across API boundaries.", Source: "go-context"},
		pctx.Document{Content: "defer schedules a function call to run when the surrounding function returns, usually for cleanup.", Source: "go-defer"},
	)

	reranker := &rerank.Reranker{
		Scorer: &rerank.SimpleScorer{},
		Policy: rerank.SelectionPolicy{
			TopK:     3,
			MinScore: 0.1,
		},
	}

	contextProvider := &pctx.ContextBuilder{
		Retriever: retriever,
		MaxChunks: 8,
	}

	rerankAdapter := &rerankBridge{reranker: reranker}
	builtinRerank := orchestrate.RerankHandler(rerankAdapter)
	builtinInfer := orchestrate.InferHandler(engine)
	builtinValidate := orchestrate.ValidateHandler(jsonCheck)

	plan := orchestrate.ExecutionPlan{
		Steps: []orchestrate.Step{
			{Name: "retrieve", Type: orchestrate.StepRetrieve, Input: prompt},
			{Name: "rerank", Type: orchestrate.StepRerank},
			{Name: "infer", Type: orchestrate.StepInfer},
			{Name: "validate", Type: orchestrate.StepValidate},
		},
	}

	executor := orchestrate.NewPlanExecutor(map[orchestrate.StepType]orchestrate.StepHandler{
		orchestrate.StepRetrieve: orchestrate.RetrieveHandler(contextProvider),
		orchestrate.StepRerank:   rerankStepHandler(builtinRerank, prompt),
		orchestrate.StepInfer:    inferStepHandler(builtinInfer, prompt),
		orchestrate.StepValidate: validateStepHandler(builtinValidate),
	})

	results, err := executor.Execute(ctx, plan)
	if err != nil {
		sysErr := &orchestrate.SystemError{
			Code:      "PIPELINE_FAILED",
			Category:  orchestrate.OrchestrationFailure,
			Retryable: false,
			Message:   "pipeline execution failed",
			CausedBy:  err,
		}
		fmt.Fprintf(os.Stderr, "System error: %s\n", sysErr.Error())
		if cause := sysErr.Unwrap(); cause != nil {
			fmt.Fprintf(os.Stderr, "Caused by: %v\n", cause)
		}
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Fprintf(os.Stderr, "Error: empty execution result\n")
		os.Exit(1)
	}

	output, ok := results[len(results)-1].Output.(string)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: final output must be string, got %T\n", results[len(results)-1].Output)
		os.Exit(1)
	}

	fmt.Println(output)
}

// rerankBridge adapts rerank.Reranker to orchestrate.CandidateReranker.
type rerankBridge struct {
	reranker *rerank.Reranker
}

// Rerank converts orchestrate candidates into rerank candidates, delegates to
// layer 6, and converts the result back into plan-level candidates.
func (rb *rerankBridge) Rerank(ctx context.Context, input orchestrate.RerankInput) ([]orchestrate.RerankCandidate, error) {
	candidates := make([]rerank.Candidate, len(input.Candidates))
	for i, c := range input.Candidates {
		candidates[i] = rerank.Candidate{
			Content:  c.Content,
			Score:    c.Score,
			Source:   c.Source,
			Metadata: c.Metadata,
		}
	}

	reranked, err := rb.reranker.Rerank(ctx, rerank.CandidateSet{
		Query:      input.Query,
		Candidates: candidates,
	})
	if err != nil {
		return nil, err
	}

	result := make([]orchestrate.RerankCandidate, len(reranked))
	for i, c := range reranked {
		result[i] = orchestrate.RerankCandidate{
			Content:  c.Content,
			Score:    c.Score,
			Source:   c.Source,
			Metadata: c.Metadata,
		}
	}

	return result, nil
}

// rerankStepHandler adapts retrieved messages into RerankInput and then
// delegates to orchestrate.RerankHandler.
func rerankStepHandler(builtinRerank orchestrate.StepHandler, query string) orchestrate.StepHandler {
	return func(ctx context.Context, step orchestrate.Step) (orchestrate.Step, error) {
		msgs, ok := step.Input.([]core.Message)
		if !ok {
			return step, fmt.Errorf("rerank bridge: expected []core.Message, got %T", step.Input)
		}

		candidates := make([]orchestrate.RerankCandidate, 0, len(msgs))
		for i, msg := range msgs {
			if strings.TrimSpace(msg.Content) == "" {
				continue
			}
			candidates = append(candidates, orchestrate.RerankCandidate{
				Content: msg.Content,
				Source:  fmt.Sprintf("%s-%d", msg.Role, i+1),
				Metadata: map[string]any{
					"role": string(msg.Role),
				},
			})
		}

		step.Input = orchestrate.RerankInput{
			Query:      query,
			Candidates: candidates,
		}

		result, err := builtinRerank(ctx, step)
		if err != nil {
			return step, err
		}

		if reranked, ok := result.Output.([]orchestrate.RerankCandidate); ok {
			fmt.Fprintf(os.Stderr, "Reranked %d -> %d candidates\n", len(candidates), len(reranked))
		}

		return result, nil
	}
}

// inferStepHandler adapts reranked candidates into inference.Request and then
// delegates to orchestrate.InferHandler.
func inferStepHandler(builtinInfer orchestrate.StepHandler, query string) orchestrate.StepHandler {
	return func(ctx context.Context, step orchestrate.Step) (orchestrate.Step, error) {
		candidates, ok := step.Input.([]orchestrate.RerankCandidate)
		if !ok {
			return step, fmt.Errorf("infer bridge: expected []orchestrate.RerankCandidate, got %T", step.Input)
		}

		var contextText strings.Builder
		if len(candidates) == 0 {
			contextText.WriteString("No relevant context was retrieved.")
		} else {
			for i, candidate := range candidates {
				fmt.Fprintf(&contextText, "[%d] source=%s score=%.2f\n%s", i+1, candidate.Source, candidate.Score, candidate.Content)
				if i < len(candidates)-1 {
					contextText.WriteString("\n---\n")
				}
			}
		}

		step.Input = inference.Request{
			Messages: []core.Message{
				core.NewSystemMessage("Answer using only supplied context. Respond in valid JSON with fields answer and sources."),
				core.NewSystemMessage("Reranked context:\n" + contextText.String()),
				core.NewUserMessage(query),
			},
			MaxTokens: 2048,
		}

		result, err := builtinInfer(ctx, step)
		if err != nil {
			return step, err
		}

		if inferResult, ok := result.Output.(*inference.Result); ok {
			fmt.Fprintf(os.Stderr, "Generated %d bytes of output\n", len(inferResult.Content))
		}

		return result, nil
	}
}

// validateStepHandler extracts result content and then delegates to
// orchestrate.ValidateHandler.
func validateStepHandler(builtinValidate orchestrate.StepHandler) orchestrate.StepHandler {
	return func(ctx context.Context, step orchestrate.Step) (orchestrate.Step, error) {
		result, ok := step.Input.(*inference.Result)
		if !ok {
			return step, fmt.Errorf("validate bridge: expected *inference.Result, got %T", step.Input)
		}

		step.Input = result.Content
		validated, err := builtinValidate(ctx, step)
		if err != nil {
			return step, err
		}

		fmt.Fprintf(os.Stderr, "Validation passed\n")
		return validated, nil
	}
}

// jsonCheck verifies that output is valid JSON text.
func jsonCheck(output string) error {
	if !json.Valid([]byte(output)) {
		return fmt.Errorf("output is not valid JSON")
	}
	return nil
}

// Usage:
//
//	go run examples/cli/pipeline-query/main.go "How do Go interfaces work?"
//	go run examples/cli/pipeline-query/main.go "Explain goroutines and channels"
//	echo "What is Go context used for?" | go run examples/cli/pipeline-query/main.go
