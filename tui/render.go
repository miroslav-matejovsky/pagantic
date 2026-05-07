package tui

import (
	"fmt"
	"io"
)

// UsageStats holds token usage statistics for display purposes.
// It mirrors the display-relevant fields of LLM usage data without
// importing any LLM-specific packages, keeping pkg/tui dependency-free.
type UsageStats struct {
	PromptTokens    int
	ReasoningTokens int
	OutputTokens    int
	ContextTokens   int
	ContextWindow   int
	TokensPerSecond float64
}

// TerminalRenderer returns an onToken callback that renders streaming
// LLM/agent output to stdout with ANSI colors.
//
// Recognized token kinds:
//   - "reasoning": printed in red (model thinking tokens)
//   - "content": printed in default color
//   - "toolcall": printed in green with [TOOL] prefix
func TerminalRenderer() func(kind, text string) {
	return func(kind, text string) {
		switch kind {
		case "reasoning":
			fmt.Printf("%s%s%s", red, text, reset)
		case "content":
			fmt.Print(text)
		case "toolcall":
			fmt.Printf("\n%s[TOOL] %s%s\n", green, text, reset)
		}
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
