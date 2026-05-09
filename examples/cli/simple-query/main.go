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

	runner := cli.NewRunner(cli.RunConfig{
		Engine:       engine,
		SystemPrompt: "You are a helpful assistant. Be concise.",
		Out:          os.Stdout,
	})

	if err := runner.Run(ctx, prompt); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Usage:
//   go run examples/cli/simple-query/main.go "What is the capital of France?"
//   echo "Explain Go interfaces" | go run examples/cli/simple-query/main.go
