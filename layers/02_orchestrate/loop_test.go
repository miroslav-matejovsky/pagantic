package orchestrate

import (
	"context"
	"errors"
	"testing"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	tool "github.com/miroslav-matejovsky/pagantic/layers/04_tool"
	"github.com/stretchr/testify/require"
)

// fakeEngine replays fixed inference results.
type fakeEngine struct {
	responses []inference.Result
	calls     []inference.Request
	callIndex int
}

func (f *fakeEngine) Infer(_ context.Context, req inference.Request) (*inference.Result, error) {
	f.calls = append(f.calls, cloneTestRequest(req))
	if f.callIndex >= len(f.responses) {
		result := inference.Result{Messages: cloneMessages(req.Messages)}
		return &result, nil
	}

	result := cloneTestResult(f.responses[f.callIndex])
	f.callIndex++
	if len(result.Messages) == 0 {
		result.Messages = resolveMessages(req.Messages, &result)
	}
	return &result, nil
}

func (f *fakeEngine) ModelInfo() inference.ModelInfo {
	return inference.ModelInfo{Name: "fake"}
}

func resolveMessages(requestMessages []core.Message, result *inference.Result) []core.Message {
	if result != nil && len(result.Messages) > 0 {
		return cloneMessages(result.Messages)
	}

	messages := cloneMessages(requestMessages)
	if result == nil {
		return messages
	}

	assistant := core.Message{
		Role:      core.RoleAssistant,
		Content:   result.Content,
		ToolCalls: cloneToolCalls(result.ToolCalls),
	}
	if assistant.Content != "" || len(assistant.ToolCalls) > 0 {
		messages = append(messages, assistant)
	}
	return messages
}

// fakeTool records calls and returns fixed output.
type fakeTool struct {
	definition core.ToolDefinition
	result     string
	err        error
	called     []string
	args       []map[string]any
}

func (f *fakeTool) Info() tool.ToolInfo {
	return tool.ToolInfo{
		Name:        f.definition.Name,
		Type:        tool.TypeGo,
		Description: f.definition.Description,
	}
}

func (f *fakeTool) Definition() core.ToolDefinition {
	return f.definition
}

func (f *fakeTool) Execute(args map[string]any) (string, error) {
	f.called = append(f.called, f.definition.Name)
	f.args = append(f.args, cloneArgs(args))
	if f.err != nil {
		return "", f.err
	}
	if f.result != "" {
		return f.result, nil
	}
	return "ok", nil
}

func (f *fakeTool) Available() (bool, string) {
	return true, ""
}

func TestAgentLoop_Chat_NoTools(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{{Content: "hello there"}},
	}
	loop, err := NewAgentLoop(LoopConfig{
		SystemPrompt:      "sys",
		Engine:            eng,
		MaxTokens:         2048,
		MaxToolIterations: 20,
	})
	require.NoError(t, err)

	result, err := loop.Chat(context.Background(), "hi")
	require.NoError(t, err)
	require.Equal(t, "hello there", result.Content)
	require.Empty(t, result.ToolCalls)
	require.Len(t, eng.calls, 1)
	require.Empty(t, eng.calls[0].Tools)
}

func TestAgentLoop_Chat_WithTools(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{
			{ToolCalls: []core.ToolCall{{ID: "tc1", Name: "do_thing", Arguments: map[string]any{"x": "1"}}}},
			{Content: "done"},
		},
	}
	toolDef := &fakeTool{
		definition: core.ToolDefinition{Name: "do_thing"},
		result:     "result of do_thing",
	}
	callbackFired := false
	loop, err := NewAgentLoop(LoopConfig{
		SystemPrompt:      "sys",
		Engine:            eng,
		Tools:             tool.NewRegistry(toolDef),
		MaxTokens:         2048,
		MaxToolIterations: 20,
		OnToolResult: func(name, output string) {
			callbackFired = true
			require.Equal(t, "do_thing", name)
			require.Equal(t, "result of do_thing", output)
		},
	})
	require.NoError(t, err)

	result, err := loop.Chat(context.Background(), "go")
	require.NoError(t, err)
	require.Equal(t, "done", result.Content)
	require.Equal(t, []string{"do_thing"}, toolDef.called)
	require.Equal(t, []map[string]any{{"x": "1"}}, toolDef.args)
	require.True(t, callbackFired)
	require.Len(t, eng.calls, 2)

	toolMsg := eng.calls[1].Messages[len(eng.calls[1].Messages)-1]
	require.Equal(t, core.RoleTool, toolMsg.Role)
	require.Equal(t, "result of do_thing", toolMsg.Content)
}

