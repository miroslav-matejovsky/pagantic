package agent_test

import (
	"context"
	"testing"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
	"github.com/miroslav-matejovsky/pagantic/agent"
	"github.com/miroslav-matejovsky/pagantic/llm"
	"github.com/stretchr/testify/require"
)

// fakeEngine is a mock llm.Chat that plays back pre-configured response sequences.
// Each ChatStreaming call consumes the next sequence from responses.
type fakeEngine struct {
	responses [][]model.ChatResponse
	calls     []model.D
}

func (e *fakeEngine) ChatStreaming(_ context.Context, d model.D) (<-chan model.ChatResponse, error) {
	e.calls = append(e.calls, d)
	ch := make(chan model.ChatResponse, 8)
	if len(e.responses) > 0 {
		for _, r := range e.responses[0] {
			ch <- r
		}
		e.responses = e.responses[1:]
	}
	close(ch)
	return ch, nil
}

func (e *fakeEngine) ModelConfig() model.Config { return model.Config{} }

// fakeTools is a mock agent.ToolProvider that records Execute calls.
type fakeTools struct {
	defs    []model.D
	results map[string]string
	called  []string
}

func (f *fakeTools) Definitions() []model.D { return f.defs }

func (f *fakeTools) Execute(name string, _ map[string]any) (string, error) {
	f.called = append(f.called, name)
	if r, ok := f.results[name]; ok {
		return r, nil
	}
	return "ok", nil
}

// helpers to build model.ChatResponse values.

func strPtr(s string) *string { return &s }

// contentResp sends a content token then a stop signal - simulates a streaming content response.
func contentResp(content string) []model.ChatResponse {
	usage := &model.Usage{}
	return []model.ChatResponse{
		{Choices: []model.Choice{
			{Delta: &model.ResponseMessage{Content: content}},
		}},
		{
			Choices: []model.Choice{{FinishReasonPtr: strPtr(model.FinishReasonStop)}},
			Usage:   usage,
		},
	}
}

// toolCallResp simulates the LLM requesting a tool call.
func toolCallResp(id, name string, args map[string]any) []model.ChatResponse {
	usage := &model.Usage{}
	return []model.ChatResponse{
		{
			Choices: []model.Choice{
				{
					Delta: &model.ResponseMessage{
						ToolCalls: []model.ResponseToolCall{
							{
								ID:   id,
								Type: "function",
								Function: model.ResponseToolCallFunction{
									Name:      name,
									Arguments: model.ToolCallArguments(args),
								},
							},
						},
					},
					FinishReasonPtr: strPtr(model.FinishReasonTool),
				},
			},
			Usage: usage,
		},
	}
}

// ---- Agent tests ----

func TestAgent_Chat_NoTools(t *testing.T) {
	eng := &fakeEngine{
		responses: [][]model.ChatResponse{contentResp("hello there")},
	}
	a := agent.New(agent.Config{
		SystemPrompt: "sys",
		Engine:       eng,
	})

	result, err := a.Chat(context.Background(), "hi")
	require.NoError(t, err)
	require.Equal(t, "hello there", result.Content)
	require.Empty(t, result.ToolCalls)
	require.Len(t, eng.calls, 1)
	// tools field should NOT be present when no ToolProvider set
	_, hasTools := eng.calls[0]["tools"]
	require.False(t, hasTools)
}

func TestAgent_Chat_WithTools(t *testing.T) {
	eng := &fakeEngine{
		responses: [][]model.ChatResponse{
			toolCallResp("tc1", "do_thing", map[string]any{"x": "1"}),
			contentResp("done"),
		},
	}
	tools := &fakeTools{
		defs:    []model.D{{"type": "function", "function": model.D{"name": "do_thing"}}},
		results: map[string]string{"do_thing": "result of do_thing"},
	}
	callbackFired := false
	a := agent.New(agent.Config{
		SystemPrompt: "sys",
		Engine:       eng,
		Tools:        tools,
		OnToolCall: func(name, output string) {
			callbackFired = true
			require.Equal(t, "do_thing", name)
			require.Equal(t, "result of do_thing", output)
		},
	})

	result, err := a.Chat(context.Background(), "go")
	require.NoError(t, err)
	require.Equal(t, "done", result.Content)
	require.Equal(t, []string{"do_thing"}, tools.called)
	require.True(t, callbackFired, "OnToolCall callback should have been invoked")
	// two LLM calls: one for tool, one for final content
	require.Len(t, eng.calls, 2)
}

