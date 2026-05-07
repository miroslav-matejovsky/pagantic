package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
	"github.com/miroslav-matejovsky/pagantic/agent"
	"github.com/miroslav-matejovsky/pagantic/kronk"
	"github.com/miroslav-matejovsky/pagantic/tui"
)

const llmModel = "unsloth/Qwen3-0.6B-Q8_0"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	engine, cleanup, err := kronk.Load(ctx, kronk.Config{ModelSource: llmModel})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load engine: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	schema := model.D{
		"type": "object",
		"properties": model.D{
			"sentiment": model.D{
				"type": "string",
				"enum": []any{"positive", "neutral", "negative"},
			},
			"confidence": model.D{
				"type": "number",
			},
			"explanation": model.D{
				"type": "string",
			},
		},
		"required": []string{"sentiment", "confidence", "explanation"},
	}

	sa := agent.NewSpecialized(agent.SpecializedConfig{
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