func TestAgentLoop_Chat_ToolError(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{
			{ToolCalls: []core.ToolCall{{ID: "tc1", Name: "bad_tool"}}},
			{Content: "recovered"},
		},
	}
	toolDef := &fakeTool{
		definition: core.ToolDefinition{Name: "bad_tool"},
		err:        errors.New("boom"),
	}
	loop, err := NewAgentLoop(LoopConfig{
		SystemPrompt:      "sys",
		Engine:            eng,
		Tools:             tool.NewRegistry(toolDef),
		MaxTokens:         2048,
		MaxToolIterations: 20,
	})
	require.NoError(t, err)

	result, err := loop.Chat(context.Background(), "go")
	require.NoError(t, err)
	require.Equal(t, "recovered", result.Content)
	require.Equal(t, []string{"bad_tool"}, toolDef.called)
	require.Len(t, eng.calls, 2)

	toolMsg := eng.calls[1].Messages[len(eng.calls[1].Messages)-1]
	require.Equal(t, core.RoleTool, toolMsg.Role)
	require.Equal(t, "Error: boom", toolMsg.Content)
}

func TestAgentLoop_ChatStructured(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{{Content: `{"key":"val"}`}},
	}
	loop, err := NewAgentLoop(LoopConfig{
		SystemPrompt:      "sys",
		Engine:            eng,
		MaxTokens:         2048,
		MaxToolIterations: 20,
	})
	require.NoError(t, err)

	schema := core.Schema{
		Type: "object",
		Properties: map[string]core.Schema{
			"key": {Type: "string"},
		},
	}
	result, err := loop.ChatStructured(context.Background(), "produce json", schema)
	require.NoError(t, err)
	require.Equal(t, `{"key":"val"}`, result.Content)
	require.Len(t, eng.calls, 1)
	require.NotNil(t, eng.calls[0].Schema)
	require.NotNil(t, eng.calls[0].Temperature)
	require.InDelta(t, 0.3, *eng.calls[0].Temperature, 0.0001)
	require.Equal(t, false, eng.calls[0].Options["enable_thinking"])
	require.Len(t, eng.calls[0].Messages, 2)
}

func TestAgentLoop_ChatStructured_UsesAccumulatedHistory(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{
			{Content: "collected data"},
			{Content: `{"key":"val"}`},
		},
	}
	loop, err := NewAgentLoop(LoopConfig{
		SystemPrompt:      "sys",
		Engine:            eng,
		MaxTokens:         2048,
		MaxToolIterations: 20,
	})
	require.NoError(t, err)

	_, err = loop.Chat(context.Background(), "collect data")
	require.NoError(t, err)

	result, err := loop.ChatStructured(context.Background(), "now produce json", core.Schema{Type: "object"})
	require.NoError(t, err)
	require.Equal(t, `{"key":"val"}`, result.Content)
	require.Len(t, eng.calls, 2)
	require.Len(t, eng.calls[1].Messages, 4)
	require.NotNil(t, eng.calls[1].Schema)
}