func TestAgent_Chat_ToolError(t *testing.T) {
	eng := &fakeEngine{
		responses: [][]model.ChatResponse{
			toolCallResp("tc1", "bad_tool", nil),
			contentResp("recovered"),
		},
	}
	tools := &fakeTools{
		defs:    []model.D{{"type": "function", "function": model.D{"name": "bad_tool"}}},
		results: map[string]string{},
		// no result → returns "ok" (no error path via fakeTools)
	}
	callbackCalls := []struct {
		name, output string
	}{}
	a := agent.New(agent.Config{
		SystemPrompt: "sys",
		Engine:       eng,
		Tools:        tools,
		OnToolCall: func(name, output string) {
			callbackCalls = append(callbackCalls, struct {
				name, output string
			}{name, output})
		},
	})

	result, err := a.Chat(context.Background(), "go")
	require.NoError(t, err)
	require.Equal(t, "recovered", result.Content)
	require.Equal(t, []string{"bad_tool"}, tools.called)
	require.Len(t, callbackCalls, 1)
	require.Equal(t, "bad_tool", callbackCalls[0].name)
	require.Equal(t, "ok", callbackCalls[0].output)
}

func TestAgent_ChatStructured(t *testing.T) {
	eng := &fakeEngine{
		responses: [][]model.ChatResponse{contentResp(`{"key":"val"}`)},
	}
	a := agent.New(agent.Config{
		SystemPrompt: "sys",
		Engine:       eng,
	})

	schema := model.D{"type": "object", "properties": model.D{"key": model.D{"type": "string"}}}
	result, err := a.ChatStructured(context.Background(), "produce json", schema)
	require.NoError(t, err)
	require.Equal(t, `{"key":"val"}`, result.Content)
	// json_schema must be set in the request
	require.Contains(t, eng.calls[0], "json_schema")
	require.Equal(t, false, eng.calls[0]["enable_thinking"])
	// messages: system + user
	msgs := eng.calls[0]["messages"].([]model.D)
	require.Len(t, msgs, 2)
}

func TestAgent_ChatStructured_UsesAccumulatedHistory(t *testing.T) {
	eng := &fakeEngine{
		responses: [][]model.ChatResponse{
			// Phase 1: Chat tool loop produces content
			contentResp("collected data"),
			// Phase 2: ChatStructured produces JSON
			contentResp(`{"key":"val"}`),
		},
	}
	tools := &fakeTools{
		defs: []model.D{{"type": "function", "function": model.D{"name": "fetch"}}},
	}
	a := agent.New(agent.Config{
		SystemPrompt: "sys",
		Engine:       eng,
		Tools:        tools,
	})

	_, err := a.Chat(context.Background(), "collect data")
	require.NoError(t, err)

	schema := model.D{"type": "object"}
	result, err := a.ChatStructured(context.Background(), "now produce json", schema)
	require.NoError(t, err)
	require.Equal(t, `{"key":"val"}`, result.Content)

	// Structured call should include history from the Chat phase (system + user + assistant).
	structuredCall := eng.calls[1]
	msgs := structuredCall["messages"].([]model.D)
	require.Greater(t, len(msgs), 2, "structured call must include accumulated history from Chat")
	require.Contains(t, structuredCall, "json_schema")
}

// TestAgent_MaxTokens_Default checks that zero MaxTokens gets set to the default.
func TestAgent_MaxTokens_Default(t *testing.T) {
	eng := &fakeEngine{
		responses: [][]model.ChatResponse{contentResp("hi")},
	}
	a := agent.New(agent.Config{
		SystemPrompt: "sys",
		Engine:       eng,
	})
	_, err := a.Chat(context.Background(), "hey")
	require.NoError(t, err)
	require.Equal(t, 2048, eng.calls[0]["max_tokens"])
}

// ---- SpecializedAgent tests ----

