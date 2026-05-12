// CLI example: prompt construction via Go templates.
//
// Demonstrates prompt template rendering from layer 8 (prompt):
//
//   - Layer 8 (prompt): Template holds a raw string with {{.VarName}}
//     placeholders and a Variables map. Render() substitutes variables using
//     Go's text/template engine. Missing keys cause errors (missingkey=error).
//
//   - Separation of concerns: prompt construction is decoupled from
//     inference. Templates can be loaded from files, databases, or
//     configuration. Variables come from user input, environment, or context.
//
//   - Layer 2 (orchestrate): AgentLoop receives the rendered prompt as a
//     regular string. It has no knowledge of templates.
//
// When to use templates:
//   - Reusable prompt patterns across topics/domains
//   - A/B testing different prompt structures
//   - User-configurable prompts without code changes
//   - Multi-variable prompts (topic + format + constraints)
//
// Key pagantic concept: prompt construction is a deterministic layer.
// Templates are pure functions (input variables -> output string) with no
// probabilistic behavior.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/miroslav-matejovsky/pagantic/adapters/cli"
	"github.com/miroslav-matejovsky/pagantic/kronk"
	prompt "github.com/miroslav-matejovsky/pagantic/layers/08_prompt"
)

const llmModel = "unsloth/gemma-4-E4B-it"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	topic, err := cli.ReadPrompt(os.Args[1:], os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// System prompt built from template. Variables come from CLI args.
	systemTemplate := &prompt.Template{
		Raw: `You are an expert on {{.Topic}}. Explain concepts clearly and concisely. Use examples when helpful. Keep answers under 200 words.`,
		Variables: map[string]string{
			"Topic": topic,
		},
	}

	systemPrompt, err := systemTemplate.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Template render error: %v\n", err)
		os.Exit(1)
	}

	// User prompt also from template, with fixed structure.
	userTemplate := &prompt.Template{
		Raw: `Give me a brief overview of {{.Topic}}. What are the key concepts a beginner should know?`,
		Variables: map[string]string{
			"Topic": topic,
		},
	}

	userPrompt, err := userTemplate.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Template render error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "System: %s\n", systemPrompt)
	fmt.Fprintf(os.Stderr, "User:   %s\n\n", userPrompt)

	krn, cleanup, err := kronk.Load(ctx, kronk.Config{ModelSource: llmModel})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load engine: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	engine := kronk.NewAdapter(krn, nil)

	runner, err := cli.NewRunner(cli.RunConfig{
		Engine:       engine,
		SystemPrompt: systemPrompt,
		Out:          os.Stdout,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := runner.Run(ctx, userPrompt); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Usage:
//
//	go run examples/cli/template-query/main.go "Go interfaces"
//	go run examples/cli/template-query/main.go "distributed systems"
//	echo "machine learning" | go run examples/cli/template-query/main.go
