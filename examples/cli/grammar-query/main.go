// CLI example: GBNF grammar-constrained output.
//
// Demonstrates decoder-level output constraints via GBNF grammar:
//
//   - Layer 5 (constraint): GrammarDefinition holds a GBNF grammar string
//     that restricts what tokens the model can produce. Unlike post-hoc
//     validation (which checks output after generation), grammar constraints
//     prevent invalid tokens from being generated at all. This is the
//     strongest form of output control in pagantic.
//
//   - Layer 2 (orchestrate): SpecializedLoop passes the grammar through to
//     the inference engine via the Grammar field in SpecializedConfig. The
//     grammar and JSON schema work together - grammar constrains token
//     generation, schema validates the final structure.
//
//   - Layer 1 (inference): Request.Grammar carries the GBNF string to the
//     engine. The kronk adapter passes it to llama.cpp which enforces it
//     during decoding.
//
//   - Adapter (cli): Runner wraps the SpecializedLoop call with timeout
//     and output formatting. No grammar logic in the adapter.
//
// GBNF (GGML BNF) grammar syntax:
//
//	root        ::= <expression>       -- entry rule, always required
//	rulename    ::= <expression>       -- named rule
//	"literal"                          -- exact text match
//	[a-z]                              -- character range
//	rule1 rule2                        -- sequence
//	rule1 | rule2                      -- alternation
//	rule*                              -- zero or more
//	rule+                              -- one or more
//	rule?                              -- optional
//
// When to use grammar vs schema alone:
//   - Schema validates structure after generation (repair possible)
//   - Grammar constrains during generation (no invalid tokens produced)
//   - Use grammar for strict format requirements (e.g., enum-only output)
//   - Use schema for flexible structured output (e.g., JSON objects)
//   - Combine both for maximum control
//
// Key pagantic concept: decoder constraints are the strongest form of
// deterministic control over probabilistic output. The model literally
// cannot produce tokens outside the grammar.
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
	constraint "github.com/miroslav-matejovsky/pagantic/layers/05_constraint"

	orchestrate "github.com/miroslav-matejovsky/pagantic/layers/02_orchestrate"
)

const llmModel = "unsloth/gemma-4-E4B-it"

// sentimentGrammar constrains output to valid JSON with only allowed sentiment
// values. The model cannot produce tokens outside this grammar.
const sentimentGrammar = `root        ::= "{" ws "\"sentiment\"" ws ":" ws sentiment ws "}"
ws          ::= [ \t\n]*
sentiment   ::= "\"positive\"" | "\"negative\"" | "\"neutral\""
`

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	prompt, err := cli.ReadPrompt(os.Args[1:], os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Validate grammar before using it.
	grammar := constraint.GrammarDefinition{
		Name:    "sentiment-only",
		Grammar: sentimentGrammar,
	}
	if err := grammar.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid grammar: %v\n", err)
		os.Exit(1)
	}

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
		},
		Required: []string{"sentiment"},
	}

	// SpecializedLoop with both grammar and schema.
	// Grammar prevents invalid tokens during generation.
	// Schema validates final JSON structure after generation.
	sa := orchestrate.NewSpecializedLoop(orchestrate.SpecializedConfig{
		SystemPrompt: "Classify the sentiment of the text. Return JSON with a sentiment field.",
		Engine:       engine,
		Schema:       schema,
		Grammar:      grammar.GrammarString(),
	})

	callCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	result, err := sa.Call(callCtx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Pretty-print JSON output.
	var parsed map[string]any
	if err := json.Unmarshal([]byte(result.Content), &parsed); err == nil {
		out, _ := json.MarshalIndent(parsed, "", "  ")
		fmt.Println(string(out))
	} else {
		fmt.Println(result.Content)
	}
}

// Usage:
//
//	go run examples/cli/grammar-query/main.go "I love this product, it works great!"
//	go run examples/cli/grammar-query/main.go "The service was terrible and slow."
//	echo "Just a normal Tuesday" | go run examples/cli/grammar-query/main.go
