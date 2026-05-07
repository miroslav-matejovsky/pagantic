package llm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
)

// StreamResponse processes a streaming chat response from the LLM.
// It extracts tool calls and content tokens, updates message history,
// and returns a ChatResult. Terminal rendering is handled separately
// via the optional StreamHandler.
//
// handler receives typed streaming events. Pass nil for silent operation
// (e.g., structured plan generation).
func StreamResponse(engine Chat, messages []model.D, ch <-chan model.ChatResponse, handler *StreamHandler) (ChatResult, error) {
	var reasoning bool
	var lr model.ChatResponse
	var content strings.Builder
	var toolCalls []ToolCallInfo

loop:
	for resp := range ch {
		lr = resp

		if len(resp.Choices) == 0 {
			continue
		}

		switch resp.Choices[0].FinishReason() {
		case model.FinishReasonError:
			return ChatResult{}, fmt.Errorf("model error: %s", resp.Choices[0].Delta.Content)

		case model.FinishReasonStop:
			if resp.Choices[0].Delta != nil && resp.Choices[0].Delta.Content != "" {
				if reasoning {
					handler.emitReasoning("\n")
				}
				content.WriteString(resp.Choices[0].Delta.Content)
				handler.emitContent(resp.Choices[0].Delta.Content)
			}
			break loop

		case model.FinishReasonTool:
			if reasoning {
				handler.emitReasoning("\n")
			}

			var toolCallDocs []model.D
			for _, tool := range resp.Choices[0].Delta.ToolCalls {
				argsJSON, _ := json.Marshal(tool.Function.Arguments)

				handler.emitToolCall(tool.Function.Name, string(argsJSON))

				toolCallDocs = append(toolCallDocs, model.D{
					"id":   tool.ID,
					"type": "function",
					"function": model.D{
						"name":      tool.Function.Name,
						"arguments": string(argsJSON),
					},
				})
				toolCalls = append(toolCalls, ToolCallInfo{
					ID:        tool.ID,
					Name:      tool.Function.Name,
					Arguments: map[string]any(tool.Function.Arguments),
				})
			}

			assistantMsg := model.D{
				"role":       "assistant",
				"tool_calls": toolCallDocs,
			}
			if content.Len() > 0 {
				assistantMsg["content"] = content.String()
				content.Reset()
			}
			messages = append(messages, assistantMsg)
			break loop

		default:
			if resp.Choices[0].Delta.Reasoning != "" {
				handler.emitReasoning(resp.Choices[0].Delta.Reasoning)
				reasoning = true
				continue
			}

			if reasoning {
				reasoning = false
				handler.emitReasoning("\n")
			}

			content.WriteString(resp.Choices[0].Delta.Content)
			handler.emitContent(resp.Choices[0].Delta.Content)
		}
	}

	if content.Len() > 0 {
		messages = append(messages, model.TextMessage(model.RoleAssistant, content.String()))
	}

	usage := extractUsage(engine, lr)

	return ChatResult{
		ToolCalls: toolCalls,
		Content:   content.String(),
		Messages:  messages,
		Usage:     usage,
	}, nil
}

func extractUsage(engine Chat, lr model.ChatResponse) TokenUsage {
	contextTokens := lr.Usage.PromptTokens + lr.Usage.CompletionTokens
	return TokenUsage{
		PromptTokens:    lr.Usage.PromptTokens,
		ReasoningTokens: lr.Usage.ReasoningTokens,
		OutputTokens:    lr.Usage.OutputTokens,
		ContextTokens:   contextTokens,
		ContextWindow:   engine.ModelConfig().ContextWindow(),
		TokensPerSecond: lr.Usage.TokensPerSecond,
	}
}
