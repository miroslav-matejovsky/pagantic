package tool

import (
	"context"
	"errors"
	"time"

	"github.com/miroslav-matejovsky/pagantic/core"
	"github.com/miroslav-matejovsky/pagantic/observe"
)

// ToolExecutor wraps registry with observability.
type ToolExecutor struct {
	Registry *Registry
	Observer observe.EventLog
}

// Execute runs tool call and records events.
func (te *ToolExecutor) Execute(ctx context.Context, call core.ToolCall) core.ToolResult {
	start := time.Now()
	te.record(observe.Event{
		Timestamp: start,
		Layer:     "tool",
		Action:    "execute_start",
		Data: map[string]any{
			"call_id": call.ID,
			"name":    call.Name,
			"args":    call.Arguments,
		},
	})

	result := core.ToolResult{
		CallID: call.ID,
		Name:   call.Name,
	}

	if err := ctx.Err(); err != nil {
		result.Content = err.Error()
		result.IsError = true
		te.record(observe.Event{
			Timestamp: time.Now(),
			Layer:     "tool",
			Action:    "execute_cancelled",
			Data: map[string]any{
				"call_id": call.ID,
				"name":    call.Name,
			},
			Duration: time.Since(start),
			Error:    err,
		})
		return result
	}

	if te == nil || te.Registry == nil {
		err := errors.New("tool executor has no registry")
		result.Content = err.Error()
		result.IsError = true
		te.record(observe.Event{
			Timestamp: time.Now(),
			Layer:     "tool",
			Action:    "execute_end",
			Data: map[string]any{
				"call_id": call.ID,
				"name":    call.Name,
				"args":    call.Arguments,
			},
			Duration: time.Since(start),
			Error:    err,
		})
		return result
	}

	content, err := te.Registry.Execute(call.Name, call.Arguments)
	result.Content = content
	if err != nil {
		result.Content = err.Error()
		result.IsError = true
	}

	te.record(observe.Event{
		Timestamp: time.Now(),
		Layer:     "tool",
		Action:    "execute_end",
		Data: map[string]any{
			"call_id": call.ID,
			"name":    call.Name,
			"args":    call.Arguments,
		},
		Duration: time.Since(start),
		Error:    err,
	})

	return result
}

func (te *ToolExecutor) record(event observe.Event) {
	if te == nil || te.Observer == nil {
		return
	}
	te.Observer.Record(event)
}
