package llm

import "fmt"

// ANSI color escape codes used for terminal output.
const (
	colorReset     = "\033[0m"
	colorReasoning = "\033[91m" // bright red
	colorToolCall  = "\033[92m" // bright green
	colorUsage     = "\033[90m" // grey
)

// TerminalRenderer returns a StreamHandler that renders LLM output
// to the terminal with ANSI colors.
//   - Reasoning tokens in bright red
//   - Content tokens in default color
//   - Tool calls in bright green
func TerminalRenderer() *StreamHandler {
	return &StreamHandler{
		OnReasoning: func(text string) {
			fmt.Printf("%s%s%s", colorReasoning, text, colorReset)
		},
		OnContent: func(text string) {
			fmt.Print(text)
		},
		OnToolCall: func(name, argsJSON string) {
			fmt.Printf("\n%s[TOOL] %s(%s)%s\n", colorToolCall, name, argsJSON, colorReset)
		},
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
