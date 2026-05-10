// CLI example: query with document reranking via ExecutionPlan.
//
// Demonstrates the rerank layer wired into orchestrate's ExecutionPlan:
//
//   - Layer 6 (rerank): Reranker combines a RelevanceScorer with a
//     SelectionPolicy to score, filter, and sort candidates. SimpleScorer
//     uses keyword overlap (good for demos, not production). SelectionPolicy
//     controls TopK and MinScore thresholds.
//
//   - Layer 3 (context): InMemoryRetriever finds relevant documents by
//     keyword matching. Documents are scored by the retriever first, then
//     reranked by the rerank layer for higher precision.
//
//   - Layer 2 (orchestrate): PlanExecutor runs an ExecutionPlan - a sequence
//     of typed steps. Each step has a handler. Output of step N flows into
//     input of step N+1. This example uses Retrieve -> Rerank -> Infer
//     pipeline. CandidateReranker interface lets orchestrate use rerank
//     without importing it (structural typing pattern).
//
//   - Layer 1 (inference): Engine runs the final inference with reranked
//     context injected as system messages.
//
// Two-stage retrieval pattern:
//
//  1. Retrieve: fast, broad recall (keyword matching)
//  2. Rerank: precise, narrow selection (scoring + filtering)
//
// This is standard in production RAG pipelines. The retriever casts a wide
// net; the reranker picks the best catches.
//
// Key pagantic concept: ExecutionPlan makes multi-step pipelines explicit
// and composable. Steps are typed (infer, tool, validate, retrieve, rerank),
// handlers are pluggable, and output chaining is automatic.
package main

import (
	"context"
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

	// Domain knowledge documents for retrieval.
	retriever := pctx.NewInMemoryRetriever(
		pctx.Document{Content: "Go interfaces are satisfied implicitly. No implements keyword needed. A type satisfies an interface by having all its methods.", Source: "go-interfaces"},
		pctx.Document{Content: "Go channels are typed conduits for goroutine communication. Send and receive with the <- operator.", Source: "go-channels"},
		pctx.Document{Content: "Go slices are dynamic arrays backed by a fixed-size array. Use append to grow them.", Source: "go-slices"},
		pctx.Document{Content: "Go error handling uses explicit return values. The error interface has one method: Error() string.", Source: "go-errors"},
		pctx.Document{Content: "Go structs group fields together. Embedding composes types without inheritance.", Source: "go-structs"},
		pctx.Document{Content: "Go goroutines are lightweight threads managed by the Go runtime. Start with go keyword.", Source: "go-goroutines"},
		pctx.Document{Content: "Go context carries deadlines, cancellation signals, and request-scoped values across API boundaries.", Source: "go-context"},
		pctx.Document{Content: "Go defer pushes a function call onto a stack that executes after the surrounding function returns.", Source: "go-defer"},
	)

	// Reranker with keyword scoring and selection policy.
	reranker := &rerank.Reranker{
		Scorer: &rerank.SimpleScorer{},
		Policy: rerank.SelectionPolicy{
			TopK:     3,   // keep top 3 documents
			MinScore: 0.1, // require at least 10% keyword overlap
		},
	}

	// Build the execution plan: Retrieve -> Rerank -> Infer.
	// ContextProvider bridges retriever to orchestrate.
	contextProvider := &pctx.ContextBuilder{
		Retriever: retriever,
		MaxChunks: 8, // retrieve broadly
	}

	// Adapter wraps rerank.Reranker to satisfy orchestrate.CandidateReranker.
	rerankAdapter := &rerankBridge{reranker: reranker}

	plan := orchestrate.ExecutionPlan{
		Steps: []orchestrate.Step{
			{Name: "retrieve", Type: orchestrate.StepRetrieve, Input: prompt},
			{Name: "rerank", Type: orchestrate.StepRerank},
			{Name: "infer", Type: orchestrate.StepInfer},
		},
	}

	executor := orchestrate.NewPlanExecutor(map[orchestrate.StepType]orchestrate.StepHandler{
		orchestrate.StepRetrieve: orchestrate.RetrieveHandler(contextProvider),
		orchestrate.StepRerank:   rerankStepHandler(rerankAdapter, prompt),
		orchestrate.StepInfer:    inferWithContextHandler(engine, prompt),
	})

	results, err := executor.Execute(ctx, plan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print the final inference result.
	if len(results) > 0 {
		last := results[len(results)-1]
		if result, ok := last.Output.(*inference.Result); ok {
			fmt.Println(result.Content)
		}
	}
}

// rerankBridge adapts rerank.Reranker to orchestrate.CandidateReranker.
type rerankBridge struct {
	reranker *rerank.Reranker
}

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

// rerankStepHandler converts retrieved context messages into rerank candidates,
// reranks them, and outputs reranked candidates.
func rerankStepHandler(reranker orchestrate.CandidateReranker, query string) orchestrate.StepHandler {
	return func(ctx context.Context, step orchestrate.Step) (orchestrate.Step, error) {
		msgs, ok := step.Input.([]core.Message)
		if !ok {
			return step, fmt.Errorf("rerank step: expected []core.Message, got %T", step.Input)
		}

		candidates := make([]orchestrate.RerankCandidate, len(msgs))
		for i, msg := range msgs {
			candidates[i] = orchestrate.RerankCandidate{
				Content: msg.Content,
				Source:  fmt.Sprintf("doc-%d", i),
			}
		}

		result, err := reranker.Rerank(ctx, orchestrate.RerankInput{
			Query:      query,
			Candidates: candidates,
		})
		if err != nil {
			return step, err
		}

		fmt.Fprintf(os.Stderr, "Reranked %d -> %d documents\n", len(candidates), len(result))
		for _, c := range result {
			fmt.Fprintf(os.Stderr, "  [%.2f] %s\n", c.Score, truncate(c.Content, 60))
		}

		step.Output = result
		return step, nil
	}
}

// inferWithContextHandler builds messages from reranked candidates and runs inference.
func inferWithContextHandler(engine inference.Engine, query string) orchestrate.StepHandler {
	return func(ctx context.Context, step orchestrate.Step) (orchestrate.Step, error) {
		candidates, ok := step.Input.([]orchestrate.RerankCandidate)
		if !ok {
			return step, fmt.Errorf("infer step: expected []RerankCandidate, got %T", step.Input)
		}

		var contextParts []string
		for _, c := range candidates {
			contextParts = append(contextParts, c.Content)
		}

		messages := []core.Message{
			core.NewSystemMessage("Answer using only the provided context. If context is insufficient, say so."),
			core.NewSystemMessage("Context:\n" + strings.Join(contextParts, "\n---\n")),
			core.NewUserMessage(query),
		}

		result, err := engine.Infer(ctx, inference.Request{
			Messages:  messages,
			MaxTokens: 2048,
		})
		if err != nil {
			return step, err
		}

		step.Output = result
		return step, nil
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// Usage:
//
//	go run examples/cli/rerank-query/main.go "How do Go interfaces work?"
//	go run examples/cli/rerank-query/main.go "Explain goroutines and channels"
//	echo "What is Go context used for?" | go run examples/cli/rerank-query/main.go
