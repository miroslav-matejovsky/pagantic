// TUI example: interactive chat with per-turn context retrieval.
//
// Demonstrates pagantic as a Probabilistic Agentic Control System:
//
//   - Layer 3 (context): InMemoryRetriever stores domain documents and finds
//     relevant ones by keyword scoring. On each user message, ContextBuilder
//     retrieves fresh context so different questions get different knowledge.
//
//   - Layer 2 (orchestrate): AgentLoop calls ContextProvider before each user
//     message turn. The model receives bounded, relevant knowledge per turn
//     rather than a fixed context window. Context is injected ephemerally
//     (not stored in history), so different questions get different context
//     without accumulating stale knowledge. This is the control loop that makes
//     RAG work - deterministic retrieval feeding probabilistic inference.
//
//   - Layer 1 (inference): Engine loaded lazily on first chat message. Handles
//     model interaction. Streaming tokens displayed in real time.
//
//   - Adapter (tui): AgentREPL provides REPL with command dispatch, tool
//     status, and chat mode. Thin boundary, no business logic.
//
// Key pagantic concept shown here: per-turn context retrieval. Each user
// message triggers fresh retrieval, so the model always gets the most
// relevant knowledge for the current question.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/miroslav-matejovsky/pagantic/adapters/tui"
	"github.com/miroslav-matejovsky/pagantic/kronk"
	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	orchestrate "github.com/miroslav-matejovsky/pagantic/layers/02_orchestrate"
	pctx "github.com/miroslav-matejovsky/pagantic/layers/03_context"
	tool "github.com/miroslav-matejovsky/pagantic/layers/04_tool"
)

const llmModel = "unsloth/gemma-4-E4B-it"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	registry := tool.NewRegistry()

	// Build a context provider with domain knowledge about Go programming.
	// In a real application, documents would come from a database, file
	// system, API, or vector store.
	contextProvider := buildContextProvider()

	repl := tui.NewAgentREPL(tui.AgentConfig{
		Title:  "context-chat",
		Banner: "Chat with context retrieval. Type 'ctx-chat' to start, 'quit' to exit.\nAsk about Go programming - the model has domain knowledge loaded.",
		SystemPrompt: "You are a Go programming assistant. Answer questions using the provided context. " +
			"If context is relevant, reference it. If not, say you don't have specific information.",
		EngineLoader: func(ctx context.Context) (inference.Engine, func(), error) {
			krn, cleanup, err := kronk.Load(ctx, kronk.Config{ModelSource: llmModel})
			if err != nil {
				return nil, nil, err
			}
			return kronk.NewAdapter(krn, nil), cleanup, nil
		},
		Registry: registry,
	})

	// Register a custom command to show what documents are available.
	repl.AddCommand(tui.Command{
		Name:        "docs",
		Description: "Show available domain knowledge documents",
		Run: func(_ context.Context, _ []string) error {
			fmt.Println(tui.Bold("Domain knowledge documents:"))
			for _, doc := range domainDocs() {
				fmt.Printf("  %s %s\n", tui.Green("["+doc.Source+"]"), tui.Dim(truncate(doc.Content, 80)))
			}
			return nil
		},
	})

	// Demonstrate wiring: ContextProvider is set on AgentConfig via a
	// custom chat command that creates its own AgentLoop with context.
	// This shows how the adapter layer stays thin while the orchestrate
	// layer handles context integration.
	repl.AddCommand(tui.Command{
		Name:        "ctx-chat",
		Description: "Chat with context retrieval enabled",
		Run: func(ctx context.Context, _ []string) error {
			eng, err := repl.Engine(ctx)
			if err != nil {
				return err
			}

			loop, err := orchestrate.NewAgentLoop(orchestrate.LoopConfig{
				Engine:            eng,
				SystemPrompt:      "You are a Go programming assistant. Use provided context to answer accurately.",
				ContextProvider:   contextProvider,
				Stream:            tui.TerminalRenderer(os.Stdout),
				MaxTokens:         2048,
				MaxToolIterations: 20,
			})
			if err != nil {
				return err
			}

			fmt.Println(tui.Cyan("\n=== Context Chat Mode ==="))
			fmt.Println("Context retrieval active. Ask about Go topics.")
			fmt.Println("Type 'exit' to return.")
			fmt.Println()

			scanner := bufio.NewScanner(os.Stdin)
			for {
				line, err := tui.FPrompt(scanner, os.Stdout, tui.Bold("ctx>")+" ")
				if err != nil || line == "exit" || line == "quit" {
					break
				}

				result, err := loop.Chat(ctx, line)
				if err != nil {
					tui.FErrorf(os.Stderr, "Error: %v", err)
					continue
				}
				fmt.Println()
				tui.FPrintUsage(os.Stdout, tui.UsageStats{
					PromptTokens:    result.Usage.PromptTokens,
					ReasoningTokens: result.Usage.ReasoningTokens,
					OutputTokens:    result.Usage.OutputTokens,
				})
			}
			fmt.Println(tui.Grey("Back to main menu."))
			return nil
		},
	})

	repl.Run(ctx)
	fmt.Println("Done.")
}

func buildContextProvider() orchestrate.ContextProvider {
	return &pctx.ContextBuilder{
		Retriever: pctx.NewInMemoryRetriever(domainDocs()...),
		MaxChunks: 3,
	}
}

func domainDocs() []pctx.Document {
	return []pctx.Document{
		{
			Content: "Go interfaces are satisfied implicitly. A type implements an interface by implementing its methods. No 'implements' keyword needed. This is structural typing.",
			Source:  "go-interfaces",
		},
		{
			Content: "Go goroutines are lightweight threads managed by the Go runtime. Launch with 'go' keyword. Communicate via channels. Much cheaper than OS threads.",
			Source:  "go-concurrency",
		},
		{
			Content: "Go error handling uses explicit return values, not exceptions. Functions return error as last value. Check with 'if err != nil'. Wrap with fmt.Errorf and %w verb.",
			Source:  "go-errors",
		},
		{
			Content: "Go modules manage dependencies. go.mod declares module path and dependencies. go.sum locks checksums. Use 'go mod tidy' to clean up.",
			Source:  "go-modules",
		},
		{
			Content: "Go context package carries deadlines, cancellation signals, and request-scoped values across API boundaries. Always pass context as first parameter.",
			Source:  "go-context",
		},
		{
			Content: "Go testing uses the testing package. Test functions start with Test prefix. Run with 'go test'. Use table-driven tests for multiple cases. testify/require for assertions.",
			Source:  "go-testing",
		},
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
