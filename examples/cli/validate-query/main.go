// CLI example: post-hoc JSON validation and repair.
//
// Demonstrates output validation from layer 5 (constraint):
//
//   - Layer 5 (constraint): JSONValidator checks model output after generation.
//     Unlike grammar constraints (which prevent invalid tokens during decoding),
//     validation runs after the full response. JSONValidator can attempt repair
//     using RepairJSON before rejecting output.
//
//   - GrammarConstraint wraps GrammarDefinition as a DecoderConstraint.
//     GrammarString() returns GBNF grammar for decoder enforcement. This is
//     how the grammar travels from constraint layer to inference engine.
//
//   - Two-layer validation:
//     1. Grammar constrains decoder tokens (prevention)
//     2. JSONValidator checks final output (verification)
//     Together they provide defense-in-depth for structured output.
//
// Compare with grammar-query example:
//   - grammar-query: shows GrammarDefinition.GrammarString() directly
//   - validate-query: shows GrammarConstraint wrapper + JSONValidator
//
// When to use post-hoc validation:
//   - API-based models that don't support grammar constraints
//   - Validating output from grammar-constrained models (belt and suspenders)
//   - Output repair when minor JSON errors are tolerable
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/miroslav-matejovsky/pagantic/adapters/cli"
	"github.com/miroslav-matejovsky/pagantic/kronk"
	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	orchestrate "github.com/miroslav-matejovsky/pagantic/layers/02_orchestrate"
	constraint "github.com/miroslav-matejovsky/pagantic/layers/05_constraint"
)

const llmModel = "unsloth/gemma-4-E4B-it"

const sentimentGrammar = `root        ::= "{" ws "\"sentiment\"" ws ":" ws sentiment ws "," ws "\"reason\"" ws ":" ws reason ws "}"
ws          ::= [ \t\n]*
sentiment   ::= "\"positive\"" | "\"negative\"" | "\"neutral\""
reason      ::= "\"" [a-zA-Z0-9 .,!?'-]+ "\""
`

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	prompt, err := cli.ReadPrompt(os.Args[1:], os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// GrammarConstraint wraps GrammarDefinition as DecoderConstraint interface.
	gc := constraint.GrammarConstraint{
		Definition: constraint.GrammarDefinition{
			Name:    "sentiment-with-reason",
			Grammar: sentimentGrammar,
		},
	}

	// GrammarString() returns GBNF for decoder enforcement.
	grammarStr := gc.GrammarString()
	fmt.Fprintf(os.Stderr, "Grammar constraint: %s\n", gc.Definition.Name)

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
				Enum: []string{"positive", "negative", "neutral"},
			},
			"reason": {
				Type:        "string",
				Description: "Brief reason for the classification",
			},
		},
		Required: []string{"sentiment", "reason"},
	}

	sa, err := orchestrate.NewSpecializedLoop(orchestrate.SpecializedConfig{
		SystemPrompt:      "Classify the sentiment. Return JSON with sentiment and reason fields.",
		Engine:            engine,
		Schema:            schema,
		Grammar:           grammarStr,
		MaxTokens:         2048,
		MaxToolIterations: 20,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Inference error: %v\n", err)
		os.Exit(1)
	}

	callCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	result, err := sa.Call(callCtx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Inference error: %v\n", err)
		os.Exit(1)
	}

	// Post-hoc validation with JSONValidator. AttemptRepair=true tries
	// RepairJSON before rejecting.
	validator := constraint.NewJSONValidator(true)
	vr := validator.Validate(result.Content)

	fmt.Fprintf(os.Stderr, "\nValidation result:\n")
	fmt.Fprintf(os.Stderr, "  Valid:  %v\n", vr.Valid)
	if len(vr.Errors) > 0 {
		for _, e := range vr.Errors {
			fmt.Fprintf(os.Stderr, "  Error: %s\n", e)
		}
	}

	// Also validate a deliberately broken JSON to show repair path.
	broken := `{"sentiment": "positive", "reason": "great product"`
	brokenResult := validator.Validate(broken)
	fmt.Fprintf(os.Stderr, "\nRepair demo:\n")
	fmt.Fprintf(os.Stderr, "  Input:    %s\n", broken)
	fmt.Fprintf(os.Stderr, "  Repaired: %s\n", brokenResult.Output)
	fmt.Fprintf(os.Stderr, "  Valid:    %v\n", brokenResult.Valid)

	// Print validated output.
	var parsed map[string]any
	if err := json.Unmarshal([]byte(vr.Output), &parsed); err == nil {
		out, _ := json.MarshalIndent(parsed, "", "  ")
		fmt.Println(string(out))
	} else {
		fmt.Println(vr.Output)
	}
}

// Usage:
//
//	go run examples/cli/validate-query/main.go "I love this product, it works great!"
//	go run examples/cli/validate-query/main.go "The service was terrible and slow."
//	echo "Just a normal Tuesday" | go run examples/cli/validate-query/main.go
