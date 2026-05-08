package inference

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
	"github.com/miroslav-matejovsky/pagantic/core"
)

// KronkAdapter wraps kronk chat engine and implements Engine.
type KronkAdapter struct {
	chat    llmChat
	handler *StreamHandler
}

// llmChat is interface kronk engine satisfies.
type llmChat interface {
	ChatStreaming(ctx context.Context, d model.D) (<-chan model.ChatResponse, error)
	ModelConfig() model.Config
}

// NewKronkAdapter builds adapter for kronk chat engine.
func NewKronkAdapter(chat llmChat, handler *StreamHandler) *KronkAdapter {
	return &KronkAdapter{chat: chat, handler: handler}
}

// WithStreamHandler returns copy of adapter that emits events to handler.
func (ka *KronkAdapter) WithStreamHandler(handler *StreamHandler) Engine {
	if ka == nil {
		return nil
	}
	return &KronkAdapter{chat: ka.chat, handler: handler}
}

// Infer runs one inference request through kronk.
func (ka *KronkAdapter) Infer(ctx context.Context, req Request) (*Result, error) {
	if ka == nil || ka.chat == nil {
		return nil, fmt.Errorf("inference: nil adapter")
	}

	requestD := model.D{}
	for key, value := range req.Options {
		requestD[key] = value
	}

	requestD["messages"] = messagesToD(req.Messages)
	if req.MaxTokens > 0 {
		requestD["max_tokens"] = req.MaxTokens
	}
	if req.Temperature != nil {
		requestD["temperature"] = *req.Temperature
	}
	if len(req.Tools) > 0 {
		requestD["tools"] = toolDefsToD(req.Tools)
	}
	if req.Schema != nil {
		requestD["json_schema"] = schemaToD(*req.Schema)
		if _, exists := requestD["enable_thinking"]; !exists {
			requestD["enable_thinking"] = false
		}
	}

	ch, err := ka.chat.ChatStreaming(ctx, requestD)
	if err != nil {
		return nil, fmt.Errorf("inference chat: %w", err)
	}

	messages := append([]core.Message(nil), req.Messages...)
	var reasoning bool
	var lastResp model.ChatResponse
	var content strings.Builder
	var toolCalls []core.ToolCall

loop:
	for resp := range ch {
		lastResp = resp

		if len(resp.Choices) == 0 {
			continue
		}

		choice := resp.Choices[0]
		delta := choice.Delta

		switch choice.FinishReason() {
		case model.FinishReasonError:
			msg := "unknown model error"
			if delta != nil {
				switch {
				case delta.Content != "":
					msg = delta.Content
				case delta.Reasoning != "":
					msg = delta.Reasoning
				}
			}
			return nil, fmt.Errorf("model error: %s", msg)

		case model.FinishReasonStop:
			if delta != nil && delta.Content != "" {
				if reasoning {
					reasoning = false
					ka.handler.emitReasoning("\n")
				}
				content.WriteString(delta.Content)
				ka.handler.emitContent(delta.Content)
			}
			break loop

		case model.FinishReasonTool:
			if reasoning {
				reasoning = false
				ka.handler.emitReasoning("\n")
			}

			assistantMsg := core.Message{Role: core.RoleAssistant}
			if content.Len() > 0 {
				assistantMsg.Content = content.String()
				content.Reset()
			}

			if delta != nil {
				for _, tool := range delta.ToolCalls {
					argsJSON, _ := json.Marshal(tool.Function.Arguments)
					call := core.ToolCall{
						ID:        tool.ID,
						Name:      tool.Function.Name,
						Arguments: map[string]any(tool.Function.Arguments),
					}

					ka.handler.emitToolCall(tool.Function.Name, string(argsJSON))

					assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, call)
					toolCalls = append(toolCalls, call)
				}
			}

			if assistantMsg.Content != "" || len(assistantMsg.ToolCalls) > 0 {
				messages = append(messages, assistantMsg)
			}
			break loop

		default:
			if delta == nil {
				continue
			}

			if delta.Reasoning != "" {
				ka.handler.emitReasoning(delta.Reasoning)
				reasoning = true
				continue
			}

			if reasoning {
				reasoning = false
				ka.handler.emitReasoning("\n")
			}

			if delta.Content == "" {
				continue
			}

			content.WriteString(delta.Content)
			ka.handler.emitContent(delta.Content)
		}
	}

	if content.Len() > 0 {
		messages = append(messages, core.NewAssistantMessage(content.String()))
	}

	return &Result{
		Content:   content.String(),
		ToolCalls: toolCalls,
		Messages:  messages,
		Usage:     extractUsage(ka.chat, lastResp),
	}, nil
}

