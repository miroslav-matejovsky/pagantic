package tool_test

import (
	"context"
	"errors"
	"testing"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	"github.com/miroslav-matejovsky/pagantic/observe"
	"github.com/miroslav-matejovsky/pagantic/tool"
	"github.com/stretchr/testify/require"
)

func TestToolExecutor_ExecuteSuccess(t *testing.T) {
	log := &observe.InMemoryEventLog{}
	executor := tool.ToolExecutor{
		Registry: tool.NewRegistry(&fakeTool{name: "echo", toolType: tool.TypeGo, available: true, output: "done"}),
		Observer: log,
	}

	result := executor.Execute(context.Background(), core.ToolCall{
		ID:   "call-1",
		Name: "echo",
		Arguments: map[string]any{
			"msg": "hi",
		},
	})

	require.Equal(t, core.ToolResult{
		CallID:  "call-1",
		Name:    "echo",
		Content: "done",
		IsError: false,
	}, result)

	events := log.Events()
	require.Len(t, events, 2)
	require.Equal(t, "tool", events[0].Layer)
	require.Equal(t, "execute_start", events[0].Action)
	require.Equal(t, "echo", events[0].Data["name"])
	require.Equal(t, map[string]any{"msg": "hi"}, events[0].Data["args"])
	_, ok := events[0].Data["call_id"]
	require.True(t, ok)

	require.Equal(t, "tool", events[1].Layer)
	require.Equal(t, "execute_end", events[1].Action)
	require.Equal(t, "echo", events[1].Data["name"])
	require.NoError(t, events[1].Error)
	require.GreaterOrEqual(t, events[1].Duration, int64(0))
}

func TestToolExecutor_ExecuteFailure(t *testing.T) {
	boom := errors.New("boom")
	log := &observe.InMemoryEventLog{}
	executor := tool.ToolExecutor{
		Registry: tool.NewRegistry(&fakeTool{name: "fail", toolType: tool.TypeGo, available: true, err: boom}),
		Observer: log,
	}

	result := executor.Execute(context.Background(), core.ToolCall{
		ID:   "call-2",
		Name: "fail",
	})

	require.Equal(t, core.ToolResult{
		CallID:  "call-2",
		Name:    "fail",
		Content: "boom",
		IsError: true,
	}, result)

	events := log.Events()
	require.Len(t, events, 2)
	require.NoError(t, events[0].Error)
	require.ErrorIs(t, events[1].Error, boom)
}

func TestToolExecutor_CancelledContextReturnsError(t *testing.T) {
	log := &observe.InMemoryEventLog{}
	executor := tool.ToolExecutor{
		Registry: tool.NewRegistry(&fakeTool{name: "echo", toolType: tool.TypeGo, available: true, output: "done"}),
		Observer: log,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	result := executor.Execute(ctx, core.ToolCall{
		ID:   "call-cancelled",
		Name: "echo",
	})

	require.True(t, result.IsError)
	require.Contains(t, result.Content, "context canceled")

	events := log.Events()
	require.Len(t, events, 2)
	require.Equal(t, "execute_start", events[0].Action)
	require.Equal(t, "execute_cancelled", events[1].Action)
}

func TestToolExecutor_NilObserver(t *testing.T) {
	executor := tool.ToolExecutor{
		Registry: tool.NewRegistry(&fakeTool{name: "echo", toolType: tool.TypeGo, available: true, output: "ok"}),
	}

	require.NotPanics(t, func() {
		result := executor.Execute(context.Background(), core.ToolCall{ID: "call-3", Name: "echo"})
		require.Equal(t, "ok", result.Content)
		require.False(t, result.IsError)
	})
}
