// CLI example: direct tool execution with observability.
//
// Demonstrates layer 4 (tool) and layer 10 (observe) without LLM inference:
//
//   - Layer 4 (tool): ToolExecutor wraps Registry with event recording.
//     Execute() dispatches tool calls and records start/end events.
//     Registry.AllDefinitions() returns schemas for all registered tools
//     (vs Definitions() which returns only available ones).
//
//   - Layer 10 (observe): InMemoryEventLog stores events in memory.
//     ToolExecutor records events through the EventLog interface.
//     After execution, Events() returns all recorded events for inspection.
//
//   - No LLM needed: this example demonstrates the deterministic tool layer
//     in isolation. In production, AgentLoop calls Registry.Execute directly.
//     ToolExecutor adds observability on top.
//
// When to use ToolExecutor vs Registry.Execute:
//   - ToolExecutor: when you need event recording, audit trail, or debugging
//   - Registry.Execute: when you just need tool dispatch (AgentLoop path)
//
// Key pagantic concept: tools are fully deterministic. Same input always
// produces same output. The executor adds observability without changing
// behavior.
package main

import (
	"context"
	"fmt"
	"os"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	tool "github.com/miroslav-matejovsky/pagantic/layers/04_tool"
	observe "github.com/miroslav-matejovsky/pagantic/layers/10_observe"
)

func main() {
	registry := tool.NewRegistry(&calcTool{}, &greetTool{})

	// AllDefinitions returns schemas for ALL tools (including unavailable).
	allDefs := registry.AllDefinitions()
	fmt.Fprintf(os.Stderr, "Registered tools (%d):\n", len(allDefs))
	for _, def := range allDefs {
		fmt.Fprintf(os.Stderr, "  - %s: %s\n", def.Name, def.Description)
	}
	fmt.Fprintln(os.Stderr)

	// ToolExecutor wraps registry with event recording.
	eventLog := &observe.InMemoryEventLog{}
	executor := &tool.ToolExecutor{
		Registry: registry,
		Observer: eventLog,
	}

	// Execute tool calls and collect results.
	calls := []core.ToolCall{
		{ID: "call-1", Name: "calculate", Arguments: map[string]any{"operation": "add", "a": float64(15), "b": float64(27)}},
		{ID: "call-2", Name: "greet", Arguments: map[string]any{"name": "World"}},
		{ID: "call-3", Name: "calculate", Arguments: map[string]any{"operation": "multiply", "a": float64(6), "b": float64(7)}},
		{ID: "call-4", Name: "unknown_tool", Arguments: map[string]any{}},
	}

	for _, call := range calls {
		result := executor.Execute(context.Background(), call)
		if result.IsError {
			fmt.Fprintf(os.Stderr, "[ERROR] %s: %s\n", call.Name, result.Content)
		} else {
			fmt.Printf("%s(%v) = %s\n", call.Name, call.Arguments, result.Content)
		}
	}

	// Print recorded events from observer.
	events := eventLog.Events()
	fmt.Fprintf(os.Stderr, "\nRecorded events (%d):\n", len(events))
	for _, ev := range events {
		errStr := ""
		if ev.Error != nil {
			errStr = fmt.Sprintf(" error=%v", ev.Error)
		}
		fmt.Fprintf(os.Stderr, "  [%s] %s.%s duration=%v%s\n",
			ev.Data["call_id"], ev.Layer, ev.Action, ev.Duration, errStr)
	}
}

// calcTool does basic arithmetic. Implements tool.Tool.
type calcTool struct{}

func (c *calcTool) Info() tool.ToolInfo {
	return tool.ToolInfo{
		Name:        "calculate",
		Type:        tool.TypeGo,
		Description: "Perform basic arithmetic operations",
	}
}

func (c *calcTool) Definition() core.ToolDefinition {
	return core.ToolDefinition{
		Name:        "calculate",
		Description: "Perform basic arithmetic (add, subtract, multiply, divide)",
		Parameters: core.Schema{
			Type: "object",
			Properties: map[string]core.Schema{
				"operation": {
					Type:        "string",
					Description: "Operation: add, subtract, multiply, divide",
					Enum:        []string{"add", "subtract", "multiply", "divide"},
				},
				"a": {Type: "number", Description: "First operand"},
				"b": {Type: "number", Description: "Second operand"},
			},
			Required: []string{"operation", "a", "b"},
		},
	}
}

func (c *calcTool) Execute(args map[string]any) (string, error) {
	op, _ := args["operation"].(string)
	a, _ := args["a"].(float64)
	b, _ := args["b"].(float64)

	switch op {
	case "add":
		return fmt.Sprintf("%.0f", a+b), nil
	case "subtract":
		return fmt.Sprintf("%.0f", a-b), nil
	case "multiply":
		return fmt.Sprintf("%.0f", a*b), nil
	case "divide":
		if b == 0 {
			return "", fmt.Errorf("division by zero")
		}
		return fmt.Sprintf("%.2f", a/b), nil
	default:
		return "", fmt.Errorf("unknown operation: %s", op)
	}
}

func (c *calcTool) Available() (bool, string) { return true, "" }

// greetTool returns a greeting. Implements tool.Tool.
type greetTool struct{}

func (g *greetTool) Info() tool.ToolInfo {
	return tool.ToolInfo{
		Name:        "greet",
		Type:        tool.TypeGo,
		Description: "Return a greeting message",
	}
}

func (g *greetTool) Definition() core.ToolDefinition {
	return core.ToolDefinition{
		Name:        "greet",
		Description: "Generate a greeting for the given name",
		Parameters: core.Schema{
			Type: "object",
			Properties: map[string]core.Schema{
				"name": {Type: "string", Description: "Name to greet"},
			},
			Required: []string{"name"},
		},
	}
}

func (g *greetTool) Execute(args map[string]any) (string, error) {
	name, _ := args["name"].(string)
	if name == "" {
		name = "stranger"
	}
	return fmt.Sprintf("Hello, %s!", name), nil
}

func (g *greetTool) Available() (bool, string) { return true, "" }

// Usage:
//
//	go run examples/cli/tool-query/main.go
