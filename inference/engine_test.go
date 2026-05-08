package inference

import (
	"testing"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	"github.com/stretchr/testify/require"
)

func TestRequestAndResultConstruction(t *testing.T) {
	temp := 0.4
	schema := &core.Schema{Type: "object"}

	req := Request{
		Messages:    []core.Message{core.NewUserMessage("hello")},
		Tools:       []core.ToolDefinition{{Name: "search"}},
		Schema:      schema,
		MaxTokens:   32,
		Temperature: &temp,
		Options:     map[string]any{"top_p": 0.9},
	}
	result := Result{
		Content:   "done",
		ToolCalls: []core.ToolCall{{ID: "call-1", Name: "search"}},
		Messages:  []core.Message{core.NewAssistantMessage("done")},
		Usage:     core.TokenUsage{OutputTokens: 5},
	}

	require.Len(t, req.Messages, 1)
	require.Equal(t, 32, req.MaxTokens)
	require.Same(t, schema, req.Schema)
	require.Equal(t, temp, *req.Temperature)
	require.Equal(t, 0.9, req.Options["top_p"])
	require.Equal(t, "done", result.Content)
	require.Len(t, result.ToolCalls, 1)
	require.Equal(t, 5, result.Usage.OutputTokens)
}
