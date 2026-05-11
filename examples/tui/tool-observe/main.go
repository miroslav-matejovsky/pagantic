// TUI example: tool observability with ToolExecutor.
//
// Demonstrates two distinct tool dispatch paths side-by-side:
//
//   - ToolExecutor.Execute: direct tool dispatch with per-call observability.
//     ToolExecutor wraps Registry with InMemoryEventLog, recording
//     execute_start and execute_end events with timing, args, and errors.
//     Use the 'exec' command to call tools directly and 'audit' to inspect
//     the event log.
//
//   - Registry.Execute (via AgentLoop): model-driven dispatch. The model
//     decides which tools to call. AgentLoop calls Registry.Execute directly,
//     without per-call event recording. Use 'chat' to drive this path.
//
// When to choose which:
//   - ToolExecutor: audit trail, debugging, timing, or manual tool dispatch.
//   - Registry.Execute via AgentLoop: standard model-driven tool use where
//     observability is at the loop level (LoopConfig.Observer), not per-call.
//
// Key pagantic concept: ToolExecutor adds observability without changing
// behavior. Same registry, same tool execution, but every call is now
// visible in the event log.
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/miroslav-matejovsky/pagantic/adapters/tui"
	"github.com/miroslav-matejovsky/pagantic/kronk"
	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	orchestrate "github.com/miroslav-matejovsky/pagantic/layers/02_orchestrate"
	tool "github.com/miroslav-matejovsky/pagantic/layers/04_tool"
	observe "github.com/miroslav-matejovsky/pagantic/layers/10_observe"
)

const llmModel = "unsloth/gemma-4-E4B-it"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	registry := tool.NewRegistry(&calcTool{}, &greetTool{})
	eventLog := &observe.InMemoryEventLog{}
	executor := &tool.ToolExecutor{
		Registry: registry,
		Observer: eventLog,
	}

	var engine inference.Engine
	var engineCleanup func()

	loadEngine := func(ctx context.Context) error {
		if engine != nil {
			return nil
		}
		tui.Warnf("Loading inference engine...")
		krn, cleanup, err := kronk.Load(ctx, kronk.Config{ModelSource: llmModel})
		if err != nil {
			return fmt.Errorf("load engine: %w", err)
		}
		engine = kronk.NewAdapter(krn, nil)
		engineCleanup = cleanup
		return nil
	}

	repl := tui.NewREPL(tui.Bold("tool-observe>") + " ")
	repl.SetBanner("Tool observability demo. Commands: tools, exec, audit, chat, quit.")

	repl.AddCommand(tui.Command{
		Name:        "tools",
		Description: "List registered tools",
		Run: func(_ context.Context, _ []string) error {
			for _, s := range registry.CheckAvailability() {
				status := tui.Green("[OK]")
				if !s.Available {
					status = tui.Red("[--]")
				}
				_, _ = fmt.Fprintf(repl.Out, "  %s %-15s - %s\n", status, s.Name, s.Description)
			}
			return nil
		},
	})

	repl.AddCommand(tui.Command{
		Name:        "exec",
		Description: "Run tool directly via ToolExecutor (records events). Usage: exec calculate add 10 20 | exec greet Alice",
		Args:        "<tool> [args...]",
		Run: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: exec calculate <op> <a> <b>  |  exec greet <name>")
			}
			call, err := parseCall(args)
			if err != nil {
				return err
			}
			result := executor.Execute(ctx, call)
			if result.IsError {
				tui.FError(repl.Out, result.Content)
			} else {
				_, _ = fmt.Fprintln(repl.Out, tui.Green("=> ")+result.Content)
			}
			return nil
		},
	})

	repl.AddCommand(tui.Command{
		Name:        "audit",
		Description: "Print recorded events from ToolExecutor",
		Run: func(_ context.Context, _ []string) error {
			events := eventLog.Events()
			if len(events) == 0 {
				tui.FInfo(repl.Out, "No events yet. Run 'exec' first.")
				return nil
			}
			_, _ = fmt.Fprintf(repl.Out, tui.Cyan("Events (%d):\n"), len(events))
			for _, ev := range events {
				errStr := ""
				if ev.Error != nil {
					errStr = tui.Red(" error=" + ev.Error.Error())
				}
				_, _ = fmt.Fprintf(repl.Out, "  [%s] %s.%s%s  duration=%v\n",
					tui.Dim(ev.Timestamp.Format("15:04:05.000")),
					ev.Layer, ev.Action, errStr, time.Duration(ev.Duration))
			}
			return nil
		},
	})

	repl.AddCommand(tui.Command{
		Name:        "chat",
		Description: "Chat with the model - it drives tool calls via Registry.Execute (no ToolExecutor events)",
		Run: func(ctx context.Context, _ []string) error {
			if err := loadEngine(ctx); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(repl.Out, "\n"+tui.Cyan("=== Chat Mode ==="))
			_, _ = fmt.Fprintln(repl.Out, "Model drives tool calls. Events will NOT appear in 'audit'.")
			_, _ = fmt.Fprintln(repl.Out, "Type 'exit' to return.")
			_, _ = fmt.Fprintln(repl.Out)

			chatAgent := orchestrate.NewAgentLoop(orchestrate.LoopConfig{
				SystemPrompt: "You are a helpful assistant with access to calculate and greet tools. Use them when needed.",
				Engine:       engine,
				Tools:        registry,
				Stream:       tui.TerminalRenderer(repl.Out),
				OnToolResult: func(name, output string) {
					_, _ = fmt.Fprintf(repl.Out, "\n%s\n%s\n", tui.Green("Tool: "+name), tui.SanitizeOutput(output))
				},
			})

			scanner := bufio.NewScanner(repl.In)
			for {
				line, err := tui.FPrompt(scanner, repl.Out, tui.Bold("chat>")+" ")
				if err != nil {
					if !errors.Is(err, io.EOF) {
						tui.FError(repl.ErrOut, err.Error())
					}
					break
				}
				if line == "exit" || line == "quit" {
					break
				}
				chatCtx, cancel := context.WithTimeout(ctx, 180*time.Second)
				result, chatErr := chatAgent.Chat(chatCtx, line)
				cancel()
				if chatErr != nil {
					tui.FErrorf(repl.ErrOut, "chat error: %v", chatErr)
					continue
				}
				_, _ = fmt.Fprintln(repl.Out)
				tui.FPrintUsage(repl.Out, toUsageStats(result.Usage))
			}
			tui.FInfo(repl.Out, "Back to main menu.")
			return nil
		},
	})

	repl.Run(ctx)
	if engineCleanup != nil {
		engineCleanup()
	}
	_, _ = fmt.Fprintln(repl.Out, tui.Grey("Bye."))
}

