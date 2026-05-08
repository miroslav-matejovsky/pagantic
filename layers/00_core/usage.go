package core

// TokenUsage tracks token consumption and throughput for a single
// inference call.
type TokenUsage struct {
	PromptTokens    int
	ReasoningTokens int
	OutputTokens    int
	ContextTokens   int
	ContextWindow   int
	TokensPerSecond float64
}
