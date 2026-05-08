package orchestrate

import (
	"context"
	"testing"

	"github.com/miroslav-matejovsky/pagantic/core"
	"github.com/miroslav-matejovsky/pagantic/inference"
	"github.com/miroslav-matejovsky/pagantic/tool"
	"github.com/stretchr/testify/require"
)

func TestSpecializedLoop_Call_NoTools(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{{Content: `[{"email":"a@b.com"}]`}},
	}
	loop := NewSpecializedLoop(SpecializedConfig{
		SystemPrompt: "sys",
		Engine:       eng,
		Schema:       core.Schema{Type: "array"},
	})

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
	loop := NewSpecializedLoop(SpecializedConfig{
		SystemPrompt: "sys",
		Engine:       eng,
		Schema:       core.Schema{Type: "array"},
		Tools:        tool.NewRegistry(toolDef),
	})

	result, err := loop.Call(context.Background(), "analyze repo")
	require.NoError(t, err)
	require.Equal(t, []string{"fetch_data"}, toolDef.called)
	require.Len(t, eng.calls, 3)
	require.NotNil(t, eng.calls[2].Schema)
	require.Greater(t, len(eng.calls[2].Messages), 2)
	require.Contains(t, result.Content, "x@y.com")
}

func TestSpecializedLoop_MaxTokens_Default(t *testing.T) {
	eng := &fakeEngine{
		responses: []inference.Result{{Content: `{}`}},
	}
	loop := NewSpecializedLoop(SpecializedConfig{
		SystemPrompt: "sys",
		Engine:       eng,
		Schema:       core.Schema{Type: "object"},
	})

	_, err := loop.Call(context.Background(), "go")
	require.NoError(t, err)
	require.Equal(t, 2048, eng.calls[0].MaxTokens)
}
