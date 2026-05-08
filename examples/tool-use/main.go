package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"

	"github.com/miroslav-matejovsky/pagantic/inference"
	"github.com/miroslav-matejovsky/pagantic/kronk"
	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	"github.com/miroslav-matejovsky/pagantic/tool"
	"github.com/miroslav-matejovsky/pagantic/tui"
)

const llmModel = "unsloth/Qwen3-0.6B-Q8_0"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	registry := tool.NewRegistry(&diceTool{})

	repl := tui.NewAgentREPL(tui.AgentConfig{
		Title:  "tool-use",
		Banner: "Chat with an LLM that can roll dice. Try: chat, then ask it to roll 2d6.",
		SystemPrompt: `You are a dice assistant. When the user asks to roll dice or generate random results,
call the roll_dice tool. Do not invent dice results yourself.`,
		EngineLoader: func(ctx context.Context) (inference.Engine, func(), error) {
			krn, cleanup, err := kronk.Load(ctx, kronk.Config{ModelSource: llmModel})
			if err != nil {
				return nil, nil, err
			}
			return inference.NewKronkAdapter(krn, nil), cleanup, nil
		},
		Registry: registry,
	})

	repl.AddCommand(tui.Command{
		Name:        "info",
		Description: "Show example info",
		Run: func(_ context.Context, _ []string) error {
			tui.Infof("Local dir: %s", repl.LocalDir())
			tui.Warnf("Engine loads on first chat use")
			fmt.Println(tui.Dim("This is a tool-use demo"))
			return nil
		},
	})

	repl.AddCommand(tui.Command{
		Name:        "engine",
		Description: "Load engine and show model info",
		Run: func(ctx context.Context, _ []string) error {
			eng, err := repl.Engine(ctx)
			if err != nil {
				return err
			}
			info := eng.ModelInfo()
			tui.Infof("Context window: %d", info.ContextWindow)
			return nil
		},
	})

	repl.Run(ctx)
	fmt.Println("Done.")
}

// diceTool rolls random dice. Implements tool.Tool.
type diceTool struct{}

func (d *diceTool) Info() tool.ToolInfo {
	return tool.ToolInfo{
		Name:        "roll_dice",
		Type:        tool.TypeGo,
		Description: "Roll dice with configurable sides and count",
	}
}

func (d *diceTool) Definition() core.ToolDefinition {
	return core.ToolDefinition{
		Name:        "roll_dice",
		Description: "Roll one or more dice and return the results",
		Parameters: core.Schema{
			Type: "object",
			Properties: map[string]core.Schema{
				"sides": {
					Type:        "integer",
					Description: "Number of sides per die (default 6)",
				},
				"count": {
					Type:        "integer",
					Description: "Number of dice to roll (default 1)",
				},
			},
			Required: []string{},
		},
	}
}

func (d *diceTool) Execute(args map[string]any) (string, error) {
	sides := intArg(args, "sides", 6)
	count := intArg(args, "count", 1)

	if sides < 2 || sides > 1000 {
		return "Error: sides must be between 2 and 1000", nil
	}
	if count < 1 || count > 20 {
		return "Error: count must be between 1 and 20", nil
	}

	results := make([]string, count)
	total := 0
	for i := range results {
		v := rand.Intn(sides) + 1
		results[i] = fmt.Sprintf("%d", v)
		total += v
	}

	return fmt.Sprintf("Rolled %dd%d: [%s], total=%d", count, sides, strings.Join(results, " "), total), nil
}

func (d *diceTool) Available() (bool, string) {
	return true, ""
}

// intArg extracts an integer argument with a default value.
func intArg(args map[string]any, key string, def int) int {
	v, ok := args[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return def
	}
}
