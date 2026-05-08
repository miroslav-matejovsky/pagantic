package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/miroslav-matejovsky/pagantic/core"
	"github.com/miroslav-matejovsky/pagantic/inference"
	"github.com/miroslav-matejovsky/pagantic/kronk"
	"github.com/miroslav-matejovsky/pagantic/orchestrate"
	"github.com/miroslav-matejovsky/pagantic/tui"
)

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

	engine := inference.NewKronkAdapter(krn, nil)
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
