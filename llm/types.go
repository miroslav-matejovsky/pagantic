package llm

import (
	"context"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
)

// Chat abstracts LLM interaction. Implemented by *kronk.Kronk.
// Phase agents depend on this interface, not the concrete engine.
type Chat interface {
	ChatStreaming(ctx context.Context, d model.D) (<-chan model.ChatResponse, error)
	ModelConfig() model.Config
}

// ToolCallInfo holds parsed information about a single tool call
// requested by the LLM.
type ToolCallInfo struct {
	ID        string
	Name      string
	Arguments map[string]any
}

// ChatResult holds the outcome of a single chat round with the LLM.
// Either ToolCalls or Content will be populated, not both.
type ChatResult struct {
	ToolCalls []ToolCallInfo
	Content   string
	Messages  []model.D
	Usage     TokenUsage
}

// TokenUsage holds token statistics from a single LLM call.
type TokenUsage struct {
	PromptTokens    int
	ReasoningTokens int
	OutputTokens    int
	ContextTokens   int
	ContextWindow   int
	TokensPerSecond float64
}