func TestAgentLoop_New_Validation(t *testing.T) {
	eng := &fakeEngine{}

	_, err := NewAgentLoop(LoopConfig{MaxTokens: 2048, MaxToolIterations: 20})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Engine")

	_, err = NewAgentLoop(LoopConfig{Engine: eng, MaxToolIterations: 20})
	require.Error(t, err)
	require.Contains(t, err.Error(), "MaxTokens")

	_, err = NewAgentLoop(LoopConfig{Engine: eng, MaxTokens: 2048})
	require.Error(t, err)
	require.Contains(t, err.Error(), "MaxToolIterations")
}

func TestAgentLoop_Chat_MaxTokens_PassedToEngine(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{{Content: "hi"}},
	}
	loop, err := NewAgentLoop(LoopConfig{
		SystemPrompt:      "sys",
		Engine:            eng,
		MaxTokens:         4096,
		MaxToolIterations: 20,
	})
	require.NoError(t, err)

	_, err = loop.Chat(context.Background(), "hey")
	require.NoError(t, err)
	require.Equal(t, 4096, eng.calls[0].MaxTokens)
}

func TestAgentLoop_Chat_MaxToolIterations(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{
			{ToolCalls: []core.ToolCall{{ID: "tc1", Name: "loop_tool"}}},
			{ToolCalls: []core.ToolCall{{ID: "tc2", Name: "loop_tool"}}},
			{ToolCalls: []core.ToolCall{{ID: "tc3", Name: "loop_tool"}}},
			{ToolCalls: []core.ToolCall{{ID: "tc4", Name: "loop_tool"}}},
		},
	}
	toolDef := &fakeTool{
		definition: core.ToolDefinition{Name: "loop_tool"},
		result:     "looping",
	}
	loop, err := NewAgentLoop(LoopConfig{
		Engine:            eng,
		Tools:             tool.NewRegistry(toolDef),
		MaxTokens:         2048,
		MaxToolIterations: 3,
	})
	require.NoError(t, err)

	_, err = loop.Chat(context.Background(), "go")
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeded max tool iterations")
}

func TestAgentLoop_ChatStructured_SchemaValidation(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{{Content: `{"name":"grog"}`}},
	}
	loop, err := NewAgentLoop(LoopConfig{Engine: eng, MaxTokens: 2048, MaxToolIterations: 20})
	require.NoError(t, err)

	schema := core.Schema{
		Type: "object",
		Properties: map[string]core.Schema{
			"name": {Type: "string"},
			"age":  {Type: "number"},
		},
		Required: []string{"name", "age"},
	}
	_, err = loop.ChatStructured(context.Background(), "produce json", schema)
	require.Error(t, err)
	require.Contains(t, err.Error(), "schema validation failed")
}

func cloneTestRequest(req inference.Request) inference.Request {
	cloned := req
	cloned.Messages = cloneMessages(req.Messages)
	cloned.Tools = append([]core.ToolDefinition(nil), req.Tools...)
	if req.Schema != nil {
		schema := *req.Schema
		cloned.Schema = &schema
	}
	if req.Temperature != nil {
		temperature := *req.Temperature
		cloned.Temperature = &temperature
	}
	if req.Options != nil {
		cloned.Options = make(map[string]any, len(req.Options))
		for key, value := range req.Options {
			cloned.Options[key] = value
		}
	}
	return cloned
}

func cloneTestResult(result inference.Result) inference.Result {
	cloned := result
	cloned.Messages = cloneMessages(result.Messages)
	cloned.ToolCalls = cloneToolCalls(result.ToolCalls)
	return cloned
}

var _ inference.Engine = (*fakeEngine)(nil)
var _ tool.Tool = (*fakeTool)(nil)

// fakeContextProvider returns fixed context messages.
type fakeContextProvider struct {
	messages []core.Message
	err      error
	queries  []string
}

func (f *fakeContextProvider) Build(_ context.Context, query string) ([]core.Message, error) {
	f.queries = append(f.queries, query)
	if f.err != nil {
		return nil, f.err
	}
	return f.messages, nil
}

