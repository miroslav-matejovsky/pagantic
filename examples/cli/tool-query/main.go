// CLI example: agent loop with tools and streaming.
//
// Demonstrates model-driven tool use through layers 2, 4, and the cli adapter:
//
//   - Layer 4 (tool): Registry wraps calcTool and greetTool. Definitions()
//     returns schemas the model receives. Execute dispatches calls the model
//     makes.
//
//   - Layer 2 (orchestrate): AgentLoop runs the inference-tool loop. The
//     model decides when to call tools; AgentLoop dispatches them and feeds
//     results back until the model produces a final text answer.
//
//   - CLI adapter: Runner wires model + registry. When Stream is nil,
//     Runner creates a default handler that prints model info to stderr and
//     streams response tokens to stdout as they arrive.
//
// Contrast with plain Registry.Execute (no model):
//
//   - Registry.Execute: deterministic dispatch, caller supplies the tool call.
//   - AgentLoop + Registry: the model decides which tools to call and when,
//     based on the user prompt.
//
// Key pagantic concept: tools are fully deterministic. The model layer is
// probabilistic only in deciding which tools to invoke; execution is always
// deterministic. Same tool arguments always produce the same result.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/miroslav-matejovsky/pagantic/adapters/cli"
	"github.com/miroslav-matejovsky/pagantic/kronk"
	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	tool "github.com/miroslav-matejovsky/pagantic/layers/04_tool"
)

const llmModel = "unsloth/gemma-4-E4B-it"

// defaultQuery is used when no arguments are passed. It exercises both tools.
const defaultQuery = "What is 15 plus 27? Also greet Alice."

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	prompt, err := cli.ReadPrompt(os.Args[1:], os.Stdin)
	if err != nil {
		if !errors.Is(err, cli.ErrNoPrompt) {
			fmt.Fprintf(os.Stderr, "Error reading prompt: %v\n", err)
			os.Exit(1)
		}
		prompt = defaultQuery
	}

	krn, cleanup, err := kronk.Load(ctx, kronk.Config{ModelSource: llmModel})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load engine: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	engine := kronk.NewAdapter(krn, nil)

	registry := tool.NewRegistry(&calcTool{}, &greetTool{})

	fmt.Fprintf(os.Stderr, "Registered tools (%d):\n", len(registry.AllDefinitions()))
	for _, def := range registry.AllDefinitions() {
		fmt.Fprintf(os.Stderr, "  - %s: %s\n", def.Name, def.Description)
	}
	fmt.Fprintln(os.Stderr)

	runner, err := cli.NewRunner(cli.RunConfig{
		Engine:            engine,
		Registry:          registry,
		SystemPrompt:      "You are a helpful assistant with access to tools. Use them when needed.",
		Out:               os.Stdout,
		MaxTokens:         2048,
		MaxToolIterations: 20,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := runner.Run(ctx, prompt); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
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
	op, ok := args["operation"].(string)
	if !ok || op == "" {
		return "", fmt.Errorf("missing or invalid argument: operation (string required)")
	}
	a, ok := args["a"].(float64)
	if !ok {
		return "", fmt.Errorf("missing or invalid argument: a (number required)")
	}
	b, ok := args["b"].(float64)
	if !ok {
		return "", fmt.Errorf("missing or invalid argument: b (number required)")
	}

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
//	go run examples/cli/tool-query/main.go "What is 100 divided by 4? Greet Bob too."