// ModelInfo reports model identity and limits.
func (ka *KronkAdapter) ModelInfo() ModelInfo {
	if ka == nil || ka.chat == nil {
		return ModelInfo{}
	}

	cfg := ka.chat.ModelConfig()
	name := ""
	if len(cfg.ModelFiles) > 0 {
		name = filepath.Base(cfg.ModelFiles[0])
	}

	return ModelInfo{
		Name:          name,
		ContextWindow: cfg.ContextWindow(),
	}
}

// messagesToD converts core.Message slice to model.D format.
func messagesToD(msgs []core.Message) []model.D {
	docs := make([]model.D, 0, len(msgs))
	for _, msg := range msgs {
		docs = append(docs, messageToD(msg))
	}
	return docs
}

// messageToD converts one core.Message to model.D.
func messageToD(msg core.Message) model.D {
	doc := model.D{
		"role": string(msg.Role),
	}

	switch msg.Role {
	case core.RoleAssistant:
		if msg.Content != "" {
			doc["content"] = msg.Content
		}
		if len(msg.ToolCalls) > 0 {
			doc["tool_calls"] = toolCallsToD(msg.ToolCalls)
		}
	case core.RoleTool:
		doc["content"] = msg.Content
		if msg.ToolCallID != "" {
			doc["tool_call_id"] = msg.ToolCallID
		}
		if msg.Name != "" {
			doc["name"] = msg.Name
		}
	default:
		doc["content"] = msg.Content
	}

	return doc
}

func toolCallsToD(calls []core.ToolCall) []model.D {
	docs := make([]model.D, 0, len(calls))
	for _, call := range calls {
		argsJSON, _ := json.Marshal(call.Arguments)
		doc := model.D{
			"type": "function",
			"function": model.D{
				"name":      call.Name,
				"arguments": string(argsJSON),
			},
		}
		if call.ID != "" {
			doc["id"] = call.ID
		}
		docs = append(docs, doc)
	}
	return docs
}

// toolDefsToD converts core.ToolDefinition slice to model.D format.
func toolDefsToD(defs []core.ToolDefinition) []model.D {
	docs := make([]model.D, 0, len(defs))
	for _, def := range defs {
		function := model.D{
			"name":       def.Name,
			"parameters": schemaToD(def.Parameters),
		}
		if def.Description != "" {
			function["description"] = def.Description
		}
		docs = append(docs, model.D{
			"type":     "function",
			"function": function,
		})
	}
	return docs
}

// schemaToD converts core.Schema to model.D.
func schemaToD(s core.Schema) model.D {
	doc := model.D{}
	if s.Type != "" {
		doc["type"] = s.Type
	}
	if s.Description != "" {
		doc["description"] = s.Description
	}
	if len(s.Properties) > 0 {
		props := model.D{}
		for name, prop := range s.Properties {
			props[name] = schemaToD(prop)
		}
		doc["properties"] = props
	}
	if len(s.Required) > 0 {
		doc["required"] = s.Required
	}
	if len(s.Enum) > 0 {
		doc["enum"] = s.Enum
	}
	if s.Items != nil {
		doc["items"] = schemaToD(*s.Items)
	}
	if s.Default != nil {
		doc["default"] = s.Default
	}
	return doc
}

// extractUsage extracts TokenUsage from kronk response.
func extractUsage(chat llmChat, resp model.ChatResponse) core.TokenUsage {
	usage := core.TokenUsage{}
	if chat != nil {
		usage.ContextWindow = chat.ModelConfig().ContextWindow()
	}
	if resp.Usage == nil {
		return usage
	}

	usage.PromptTokens = resp.Usage.PromptTokens
	usage.ReasoningTokens = resp.Usage.ReasoningTokens
	usage.OutputTokens = resp.Usage.OutputTokens
	usage.ContextTokens = resp.Usage.PromptTokens + resp.Usage.CompletionTokens
	usage.TokensPerSecond = resp.Usage.TokensPerSecond
	return usage
}
