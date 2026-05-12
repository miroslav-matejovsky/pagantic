package orchestrate

import (
	"context"
	"testing"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	tool "github.com/miroslav-matejovsky/pagantic/layers/04_tool"
	"github.com/stretchr/testify/require"
)

func TestSpecializedLoop_Call_NoTools(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{{Content: `[{"email":"a@b.com"}]`}},
	}
	loop, err := NewSpecializedLoop(SpecializedConfig{
		SystemPrompt:      "sys",
		Engine:            eng,
		Schema:            core.Schema{Type: "array"},
		MaxTokens:         2048,
		MaxToolIterations: 20,
	})
	require.NoError(t, err)

	result, err := loop.Call(context.Background(), "analyze")
	require.NoError(t, err)
	require.Equal(t, `[{"email":"a@b.com"}]`, result.Content)
	require.Len(t, eng.calls, 1)
	require.NotNil(t, eng.calls[0].Schema)
}

func TestSpecializedLoop_Call_WithTools(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{
			{ToolCalls: []core.ToolCall{{ID: "tc1", Name: "fetch_data"}}},
			{Content: "collected the data"},
			{Content: `[{"email":"x@y.com","complexity":"low","quality_notes":"ok","commit_frequency":"daily"}]`},
		},
	}
	toolDef := &fakeTool{
		definition: core.ToolDefinition{Name: "fetch_data"},
		result:     `[{"name":"X","email":"x@y.com","total_commits":5}]`,
	}
	loop, err := NewSpecializedLoop(SpecializedConfig{
		SystemPrompt:      "sys",
		Engine:            eng,
		Schema:            core.Schema{Type: "array"},
		Tools:             tool.NewRegistry(toolDef),
		MaxTokens:         2048,
		MaxToolIterations: 20,
	})
	require.NoError(t, err)

	result, err := loop.Call(context.Background(), "analyze repo")
	require.NoError(t, err)
	require.Equal(t, []string{"fetch_data"}, toolDef.called)
	require.Len(t, eng.calls, 3)
	require.NotNil(t, eng.calls[2].Schema)
	require.Greater(t, len(eng.calls[2].Messages), 2)
	require.Contains(t, result.Content, "x@y.com")
}

func TestSpecializedLoop_New_Validation(t *testing.T) {
	eng := &fakeEngine{}
	schema := core.Schema{Type: "object"}

	_, err := NewSpecializedLoop(SpecializedConfig{Schema: schema, MaxTokens: 2048, MaxToolIterations: 20})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Engine")

	_, err = NewSpecializedLoop(SpecializedConfig{Engine: eng, Schema: schema, MaxToolIterations: 20})
	require.Error(t, err)
	require.Contains(t, err.Error(), "MaxTokens")

	_, err = NewSpecializedLoop(SpecializedConfig{Engine: eng, Schema: schema, MaxTokens: 2048})
	require.Error(t, err)
	require.Contains(t, err.Error(), "MaxToolIterations")
}

func TestSpecializedLoop_Call_WithContext_NoTools(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{{Content: `{"answer":"yes"}`}},
	}
	provider := &fakeContextProvider{
		messages: []core.Message{core.NewSystemMessage("context info")},
	}
	loop, err := NewSpecializedLoop(SpecializedConfig{
		SystemPrompt:      "sys",
		Engine:            eng,
		Schema:            core.Schema{Type: "object"},
		ContextProvider:   provider,
		MaxTokens:         2048,
		MaxToolIterations: 20,
	})
	require.NoError(t, err)

	result, err := loop.Call(context.Background(), "analyze")
	require.NoError(t, err)
	require.Equal(t, `{"answer":"yes"}`, result.Content)
	require.Equal(t, []string{"analyze"}, provider.queries)

	// Context injected into messages before user prompt.
	msgs := eng.calls[0].Messages
	require.Equal(t, core.RoleSystem, msgs[0].Role)
	require.Equal(t, "sys", msgs[0].Content)
	require.Equal(t, core.RoleSystem, msgs[1].Role)
	require.Equal(t, "context info", msgs[1].Content)
}

func TestSpecializedLoop_Call_WithContext_WithTools(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{
			{ToolCalls: []core.ToolCall{{ID: "tc1", Name: "fetch"}}},
			{Content: "fetched"},
			{Content: `{"result":"ok"}`},
		},
	}
	toolDef := &fakeTool{
		definition: core.ToolDefinition{Name: "fetch"},
		result:     "data",
	}
	provider := &fakeContextProvider{
		messages: []core.Message{core.NewSystemMessage("context info")},
	}
	loop, err := NewSpecializedLoop(SpecializedConfig{
		SystemPrompt:      "sys",
		Engine:            eng,
		Schema:            core.Schema{Type: "object"},
		Tools:             tool.NewRegistry(toolDef),
		ContextProvider:   provider,
		MaxTokens:         2048,
		MaxToolIterations: 20,
	})
	require.NoError(t, err)

	result, err := loop.Call(context.Background(), "analyze")
	require.NoError(t, err)
	require.Equal(t, `{"result":"ok"}`, result.Content)
	// Context retrieved once using original prompt, not phase2Prompt.
	require.Equal(t, []string{"analyze"}, provider.queries)
}
