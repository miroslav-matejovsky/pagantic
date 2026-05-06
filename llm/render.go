package llm

import "fmt"

// ANSI color escape codes used for terminal output.
const (
	colorReset     = "\033[0m"
	colorReasoning = "\033[91m" // bright red
	colorToolCall  = "\033[92m" // bright green
	colorUsage     = "\033[90m" // grey
)

// TerminalRenderer returns an onToken callback that renders LLM output
// to the terminal with ANSI colors.
//   - Reasoning tokens in bright red
//   - Content tokens in default color
//   - Tool calls in bright green
func TerminalRenderer() func(kind, text string) {
	return func(kind, text string) {
		switch kind {
		case "reasoning":
			fmt.Printf("%s%s%s", colorReasoning, text, colorReset)
		case "content":
			fmt.Print(text)
		case "toolcall":
			fmt.Printf("\n%s[TOOL] %s%s\n", colorToolCall, text, colorReset)
		}
	}
}

// PrintUsage displays token usage statistics in grey.
func PrintUsage(u TokenUsage) {
	percentage := (float64(u.ContextTokens) / float64(u.ContextWindow)) * 100
	of := float32(u.ContextWindow) / float32(1024)

	fmt.Printf("\n%sTokens - Input: %d  Reasoning: %d  Output: %d  Window: %d (%.0f%% of %.0fK)  TPS: %.2f%s\n",
		colorUsage,
		u.PromptTokens, u.ReasoningTokens, u.OutputTokens,
		u.ContextTokens, percentage, of, u.TokensPerSecond,
		colorReset)
}