// parseCall builds a ToolCall from REPL command args.
//
//	exec calculate <op> <a> <b>
//	exec greet <name>
func parseCall(args []string) (core.ToolCall, error) {
	toolName := strings.ToLower(args[0])
	rest := args[1:]

	switch toolName {
	case "calculate":
		if len(rest) < 3 {
			return core.ToolCall{}, fmt.Errorf("usage: exec calculate <op> <a> <b>  (ops: add subtract multiply divide)")
		}
		a, err := strconv.ParseFloat(rest[1], 64)
		if err != nil {
			return core.ToolCall{}, fmt.Errorf("invalid number: %s", rest[1])
		}
		b, err := strconv.ParseFloat(rest[2], 64)
		if err != nil {
			return core.ToolCall{}, fmt.Errorf("invalid number: %s", rest[2])
		}
		return core.ToolCall{
			ID:   fmt.Sprintf("exec-%d", time.Now().UnixNano()),
			Name: "calculate",
			Arguments: map[string]any{
				"operation": rest[0],
				"a":         a,
				"b":         b,
			},
		}, nil

	case "greet":
		name := "stranger"
		if len(rest) > 0 {
			name = strings.Join(rest, " ")
		}
		return core.ToolCall{
			ID:   fmt.Sprintf("exec-%d", time.Now().UnixNano()),
			Name: "greet",
			Arguments: map[string]any{
				"name": name,
			},
		}, nil

	default:
		return core.ToolCall{}, fmt.Errorf("unknown tool: %s (available: calculate, greet)", toolName)
	}
}

func toUsageStats(u core.TokenUsage) tui.UsageStats {
	return tui.UsageStats{
		PromptTokens:    u.PromptTokens,
		ReasoningTokens: u.ReasoningTokens,
		OutputTokens:    u.OutputTokens,
		ContextTokens:   u.ContextTokens,
		ContextWindow:   u.ContextWindow,
		TokensPerSecond: u.TokensPerSecond,
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
		return fmt.Sprintf("%.6g", a+b), nil
	case "subtract":
		return fmt.Sprintf("%.6g", a-b), nil
	case "multiply":
		return fmt.Sprintf("%.6g", a*b), nil
	case "divide":
		if b == 0 {
			return "", fmt.Errorf("division by zero")
		}
		return fmt.Sprintf("%.6g", a/b), nil
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
//	go run examples/tui/tool-observe/main.go
//
// Then try:
//
//	tools                        - list registered tools
//	exec calculate add 10 20     - run directly via ToolExecutor
//	exec calculate divide 9 0    - demonstrates error recording
//	exec greet Alice             - direct tool call with observability
//	audit                        - inspect event log from ToolExecutor
//	chat                         - model-driven tool use (no audit events)
//	quit
