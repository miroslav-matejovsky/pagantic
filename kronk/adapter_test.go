package kronk

import (
	"context"
	"testing"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	"github.com/stretchr/testify/require"
)

type mockChat struct {
	request   model.D
	responses []model.ChatResponse
	cfg       model.Config
}

func (m *mockChat) ChatStreaming(_ context.Context, d model.D) (<-chan model.ChatResponse, error) {
	m.request = d
	ch := make(chan model.ChatResponse, len(m.responses))
	for _, resp := range m.responses {
		ch <- resp
	}
	close(ch)
	return ch, nil
}

func (m *mockChat) ModelConfig() model.Config { return m.cfg }

func stringPtr(s string) *string { return &s }

func intPtr(n int) *int { return &n }

func TestAdapterInferContentOnly(t *testing.T) {
	temp := 0.2
	chat := &mockChat{
		cfg: model.Config{
			PtrContextWindow: intPtr(8192),
			ModelFiles:       []string{"models\\test-model.gguf"},
		},
		responses: []model.ChatResponse{
			{Choices: []model.Choice{{Delta: &model.ResponseMessage{Content: "hel"}}}},
			{
				Choices: []model.Choice{{
					Delta:           &model.ResponseMessage{Content: "lo"},
					FinishReasonPtr: stringPtr(model.FinishReasonStop),
				}},
				Usage: &model.Usage{
					PromptTokens:     10,
					ReasoningTokens:  2,
					CompletionTokens: 5,
					OutputTokens:     3,
					TokensPerSecond:  12.5,
				},
			},
		},
	}

	adapter := NewAdapter(chat, nil)
	schema := &core.Schema{
		Type: "object",
		Properties: map[string]core.Schema{
			"value": {Type: "string"},
		},
		Required: []string{"value"},
	}
	result, err := adapter.Infer(context.Background(), inference.Request{
		Messages: []core.Message{
			core.NewSystemMessage("sys"),
			core.NewUserMessage("hello"),
		},
		Schema:      schema,
		MaxTokens:   64,
		Temperature: &temp,
		Options: map[string]any{
			"top_p": 0.9,
		},
	})

	require.NoError(t, err)
	require.Equal(t, "hello", result.Content)
	require.Empty(t, result.ToolCalls)
	require.Len(t, result.Messages, 3)
	require.Equal(t, core.RoleAssistant, result.Messages[2].Role)
	require.Equal(t, "hello", result.Messages[2].Content)
	require.Equal(t, core.TokenUsage{
		PromptTokens:    10,
		ReasoningTokens: 2,
		OutputTokens:    3,
		ContextTokens:   15,
		ContextWindow:   8192,
		TokensPerSecond: 12.5,
	}, result.Usage)

	require.Equal(t, 64, chat.request["max_tokens"])
	require.Equal(t, temp, chat.request["temperature"])
	require.Equal(t, 0.9, chat.request["top_p"])
	require.Equal(t, false, chat.request["enable_thinking"])
	require.NotNil(t, chat.request["json_schema"])

	msgs, ok := chat.request["messages"].([]model.D)
	require.True(t, ok)
	require.Len(t, msgs, 2)
	require.Equal(t, "system", msgs[0]["role"])
	require.Equal(t, "sys", msgs[0]["content"])
	require.Equal(t, "user", msgs[1]["role"])
	require.Equal(t, "hello", msgs[1]["content"])
}

