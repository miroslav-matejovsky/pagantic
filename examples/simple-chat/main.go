package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/miroslav-matejovsky/pagantic/kronk"
	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	tool "github.com/miroslav-matejovsky/pagantic/layers/04_tool"
	"github.com/miroslav-matejovsky/pagantic/tui"
)

const llmModel = "unsloth/Qwen3-0.6B-Q8_0"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	registry := tool.NewRegistry()

	repl := tui.NewAgentREPL(tui.AgentConfig{
		Title:        "simple-chat",
		Banner:       "Type 'chat' to start chatting, 'quit' to exit.",
		SystemPrompt: "You are a helpful assistant. Be concise.",
		EngineLoader: func(ctx context.Context) (inference.Engine, func(), error) {
			krn, cleanup, err := kronk.Load(ctx, kronk.Config{ModelSource: llmModel})
			if err != nil {
				return nil, nil, err
			}
			return kronk.NewAdapter(krn, nil), cleanup, nil
		},
		Registry: registry,
	})

	repl.Run(ctx)
	fmt.Println("Done.")
}
