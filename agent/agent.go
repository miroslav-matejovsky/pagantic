package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
	"github.com/miroslav-matejovsky/pagantic/llm"
)

const defaultMaxTokens = 2048

// ToolProvider supplies tool definitions and execution for an agent.
// Implement this interface to give an agent access to external tools.
type ToolProvider interface {
	// Definitions returns OpenAI-style tool definitions to pass to the LLM.
	Definitions() []model.D
	// Execute runs the named tool with the given arguments and returns the result.
	Execute(name string, args map[string]any) (string, error)
}

// Config controls agent creation.
type Config struct {
	// SystemPrompt defines the agent's role and expertise. Required.
	SystemPrompt string
	// Engine is the LLM backend. Required.
	Engine llm.Chat
	// Tools is an optional tool provider. When nil, no tools are offered to the LLM.
	Tools ToolProvider
	// MaxTokens caps the LLM output per call. Defaults to 2048 when zero.
	MaxTokens int
	// OnToken is an optional callback for streaming output. Pass nil for silent operation.
	OnToken func(kind, text string)
	// OnToolCall is an optional callback for tool execution. Called with tool name and output.
	OnToolCall func(name, output string)
}

// Agent is a specialized LLM agent with a fixed system prompt and optional tool support.
// It maintains conversation history across Chat calls.
type Agent struct {
	cfg      Config
	messages []model.D
}

// New creates an Agent from the given config.
// The system prompt is pre-loaded into conversation history.
func New(cfg Config) *Agent {
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = defaultMaxTokens
	}
	messages := model.DocumentArray()
	messages = append(messages, model.TextMessage(model.RoleSystem, cfg.SystemPrompt))
	return &Agent{cfg: cfg, messages: messages}
}

// Chat sends userMessage to the agent, resolves any tool calls in a loop,
// and returns the final content response. Conversation history is updated
// with each call so follow-up messages work naturally.
func (a *Agent) Chat(ctx context.Context, userMessage string) (llm.ChatResult, error) {
	a.messages = append(a.messages, model.TextMessage(model.RoleUser, userMessage))

	var toolDefs []model.D
	if a.cfg.Tools != nil {
		toolDefs = a.cfg.Tools.Definitions()
	}

	for {
		d := model.D{
			"messages":   a.messages,
			"max_tokens": a.cfg.MaxTokens,
		}
		if len(toolDefs) > 0 {
			d["tools"] = toolDefs
		}

		ch, err := a.cfg.Engine.ChatStreaming(ctx, d)
		if err != nil {
			return llm.ChatResult{}, fmt.Errorf("agent chat: %w", err)
		}

		result, err := llm.StreamResponse(a.cfg.Engine, a.messages, ch, a.cfg.OnToken)
		if err != nil {
			return llm.ChatResult{}, fmt.Errorf("agent stream: %w", err)
		}
		a.messages = result.Messages

		if len(result.ToolCalls) == 0 {
			return result, nil
		}

		for _, tc := range result.ToolCalls {
			output, execErr := a.cfg.Tools.Execute(tc.Name, tc.Arguments)
			if execErr != nil {
				output = fmt.Sprintf("Error: %v", execErr)
			}
			if a.cfg.OnToolCall != nil {
				a.cfg.OnToolCall(tc.Name, output)
			}
			a.messages = append(a.messages, model.D{
				"role":         "tool",
				"tool_call_id": tc.ID,
				"name":         tc.Name,
				"content":      output,
			})
		}
	}
}

// ChatStructured sends userMessage and constrains LLM output to match
// the given JSON Schema. The returned ChatResult.Content contains valid
// JSON conforming to the schema.
//
// Builds on accumulated conversation history so prior Chat calls (e.g. a
// tool loop that collected data) provide context for the structured output.
// Conversation history is NOT updated by this call - it is a one-shot
// structured extraction on top of the current context.
// Thinking is disabled so grammar enforcement starts from the first token.
func (a *Agent) ChatStructured(ctx context.Context, userMessage string, jsonSchema model.D) (llm.ChatResult, error) {
	messages := model.DocumentArray()
	messages = append(messages, a.messages...)
	messages = append(messages, model.TextMessage(model.RoleUser, userMessage))

	d := model.D{
		"messages":        messages,
		"json_schema":     jsonSchema,
		"enable_thinking": false,
		"temperature":     0.3,
		"max_tokens":      a.cfg.MaxTokens,
	}

	ch, err := a.cfg.Engine.ChatStreaming(ctx, d)
	if err != nil {
		return llm.ChatResult{}, fmt.Errorf("agent structured chat: %w", err)
	}

	result, err := llm.StreamResponse(a.cfg.Engine, messages, ch, a.cfg.OnToken)
	if err != nil {
		return llm.ChatResult{}, fmt.Errorf("agent structured stream: %w", err)
	}
	// Grammar-constrained generation may truncate trailing closers.
	// Only attempt repair when output is not already valid JSON.
	if !json.Valid([]byte(result.Content)) {
		repaired := repairJSON(result.Content)
		if !json.Valid([]byte(repaired)) {
			return llm.ChatResult{}, fmt.Errorf("agent structured output: invalid JSON after repair: %s", repaired)
		}
		result.Content = repaired
	}
	return result, nil
}
