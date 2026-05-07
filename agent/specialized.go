package agent

import (
	"context"
	"fmt"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
	"github.com/miroslav-matejovsky/pagantic/llm"
)

// phase2Prompt is the user message sent to the LLM to trigger structured output
// after the tool loop has finished accumulating context.
const phase2Prompt = "Produce your structured output now."

// SpecializedConfig controls SpecializedAgent creation.
type SpecializedConfig struct {
	// SystemPrompt defines the agent's role and the exact output contract. Required.
	SystemPrompt string
	// Engine is the LLM backend. Required.
	Engine llm.Chat
	// Schema is the JSON Schema the LLM output must conform to (the grammar). Required.
	// This is fixed at construction time and applied on every Call.
	Schema model.D
	// Tools is an optional tool provider. When set, Call runs a tool loop
	// (Phase 1) before the structured output call (Phase 2).
	Tools ToolProvider
	// MaxTokens caps the LLM output per call. Defaults to 2048 when zero.
	MaxTokens int
	// Stream receives typed streaming events (content, reasoning, tool-call requests).
	// Pass nil for silent operation.
	Stream *llm.StreamHandler
	// OnToolResult is called after a tool is executed with the tool name and output.
	OnToolResult func(name, output string)
}

// SpecializedAgent is a stateless LLM agent with a fixed system prompt and
// a fixed JSON schema that constrains every response. It is purpose-built for
// deterministic structured output: each Call is independent (no conversation
// history between calls).
//
// SpecializedAgent delegates to Agent internally:
//   - Phase 1 (when Tools configured): Agent.Chat drives the tool loop, accumulating
//     context in a per-call Agent instance.
//   - Phase 2: Agent.ChatStructured uses the accumulated context to produce JSON
//     conforming to the schema set at construction time.
//
// When no Tools are configured, only Phase 2 runs: a single structured call.
//
// Use SpecializedAgent when you know the output schema at construction time and
// want a clean single-call interface. Use Agent directly when you need stateful
// conversation or per-call schema variation.
type SpecializedAgent struct {
	cfg SpecializedConfig
}

// NewSpecialized creates a SpecializedAgent from the given config.
func NewSpecialized(cfg SpecializedConfig) *SpecializedAgent {
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = defaultMaxTokens
	}
	return &SpecializedAgent{cfg: cfg}
}

// Call sends prompt to the LLM and returns a response whose Content is valid
// JSON conforming to the schema set at construction time.
//
// When Tools are configured, Call runs two phases:
//   - Phase 1: Agent.Chat with tool calls until the LLM produces content.
//   - Phase 2: Agent.ChatStructured using accumulated message context + schema.
//
// When no Tools are configured, a single Agent.ChatStructured call is made.
// Each Call creates a fresh internal Agent, so calls are fully independent.
func (a *SpecializedAgent) Call(ctx context.Context, prompt string) (llm.ChatResult, error) {
	inner := New(Config{
		SystemPrompt: a.cfg.SystemPrompt,
		Engine:       a.cfg.Engine,
		MaxTokens:    a.cfg.MaxTokens,
		Tools:        a.cfg.Tools,
		Stream:       a.cfg.Stream,
		OnToolResult: a.cfg.OnToolResult,
	})

	if a.cfg.Tools != nil {
		if _, err := inner.Chat(ctx, prompt); err != nil {
			return llm.ChatResult{}, fmt.Errorf("specialized agent tool phase: %w", err)
		}
		return inner.ChatStructured(ctx, phase2Prompt, a.cfg.Schema)
	}

	return inner.ChatStructured(ctx, prompt, a.cfg.Schema)
}
