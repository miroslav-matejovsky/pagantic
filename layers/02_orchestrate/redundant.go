package orchestrate

import (
	"context"
	"fmt"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	observe "github.com/miroslav-matejovsky/pagantic/layers/10_observe"
)

// VotingStrategy selects the best result from multiple inference candidates.
// Implementations define how to pick a winner from N results.
type VotingStrategy interface {
	// Vote picks the winning content from candidates.
	// Returns winner text, confidence (0-1), and error.
	Vote(candidates []string) (winner string, confidence float64, err error)
}

// MajorityVoting picks the most frequent result. Confidence is the fraction
// of candidates that match the winner. For structured output, exact string
// comparison is used.
type MajorityVoting struct{}

// Vote picks the most frequent candidate.
func (m MajorityVoting) Vote(candidates []string) (string, float64, error) {
	if len(candidates) == 0 {
		return "", 0, fmt.Errorf("majority voting: no candidates")
	}
	if len(candidates) == 1 {
		return candidates[0], 1.0, nil
	}

	counts := make(map[string]int, len(candidates))
	for _, c := range candidates {
		counts[c]++
	}

	var winner string
	var maxCount int
	for c, count := range counts {
		if count > maxCount {
			winner = c
			maxCount = count
		}
	}

	confidence := float64(maxCount) / float64(len(candidates))
	return winner, confidence, nil
}

// UnanimityVoting requires all candidates to match. Returns error if any
// candidate differs.
type UnanimityVoting struct{}

// Vote requires all candidates match.
func (u UnanimityVoting) Vote(candidates []string) (string, float64, error) {
	if len(candidates) == 0 {
		return "", 0, fmt.Errorf("unanimity voting: no candidates")
	}

	first := candidates[0]
	for i := 1; i < len(candidates); i++ {
		if candidates[i] != first {
			return "", 0, fmt.Errorf("unanimity voting: candidate %d differs from candidate 0", i)
		}
	}

	return first, 1.0, nil
}

// RedundantResult holds output from redundant inference.
type RedundantResult struct {
	// Content is the winning result chosen by voting strategy.
	Content string
	// Confidence is how strongly the voting strategy supports the winner (0-1).
	Confidence float64
	// Candidates holds all N raw inference outputs.
	Candidates []string
	// Usage is combined token usage across all inference calls.
	Usage core.TokenUsage
}

// RedundantConfig configures redundant inference loop.
type RedundantConfig struct {
	Engine       inference.Engine
	SystemPrompt string
	Schema       core.Schema
	Grammar      string // GBNF grammar; empty means none
	N            int    // number of candidates; 0 defaults to 3
	Voting       VotingStrategy
	MaxTokens    int
	Observer     observe.EventLog
}

// RedundantLoop runs same inference N times and picks winner by voting.
// Implements Triple Modular Redundancy / N-Version programming pattern
// for LLM output.
type RedundantLoop struct {
	cfg RedundantConfig
}

// NewRedundantLoop creates RedundantLoop with given config.
func NewRedundantLoop(cfg RedundantConfig) *RedundantLoop {
	if cfg.N <= 0 {
		cfg.N = 3
	}
	if cfg.Voting == nil {
		cfg.Voting = MajorityVoting{}
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = defaultMaxTokens
	}
	return &RedundantLoop{cfg: cfg}
}

// Call runs N inference calls and applies voting strategy.
// Calls run sequentially to avoid GPU contention.
func (rl *RedundantLoop) Call(ctx context.Context, prompt string) (*RedundantResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if rl == nil || rl.cfg.Engine == nil {
		return nil, fmt.Errorf("redundant loop: nil engine")
	}

	candidates := make([]string, 0, rl.cfg.N)
	var totalUsage core.TokenUsage

	for i := 0; i < rl.cfg.N; i++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		sl := NewSpecializedLoop(SpecializedConfig{
			Engine:       rl.cfg.Engine,
			SystemPrompt: rl.cfg.SystemPrompt,
			Schema:       rl.cfg.Schema,
			Grammar:      rl.cfg.Grammar,
			MaxTokens:    rl.cfg.MaxTokens,
			Observer:     rl.cfg.Observer,
		})

		result, err := sl.Call(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("redundant loop: candidate %d failed: %w", i, err)
		}

		candidates = append(candidates, result.Content)
		totalUsage = addUsage(totalUsage, result.Usage)
	}

	winner, confidence, err := rl.cfg.Voting.Vote(candidates)
	if err != nil {
		return nil, fmt.Errorf("redundant loop: voting failed: %w", err)
	}

	return &RedundantResult{
		Content:    winner,
		Confidence: confidence,
		Candidates: candidates,
		Usage:      totalUsage,
	}, nil
}

// addUsage combines two TokenUsage records.
func addUsage(a, b core.TokenUsage) core.TokenUsage {
	return core.TokenUsage{
		PromptTokens:    a.PromptTokens + b.PromptTokens,
		ReasoningTokens: a.ReasoningTokens + b.ReasoningTokens,
		OutputTokens:    a.OutputTokens + b.OutputTokens,
		ContextTokens:   a.ContextTokens + b.ContextTokens,
		ContextWindow:   max(a.ContextWindow, b.ContextWindow),
		TokensPerSecond: (a.TokensPerSecond + b.TokensPerSecond) / 2,
	}
}
