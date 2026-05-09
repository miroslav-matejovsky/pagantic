// TUI example: interactive multi-turn chat without tools or context.
//
// Demonstrates pagantic as a Probabilistic Agentic Control System:
//
//   - Layer 2 (orchestrate): AgentLoop manages multi-turn conversation
//     state. Each user message enters the control loop, which runs inference
//     and checks for tool calls. With no tools registered, the loop always
//     completes in one iteration per turn.
//
//   - Layer 9 (memory): ConversationBuffer accumulates message history
//     across turns. The orchestrate layer owns this state - the adapter
//     never touches it directly.
//
//   - Layer 1 (inference): Engine loaded lazily via EngineLoader on first
//     chat message. This keeps startup fast and avoids loading the model
//     if the user never enters chat mode.
//
//   - Adapter (tui): AgentREPL provides command dispatch (help, tools, chat,
//     quit). The chat command enters an inner loop that reads user input
//     and delegates to AgentLoop. Streaming tokens rendered to terminal.
//
// Key pagantic concept: the adapter is a thin shell around the orchestrate
// layer. All conversation logic, tool handling, and state management live
// in the harness layers, not in the UI code.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/miroslav-matejovsky/pagantic/adapters/tui"
	"github.com/miroslav-matejovsky/pagantic/kronk"
	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	tool "github.com/miroslav-matejovsky/pagantic/layers/04_tool"
)

const llmModel = "unsloth/gemma-4-E4B-it"

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
