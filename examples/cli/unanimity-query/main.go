// CLI example: redundant inference with unanimity voting.
//
// Demonstrates strict consensus voting (UnanimityVoting) as alternative to
// MajorityVoting in redundant inference:
//
//   - Layer 2 (orchestrate): RedundantLoop runs same inference N times.
//     UnanimityVoting requires ALL candidates to match exactly. If any
//     candidate differs, voting fails with error. This is "fail closed" -
//     prefer no answer over a wrong one.
//
//   - GBNF grammar constrains output to exactly one of three sentiment
//     values. Grammar makes unanimity achievable because token generation
//     is deterministic within grammar bounds (same tokens allowed, same
//     sampling temperature).
//
//   - Layer 5 (constraint): Schema validates JSON structure. Grammar
//     constrains decoder tokens. Both applied per candidate independently.
//
// Compare with redundant-query example which uses MajorityVoting:
//   - MajorityVoting: picks most common result, always produces answer
//   - UnanimityVoting: requires all N agree, fails if any differ
//
// When to use unanimity:
//   - Safety-critical classification (medical, legal, financial)
//   - Binary yes/no decisions where wrong answer is worse than no answer
//   - Grammar-constrained output where unanimity is likely achievable
//
// When NOT to use:
//   - Open-ended generation (candidates will always differ)
//   - Tasks where approximate consensus is acceptable
//   - Without grammar constraints (free-form text rarely matches exactly)
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
)

const llmModel = "unsloth/gemma-4-E4B-it"

// sentimentGrammar forces output to exactly one JSON object with a single
// sentiment field. Decoder cannot produce anything outside this grammar.
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

	const redundantN = 3

	// UnanimityVoting: all redundantN candidates must produce identical output.
	// Grammar makes this achievable by constraining token generation.
	rl := orchestrate.NewRedundantLoop(orchestrate.RedundantConfig{
		Engine:       engine,
		SystemPrompt: "Classify the sentiment of the text. Return JSON with a sentiment field.",
		Schema:       schema,
		Grammar:      sentimentGrammar,
		N:            redundantN,
		Voting:       orchestrate.UnanimityVoting{},
	})

	callCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	fmt.Fprintf(os.Stderr, "Running %d redundant inferences with unanimity voting...\n", redundantN)
	result, err := rl.Call(callCtx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unanimity voting failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "This is expected when candidates differ. Use MajorityVoting for tolerance.\n")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "\nCandidates (all identical):\n")
	for i, c := range result.Candidates {
		fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, c)
	}
	fmt.Fprintf(os.Stderr, "\nConfidence: %.1f%% (unanimity = always 100%%)\n\n", result.Confidence*100)

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
//	go run examples/cli/unanimity-query/main.go "I absolutely love this new phone!"
//	go run examples/cli/unanimity-query/main.go "The food was cold and the service was rude."
//	echo "It is what it is" | go run examples/cli/unanimity-query/main.go
