package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/miroslav-matejovsky/pagantic/agent"
	"github.com/miroslav-matejovsky/pagantic/kronk"
	"github.com/miroslav-matejovsky/pagantic/llm"
	"github.com/miroslav-matejovsky/pagantic/tui"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	registry := agent.NewRegistry()

	repl := tui.NewAgentREPL(tui.AgentConfig{
		Title:        "simple-chat",
		Banner:       "Type 'chat' to start chatting, 'quit' to exit.",
		SystemPrompt: "You are a helpful assistant. Be concise.",
		EngineLoader: func(ctx context.Context) (llm.Chat, func(), error) {
			return kronk.Load(ctx, kronk.Config{})
		},
		Registry: registry,
	})

	repl.Run(ctx)
	fmt.Println("Done.")
}