func TestSpecializedAgent_Call_NoTools(t *testing.T) {
	eng := &fakeEngine{
		responses: [][]model.ChatResponse{contentResp(`[{"email":"a@b.com"}]`)},
	}
	schema := model.D{"type": "array"}
	a := agent.NewSpecialized(agent.SpecializedConfig{
		SystemPrompt: "sys",
		Engine:       eng,
		Schema:       schema,
	})

	result, err := a.Call(context.Background(), "analyze")
	require.NoError(t, err)
	require.Equal(t, `[{"email":"a@b.com"}]`, result.Content)
	require.Len(t, eng.calls, 1)
	require.Contains(t, eng.calls[0], "json_schema")
}

func TestSpecializedAgent_Call_WithTools(t *testing.T) {
	eng := &fakeEngine{
		responses: [][]model.ChatResponse{
			// Phase 1: tool call then content
			toolCallResp("tc1", "fetch_data", nil),
			contentResp("collected the data"),
			// Phase 2: structured output via ChatStructured
			contentResp(`[{"email":"x@y.com","complexity":"low","quality_notes":"ok","commit_frequency":"daily"}]`),
		},
	}
	tools := &fakeTools{
		defs:    []model.D{{"type": "function", "function": model.D{"name": "fetch_data"}}},
		results: map[string]string{"fetch_data": `[{"name":"X","email":"x@y.com","total_commits":5}]`},
	}
	schema := model.D{"type": "array"}
	a := agent.NewSpecialized(agent.SpecializedConfig{
		SystemPrompt: "sys",
		Engine:       eng,
		Schema:       schema,
		Tools:        tools,
	})

	result, err := a.Call(context.Background(), "analyze repo")
	require.NoError(t, err)

	require.Equal(t, []string{"fetch_data"}, tools.called)
	// Phase 1: two LLM calls (tool round + content round), Phase 2: one structured call = 3 total
	require.Len(t, eng.calls, 3)
	// Phase 2 call must have json_schema and include accumulated context
	require.Contains(t, eng.calls[2], "json_schema")
	phase2Msgs := eng.calls[2]["messages"].([]model.D)
	require.Greater(t, len(phase2Msgs), 2, "phase 2 must include history from tool loop")
	require.Contains(t, result.Content, "x@y.com")
}

// TestSpecializedAgent_Call_WithTools_NoToolsUsed verifies that even if the LLM
// does not call any tools, Phase 2 still runs and returns structured output.
func TestSpecializedAgent_Call_WithTools_NoToolsUsed(t *testing.T) {
	eng := &fakeEngine{
		responses: [][]model.ChatResponse{
			// Phase 1: LLM skips tools and goes straight to content
			contentResp("I already know the answer"),
			// Phase 2: structured output
			contentResp(`[]`),
		},
	}
	tools := &fakeTools{
		defs: []model.D{{"type": "function", "function": model.D{"name": "unused"}}},
	}
	schema := model.D{"type": "array"}
	a := agent.NewSpecialized(agent.SpecializedConfig{
		SystemPrompt: "sys",
		Engine:       eng,
		Schema:       schema,
		Tools:        tools,
	})

	result, err := a.Call(context.Background(), "analyze")
	require.NoError(t, err)
	require.Empty(t, tools.called, "no tools should have been called")
	require.Len(t, eng.calls, 2, "Phase 1 + Phase 2 = 2 calls even without tool use")
	require.Equal(t, "[]", result.Content)
}

// TestSpecializedAgent_MaxTokens_Default checks the same default for SpecializedAgent.
func TestSpecializedAgent_MaxTokens_Default(t *testing.T) {
	eng := &fakeEngine{
		responses: [][]model.ChatResponse{contentResp(`{}`)},
	}
	a := agent.NewSpecialized(agent.SpecializedConfig{
		SystemPrompt: "sys",
		Engine:       eng,
		Schema:       model.D{"type": "object"},
	})
	_, err := a.Call(context.Background(), "go")
	require.NoError(t, err)
	require.Equal(t, 2048, eng.calls[0]["max_tokens"])
}

// Compile-time checks.
var _ llm.Chat = (*fakeEngine)(nil)
var _ agent.ToolProvider = (*fakeTools)(nil)
