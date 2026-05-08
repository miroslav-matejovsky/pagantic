package orchestrate

import (
	"context"
	"errors"
	"testing"

	"github.com/miroslav-matejovsky/pagantic/core"
	"github.com/miroslav-matejovsky/pagantic/inference"
	"github.com/miroslav-matejovsky/pagantic/tool"
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
	loop := NewAgentLoop(LoopConfig{
		SystemPrompt: "sys",
		Engine:       eng,
	})

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
	loop := NewAgentLoop(LoopConfig{
		SystemPrompt: "sys",
		Engine:       eng,
		Tools:        tool.NewRegistry(toolDef),
		OnToolResult: func(name, output string) {
			callbackFired = true
			require.Equal(t, "do_thing", name)
			require.Equal(t, "result of do_thing", output)
		},
	})

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
	loop := NewAgentLoop(LoopConfig{
		SystemPrompt: "sys",
		Engine:       eng,
		Tools:        tool.NewRegistry(toolDef),
	})

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
	loop := NewAgentLoop(LoopConfig{
		SystemPrompt: "sys",
		Engine:       eng,
	})

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
	loop := NewAgentLoop(LoopConfig{
		SystemPrompt: "sys",
		Engine:       eng,
	})

	_, err := loop.Chat(context.Background(), "collect data")
	require.NoError(t, err)

	result, err := loop.ChatStructured(context.Background(), "now produce json", core.Schema{Type: "object"})
	require.NoError(t, err)
	require.Equal(t, `{"key":"val"}`, result.Content)
	require.Len(t, eng.calls, 2)
	require.Len(t, eng.calls[1].Messages, 4)
	require.NotNil(t, eng.calls[1].Schema)
}

func TestAgentLoop_MaxTokens_Default(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{{Content: "hi"}},
	}
	loop := NewAgentLoop(LoopConfig{
		SystemPrompt: "sys",
		Engine:       eng,
	})

	_, err := loop.Chat(context.Background(), "hey")
	require.NoError(t, err)
	require.Equal(t, 2048, eng.calls[0].MaxTokens)
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
	loop := NewAgentLoop(LoopConfig{
		Engine:            eng,
		Tools:             tool.NewRegistry(toolDef),
		MaxToolIterations: 3,
	})

	_, err := loop.Chat(context.Background(), "go")
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeded max tool iterations")
}

func TestAgentLoop_ChatStructured_SchemaValidation(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{{Content: `{"name":"grog"}`}},
	}
	loop := NewAgentLoop(LoopConfig{Engine: eng})

	schema := core.Schema{
		Type: "object",
		Properties: map[string]core.Schema{
			"name": {Type: "string"},
			"age":  {Type: "number"},
		},
		Required: []string{"name", "age"},
	}
	_, err := loop.ChatStructured(context.Background(), "produce json", schema)
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
