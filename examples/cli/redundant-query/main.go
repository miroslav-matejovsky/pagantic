// CLI example: redundant inference with majority voting.
//
// Demonstrates Triple Modular Redundancy (TMR) / N-Version programming
// applied to LLM inference:
//
//   - Layer 2 (orchestrate): RedundantLoop runs the same inference N times
//     (default 3) and applies a VotingStrategy to pick the best result.
//     This catches stochastic errors - if one inference call produces a
//     wrong answer, the majority vote corrects it.
//
//   - VotingStrategy interface: MajorityVoting picks the most frequent
//     result among N candidates. Confidence = count(winner)/N. For
//     structured output (JSON), exact string comparison works well since
//     schema constraints normalize the format.
//
//   - UnanimityVoting is the strict alternative: all N candidates must
//     match exactly, otherwise it errors. Use when correctness is critical
//     and you prefer to fail rather than guess.
//
//   - Layer 5 (constraint): Schema validation runs on each individual
//     candidate via SpecializedLoop. Each candidate is independently valid
//     JSON before voting compares them.
//
//   - Performance: N inference calls run sequentially (not parallel) to
//     avoid GPU contention on local models. For API-based models, parallel
//     execution would be better.
//
// When to use redundant inference:
//   - Classification tasks where wrong answers are costly
//   - Structured extraction where consistency matters
//   - Any case where you trade latency for reliability
//
// When NOT to use:
//   - Creative or open-ended tasks (results should differ)
//   - Latency-critical paths (3x inference time)
//   - Tasks where deterministic constraints (grammar, schema) suffice
//
// Key pagantic concept: the control system can repeat probabilistic
// operations and apply deterministic voting to increase reliability.
// Same idea as TMR in hardware fault tolerance, applied to LLM output.
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
			"confidence": {
				Type: "number",
			},
		},
		Required: []string{"sentiment", "confidence"},
	}

	// RedundantLoop runs 3 inference calls and picks the majority result.
	rl := orchestrate.NewRedundantLoop(orchestrate.RedundantConfig{
		Engine:       engine,
		SystemPrompt: "Analyze sentiment. Return JSON with sentiment (positive/negative/neutral) and confidence (0-1).",
		Schema:       schema,
		N:            3,
		Voting:       orchestrate.MajorityVoting{},
	})

	callCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	fmt.Fprintf(os.Stderr, "Running %d redundant inferences...\n", 3)
	result, err := rl.Call(callCtx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Show all candidates and voting result.
	fmt.Fprintf(os.Stderr, "\nCandidates:\n")
	for i, c := range result.Candidates {
		fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, c)
	}
	fmt.Fprintf(os.Stderr, "\nConfidence: %.1f%%\n\n", result.Confidence*100)

	// Pretty-print winner.
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
//	go run examples/cli/redundant-query/main.go "I absolutely love this new phone!"
//	go run examples/cli/redundant-query/main.go "The food was cold and the service was rude."
//	echo "It is what it is" | go run examples/cli/redundant-query/main.go
