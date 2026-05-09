// TUI example: schema-constrained structured output.
//
// Demonstrates pagantic as a Probabilistic Agentic Control System:
//
//   - Layer 2 (orchestrate): SpecializedLoop runs a single-shot inference
//     that produces JSON matching a schema. Unlike AgentLoop, SpecializedLoop
//     is stateless - fresh loop per call - optimized for extract-and-structure
//     patterns.
//
//   - Layer 5 (constraint): SchemaValidator, RepairJSON, and NormalizeEnumValues
//     enforce structured output. The model's probabilistic text generation is
//     constrained to valid JSON matching the schema. If the model produces
//     broken JSON, RepairJSON attempts recovery before validation.
//
//   - Layer 0 (core): Schema defines the expected output shape with types,
//     enums, and required fields. This is the contract between the harness
//     and the model.
//
//   - Layer 1 (inference): Engine handles streaming. TerminalRenderer shows
//     tokens as they arrive so the user sees progress before validation.
//
// Key pagantic concept: deterministic post-processing of probabilistic output.
// The model generates freely, then the constraint layer forces the output into
// a valid, typed structure. This is the "control" in Probabilistic Agentic
// Control System.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/miroslav-matejovsky/pagantic/adapters/tui"
	"github.com/miroslav-matejovsky/pagantic/kronk"
	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	orchestrate "github.com/miroslav-matejovsky/pagantic/layers/02_orchestrate"
)

// const llmModel = "unsloth/gemma-4-E4B-it"
const llmModel = "unsloth/Qwen3-0.6B-Q8_0"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	krn, cleanup, err := kronk.Load(ctx, kronk.Config{ModelSource: llmModel})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load engine: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	engine := kronk.NewAdapter(krn, nil)
	schema := core.Schema{
		Type: "object",
		Properties: map[string]core.Schema{
			"sentiment": {
				Type: "string",
				Enum: []string{"positive", "neutral", "negative"},
			},
			"confidence": {
				Type: "number",
			},
			"explanation": {
				Type: "string",
			},
		},
		Required: []string{"sentiment", "confidence", "explanation"},
	}

	sa := orchestrate.NewSpecializedLoop(orchestrate.SpecializedConfig{
		SystemPrompt: "Analyze the sentiment of the given text. Return structured JSON with sentiment, confidence (0-1), and a brief explanation.",
		Engine:       engine,
		Schema:       schema,
		Stream:       tui.TerminalRenderer(os.Stdout),
	})

	sentence := "The weather is absolutely beautiful today, I love it!"
	fmt.Printf("Analyzing: %q\n\n", sentence)

	callCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	result, err := sa.Call(callCtx, sentence)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println()
	fmt.Println("Repaired result:")
	fmt.Println(result.Content)
	tui.FPrintUsage(os.Stdout, tui.UsageStats{
		PromptTokens:    result.Usage.PromptTokens,
		ReasoningTokens: result.Usage.ReasoningTokens,
		OutputTokens:    result.Usage.OutputTokens,
		ContextTokens:   result.Usage.ContextTokens,
		ContextWindow:   result.Usage.ContextWindow,
		TokensPerSecond: result.Usage.TokensPerSecond,
	})
	fmt.Println("\nDone.")
}