func TestAgentLoop_Chat_WithContext(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{{Content: "answer with context"}},
	}
	provider := &fakeContextProvider{
		messages: []core.Message{core.NewSystemMessage("Relevant context:\n\n[1] (doc): bounded knowledge")},
	}
	loop, err := NewAgentLoop(LoopConfig{
		SystemPrompt:      "sys",
		Engine:            eng,
		ContextProvider:   provider,
		MaxTokens:         2048,
		MaxToolIterations: 20,
	})
	require.NoError(t, err)

	result, err := loop.Chat(context.Background(), "what is X?")
	require.NoError(t, err)
	require.Equal(t, "answer with context", result.Content)
	require.Equal(t, []string{"what is X?"}, provider.queries)

	// Context message appears between system prompt and user message.
	msgs := eng.calls[0].Messages
	require.Equal(t, core.RoleSystem, msgs[0].Role)
	require.Equal(t, "sys", msgs[0].Content)
	require.Equal(t, core.RoleSystem, msgs[1].Role)
	require.Contains(t, msgs[1].Content, "bounded knowledge")
	require.Equal(t, core.RoleUser, msgs[2].Role)
	require.Equal(t, "what is X?", msgs[2].Content)
}

func TestAgentLoop_Chat_NilContextProvider(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{{Content: "no context"}},
	}
	loop, err := NewAgentLoop(LoopConfig{
		SystemPrompt:      "sys",
		Engine:            eng,
		MaxTokens:         2048,
		MaxToolIterations: 20,
	})
	require.NoError(t, err)

	result, err := loop.Chat(context.Background(), "hi")
	require.NoError(t, err)
	require.Equal(t, "no context", result.Content)

	// Only system + user messages, no context.
	msgs := eng.calls[0].Messages
	require.Len(t, msgs, 2)
	require.Equal(t, core.RoleSystem, msgs[0].Role)
	require.Equal(t, core.RoleUser, msgs[1].Role)
}

func TestAgentLoop_Chat_ContextErrorContinuesWithout(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{{Content: "still works"}},
	}
	provider := &fakeContextProvider{
		err: errors.New("retrieval failed"),
	}
	loop, err := NewAgentLoop(LoopConfig{
		SystemPrompt:      "sys",
		Engine:            eng,
		ContextProvider:   provider,
		MaxTokens:         2048,
		MaxToolIterations: 20,
	})
	require.NoError(t, err)

	result, err := loop.Chat(context.Background(), "hi")
	require.NoError(t, err)
	require.Equal(t, "still works", result.Content)

	// No context messages injected on error.
	msgs := eng.calls[0].Messages
	require.Len(t, msgs, 2)
}

func TestAgentLoop_Chat_ContextRetrievedEveryTurn(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{
			{Content: "first answer"},
			{Content: "second answer"},
		},
	}
	provider := &fakeContextProvider{
		messages: []core.Message{core.NewSystemMessage("ctx")},
	}
	loop, err := NewAgentLoop(LoopConfig{
		SystemPrompt:      "sys",
		Engine:            eng,
		ContextProvider:   provider,
		MaxTokens:         2048,
		MaxToolIterations: 20,
	})
	require.NoError(t, err)

	_, err = loop.Chat(context.Background(), "q1")
	require.NoError(t, err)
	_, err = loop.Chat(context.Background(), "q2")
	require.NoError(t, err)

	require.Equal(t, []string{"q1", "q2"}, provider.queries)

	// Context is ephemeral: second inference call must NOT include "ctx" from turn 1.
	// Expected: [sys, ctx, q1, assistant, ctx, q2] - ephemeral ctx appears per turn.
	// The buffer (memory) must NOT grow with accumulated ctx messages.
	// Second call req.Messages: [sys, q1-ans-in-history, ctx, q2].
	secondCallMsgs := eng.calls[1].Messages
	ctxCount := 0
	for _, m := range secondCallMsgs {
		if m.Content == "ctx" {
			ctxCount++
		}
	}
	require.Equal(t, 1, ctxCount, "exactly one ctx message per turn, none accumulated from prior turn")
}
