// CLI example: single-shot query with context retrieval.
//
// Demonstrates pagantic as a Probabilistic Agentic Control System:
//
//   - Layer 3 (context): InMemoryRetriever finds relevant documents by keyword
//     matching and scores them by relevance. ContextBuilder assembles matched
//     chunks into system messages that give the model bounded knowledge.
//
//   - Layer 2 (orchestrate): AgentLoop drives multi-step inference with its
//     ContextProvider injecting retrieved context before each user message.
//     The model never acts on unconstrained knowledge - only what the retriever
//     provides.
//
//   - Layer 1 (inference): kronk engine handles raw model interaction. The
//     harness wraps probabilistic output with deterministic retrieval and
//     control flow.
//
//   - Adapter (cli): Thin boundary. Reads prompt, delegates to orchestrate
//     layer, writes output. No business logic.
//
// This pattern is Retrieval-Augmented Generation (RAG) at its simplest:
// deterministic retrieval constrains what the probabilistic model can know.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/miroslav-matejovsky/pagantic/adapters/cli"
	"github.com/miroslav-matejovsky/pagantic/kronk"
	orchestrate "github.com/miroslav-matejovsky/pagantic/layers/02_orchestrate"
	pctx "github.com/miroslav-matejovsky/pagantic/layers/03_context"
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

	// Domain knowledge documents. In production, these come from a database,
	// file system, or vector store. InMemoryRetriever uses keyword matching
	// to find relevant chunks - simple but effective for bounded domains.
	retriever := pctx.NewInMemoryRetriever(
		pctx.Document{
			Content: "Pagantic is a Probabilistic Agentic Control System. It wraps LLM inference with deterministic control layers for tools, validation, context retrieval, and structured output.",
			Source:  "architecture",
		},
		pctx.Document{
			Content: "The orchestrate layer drives multi-step agent loops. AgentLoop handles multi-turn chat with tool resolution. SpecializedLoop handles single-shot schema-constrained calls.",
			Source:  "orchestrate-docs",
		},
		pctx.Document{
			Content: "The context layer retrieves bounded knowledge for the model. Retriever finds relevant chunks. ContextBuilder assembles them into messages. The model never sees unconstrained knowledge.",
			Source:  "context-docs",
		},
		pctx.Document{
			Content: "The tool layer runs deterministic operations outside the model. All side effects live in tools, never in the model. Registry groups tools and dispatches execution.",
			Source:  "tool-docs",
		},
		pctx.Document{
			Content: "The constraint layer enforces structured output. JSON repair, schema validation, and enum normalization ensure model output matches expected format.",
			Source:  "constraint-docs",
		},
	)

	// ContextBuilder satisfies orchestrate.ContextProvider via Go structural
	// typing. No explicit interface implementation needed.
	contextProvider := &pctx.ContextBuilder{
		Retriever: retriever,
		MaxChunks: 3,
	}

	// Wire context into the agent loop. The ContextProvider retrieves
	// relevant documents before each inference call, giving the model
	// bounded domain knowledge instead of relying on training data alone.
	loop := orchestrate.NewAgentLoop(orchestrate.LoopConfig{
		Engine:          engine,
		SystemPrompt:    "You are a helpful assistant. Answer questions using only the provided context. If the context does not contain relevant information, say so.",
		ContextProvider: contextProvider,
	})

	result, err := loop.Chat(ctx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result.Content)
}

// Usage:
//   go run examples/cli/context-query/main.go "What is pagantic?"
//   go run examples/cli/context-query/main.go "How does the context layer work?"
//   echo "What does the tool layer do?" | go run examples/cli/context-query/main.go
