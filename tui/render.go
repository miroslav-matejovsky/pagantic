package tui

import (
	"fmt"
	"io"

	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
)

// UsageStats holds token usage statistics for display purposes.
// It mirrors the display-relevant fields of core.TokenUsage so callers
// can pass usage data across package boundaries without importing
// inference-specific packages.
type UsageStats struct {
	PromptTokens    int
	ReasoningTokens int
	OutputTokens    int
	ContextTokens   int
	ContextWindow   int
	TokensPerSecond float64
}

// TerminalRenderer returns an *inference.StreamHandler that renders streaming
// inference and orchestration output to w with ANSI colors.
//
// Rendering:
//   - reasoning tokens in red
//   - content tokens in default color
//   - tool calls in green with [TOOL] prefix
func TerminalRenderer(w io.Writer) *inference.StreamHandler {
	return &inference.StreamHandler{
		OnReasoning: func(text string) {
			_, _ = fmt.Fprintf(w, "%s%s%s", red, text, reset)
		},
		OnContent: func(text string) {
			_, _ = fmt.Fprint(w, text)
		},
		OnToolCall: func(name, argsJSON string) {
			_, _ = fmt.Fprintf(w, "\n%s[TOOL] %s(%s)%s\n", green, name, argsJSON, reset)
		},
	}
}

// FPrintUsage writes token usage statistics to w in grey.
func FPrintUsage(w io.Writer, u UsageStats) {
	if u.ContextWindow > 0 {
		percentage := (float64(u.ContextTokens) / float64(u.ContextWindow)) * 100
		of := float32(u.ContextWindow) / float32(1024)
		_, _ = fmt.Fprintf(w, "\n%sTokens - Input: %d  Reasoning: %d  Output: %d  Window: %d (%.0f%% of %.0fK)  TPS: %.2f%s\n",
			grey,
			u.PromptTokens, u.ReasoningTokens, u.OutputTokens,
			u.ContextTokens, percentage, of, u.TokensPerSecond,
			reset)
	} else {
		_, _ = fmt.Fprintf(w, "\n%sTokens - Input: %d  Reasoning: %d  Output: %d  TPS: %.2f%s\n",
			grey,
			u.PromptTokens, u.ReasoningTokens, u.OutputTokens,
			u.TokensPerSecond,
			reset)
	}
}