func TestAdapterInferToolCall(t *testing.T) {
	chat := &mockChat{
		responses: []model.ChatResponse{
			{
				Choices: []model.Choice{{
					Delta: &model.ResponseMessage{
						ToolCalls: []model.ResponseToolCall{{
							ID:   "call-1",
							Type: "function",
							Function: model.ResponseToolCallFunction{
								Name:      "search",
								Arguments: model.ToolCallArguments{"query": "moon"},
							},
						}},
					},
					FinishReasonPtr: stringPtr(model.FinishReasonTool),
				}},
				Usage: &model.Usage{PromptTokens: 4, CompletionTokens: 2, OutputTokens: 1},
			},
		},
	}

	adapter := NewAdapter(chat, nil)
	result, err := adapter.Infer(context.Background(), inference.Request{
		Messages: []core.Message{core.NewUserMessage("find moon")},
		Tools: []core.ToolDefinition{{
			Name:        "search",
			Description: "Search docs",
			Parameters: core.Schema{
				Type: "object",
				Properties: map[string]core.Schema{
					"query": {Type: "string"},
				},
				Required: []string{"query"},
			},
		}},
	})

	require.NoError(t, err)
	require.Empty(t, result.Content)
	require.Len(t, result.ToolCalls, 1)
	require.Equal(t, core.ToolCall{ID: "call-1", Name: "search", Arguments: map[string]any{"query": "moon"}}, result.ToolCalls[0])
	require.Len(t, result.Messages, 2)
	require.Equal(t, core.RoleAssistant, result.Messages[1].Role)
	require.Len(t, result.Messages[1].ToolCalls, 1)
	require.Equal(t, "search", result.Messages[1].ToolCalls[0].Name)

	tools, ok := chat.request["tools"].([]model.D)
	require.True(t, ok)
	require.Len(t, tools, 1)
	require.Equal(t, "function", tools[0]["type"])
	function, ok := tools[0]["function"].(model.D)
	require.True(t, ok)
	require.Equal(t, "search", function["name"])
}

func TestAdapterInferStreamHandler(t *testing.T) {
	chat := &mockChat{
		responses: []model.ChatResponse{
			{Choices: []model.Choice{{Delta: &model.ResponseMessage{Reasoning: "think"}}}},
			{Choices: []model.Choice{{Delta: &model.ResponseMessage{Content: "he"}}}},
			{
				Choices: []model.Choice{{
					Delta:           &model.ResponseMessage{Content: "y"},
					FinishReasonPtr: stringPtr(model.FinishReasonStop),
				}},
				Usage: &model.Usage{},
			},
		},
	}

	var reasoning []string
	var content []string
	adapter := NewAdapter(chat, &inference.StreamHandler{
		OnReasoning: func(text string) {
			reasoning = append(reasoning, text)
		},
		OnContent: func(text string) {
			content = append(content, text)
		},
	})

	result, err := adapter.Infer(context.Background(), inference.Request{
		Messages: []core.Message{core.NewUserMessage("hi")},
	})

	require.NoError(t, err)
	require.Equal(t, "hey", result.Content)
	require.Equal(t, []string{"think", "\n"}, reasoning)
	require.Equal(t, []string{"he", "y"}, content)
}

func TestAdapterModelInfo(t *testing.T) {
	adapter := NewAdapter(&mockChat{cfg: model.Config{
		PtrContextWindow: intPtr(32000),
		ModelFiles:       []string{"C:\\models\\qwen.gguf"},
	}}, nil)

	require.Equal(t, inference.ModelInfo{
		Name:          "qwen.gguf",
		ContextWindow: 32000,
	}, adapter.ModelInfo())
}

func TestMessagesToDSanity(t *testing.T) {
	msgs := []core.Message{
		core.NewSystemMessage("sys"),
		core.NewUserMessage("user"),
		{
			Role:    core.RoleAssistant,
			Content: "partial",
			ToolCalls: []core.ToolCall{{
				ID:        "call-1",
				Name:      "search",
				Arguments: map[string]any{"query": "moon"},
			}},
		},
		core.NewToolResultMessage("call-1", "search", "done"),
	}

	docs := messagesToD(msgs)

	require.Len(t, docs, 4)
	require.Equal(t, "system", docs[0]["role"])
	require.Equal(t, "sys", docs[0]["content"])
	require.Equal(t, "user", docs[1]["role"])
	require.Equal(t, "assistant", docs[2]["role"])
	require.Equal(t, "partial", docs[2]["content"])
	toolCalls, ok := docs[2]["tool_calls"].([]model.D)
	require.True(t, ok)
	require.Len(t, toolCalls, 1)
	function, ok := toolCalls[0]["function"].(model.D)
	require.True(t, ok)
	require.Equal(t, "search", function["name"])
	require.Equal(t, `{"query":"moon"}`, function["arguments"])
	require.Equal(t, "tool", docs[3]["role"])
	require.Equal(t, "call-1", docs[3]["tool_call_id"])
	require.Equal(t, "search", docs[3]["name"])
	require.Equal(t, "done", docs[3]["content"])
}
