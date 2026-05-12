// CLI example: single-shot query without context or tools.
//
// Demonstrates the simplest pagantic pattern - direct inference with
// deterministic control:
//
//   - Layer 2 (orchestrate): AgentLoop wraps one inference call. Even this
//     minimal case uses the full control loop - the loop just completes in
//     one iteration since there are no tools to resolve.
//
//   - Layer 1 (inference): kronk engine handles raw model interaction. The
//     Engine interface hides backend details (model loading, tokenization,
//     streaming protocol).
//
//   - Adapter (cli): ReadPrompt reads from args or stdin. Runner delegates
//     to orchestrate layer. No business logic in the adapter.
//
// Key pagantic concept: even the simplest query goes through the harness.
// The control system wraps probabilistic inference with deterministic
// scaffolding (timeouts, error handling, streaming) from the start.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/miroslav-matejovsky/pagantic/adapters/cli"
	"github.com/miroslav-matejovsky/pagantic/kronk"
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

	runner, err := cli.NewRunner(cli.RunConfig{
		Engine:       engine,
		SystemPrompt: "You are a helpful assistant. Be concise.",
		Out:          os.Stdout,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := runner.Run(ctx, prompt); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Usage:
//   go run examples/cli/simple-query/main.go "What is the capital of France?"
//   echo "Explain Go interfaces" | go run examples/cli/simple-query/main.go
