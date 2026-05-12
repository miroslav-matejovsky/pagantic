// TUI example: memory-aware chat shell for layer 9 primitives.
//
// Demonstrates pagantic as a Probabilistic Agentic Control System with explicit
// memory separated by lifetime and purpose.
//
//   - SessionState is durable key-value memory for data that should survive
//     across commands and chat turns. It is thread-safe, explicit, and small on
//     purpose. Use it for facts the user wants to keep around, such as names,
//     preferences, IDs, or flags. In this example the remember, recall, forget,
//     and keys commands manipulate that state directly so the user can inspect
//     persistent memory without hiding anything in prompt text.
//
//   - WorkingMemory is transient per-step state. It exists only for current work
//     item, not for long-term conversation history. The example stores turn
//     number before inference and last_response after inference, then resets the
//     structure between turns. This makes lifetime boundary visible: working
//     memory is scratch space, not durable storage.
//
//   - ConversationBuffer is ordered message history for multi-turn interaction.
//     AgentLoop owns its own internal buffer for actual chat context. This
//     example also keeps a separate tracker buffer so the adapter can measure and
//     display message count with ConversationBuffer.Len without reaching inside
//     orchestrate internals. Each user turn appends one user message and one
//     assistant message to that tracker.
//
// Memory layer idea is simple: state should be explicit, inspectable, and easy
// to reset. Persistent facts, transient scratch data, and conversation history
// have different semantics, so they should live in different containers instead
// of being mixed together in one prompt string.
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

	"github.com/miroslav-matejovsky/pagantic/adapters/tui"
	"github.com/miroslav-matejovsky/pagantic/kronk"
	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	orchestrate "github.com/miroslav-matejovsky/pagantic/layers/02_orchestrate"
	tool "github.com/miroslav-matejovsky/pagantic/layers/04_tool"
	memory "github.com/miroslav-matejovsky/pagantic/layers/09_memory"
)

const llmModel = "unsloth/gemma-4-E4B-it"

const systemPrompt = "You are a helpful assistant. Be concise."

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	session := memory.NewSessionState()
	working := memory.NewWorkingMemory()
	tracker := memory.NewConversationBuffer(0)
	registry := tool.NewRegistry()

	repl := tui.NewAgentREPL(tui.AgentConfig{
		Title:        "memory-chat",
		Banner:       "Try: remember name Mira, recall name, keys, status, mchat",
		SystemPrompt: systemPrompt,
		EngineLoader: func(ctx context.Context) (inference.Engine, func(), error) {
			krn, cleanup, err := kronk.Load(ctx, kronk.Config{ModelSource: llmModel})
			if err != nil {
				return nil, nil, err
			}
			return kronk.NewAdapter(krn, nil), cleanup, nil
		},
		Registry: registry,
	})

	repl.AddCommand(tui.Command{
		Name:        "remember",
		Aliases:     []string{"rem"},
		Description: "Store value in session state",
		Args:        "<key> <value...>",
		Run: func(_ context.Context, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: remember <key> <value...>")
			}

			session.Set(args[0], strings.Join(args[1:], " "))
			tui.Infof("Remembered %q", args[0])
			return nil
		},
	})

	repl.AddCommand(tui.Command{
		Name:        "recall",
		Description: "Load value from session state",
		Args:        "<key>",
		Run: func(_ context.Context, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("usage: recall <key>")
			}

			value, ok := session.Get(args[0])
			if !ok {
				tui.Warnf("%q not found", args[0])
				return nil
			}

			tui.Infof("%s = %v", args[0], value)
			return nil
		},
	})

	repl.AddCommand(tui.Command{
		Name:        "forget",
		Description: "Delete value from session state",
		Args:        "<key>",
		Run: func(_ context.Context, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("usage: forget <key>")
			}

			session.Delete(args[0])
			tui.Infof("Forgot %q", args[0])
			return nil
		},
	})

	repl.AddCommand(tui.Command{
		Name:        "keys",
		Description: "List session state keys",
		Run: func(_ context.Context, _ []string) error {
			keys := session.Keys()
			if len(keys) == 0 {
				tui.Warnf("Session state empty")
				return nil
			}

			tui.Infof("Session keys: %s", strings.Join(keys, ", "))
			return nil
		},
	})

	repl.AddCommand(tui.Command{
		Name:        "status",
		Description: "Show memory layer status",
		Run: func(_ context.Context, _ []string) error {
			keys := session.Keys()
			tui.Infof("Conversation messages: %d", tracker.Len())
			tui.Infof("Session key count: %d", len(keys))

			resultKeys := make([]string, 0, len(working.StepResults))
			for key := range working.StepResults {
				resultKeys = append(resultKeys, key)
			}
			sort.Strings(resultKeys)

			if len(resultKeys) == 0 {
				tui.Warnf("Working memory empty")
			} else {
				tui.Infof("Working memory:")
				for _, key := range resultKeys {
					value, ok := working.GetResult(key)
					if !ok {
						continue
					}
					fmt.Printf("  %s = %v\n", key, value)
				}
			}

			if len(keys) == 0 {
				fmt.Println(tui.Dim("Session keys: empty"))
			} else {
				fmt.Println(tui.Dim("Session keys: " + strings.Join(keys, ", ")))
			}
			return nil
		},
	})

	repl.AddCommand(tui.Command{
		Name:        "mchat",
		Description: "Interactive chat with memory tracking",
		Run: func(ctx context.Context, _ []string) error {
			eng, err := repl.Engine(ctx)
			if err != nil {
				return err
			}

			working.Reset()
			tracker = memory.NewConversationBuffer(0)
			turnNum := 0

			chatAgent, err := orchestrate.NewAgentLoop(orchestrate.LoopConfig{
				Engine:            eng,
				Tools:             registry,
				SystemPrompt:      systemPrompt,
				Stream:            tui.TerminalRenderer(os.Stdout),
				MaxTokens:         2048,
				MaxToolIterations: 20,
				OnToolResult: func(name, output string) {
					_, _ = fmt.Fprintf(os.Stdout, "\n%s\n%s\n", tui.Green("Tool: "+name), tui.SanitizeOutput(output))
				},
			})
			if err != nil {
				return err
			}

			fmt.Println(tui.Cyan("\n=== Memory Chat Mode ==="))
			fmt.Println("SessionState stays across commands.")
			fmt.Println("WorkingMemory resets between turns.")
			fmt.Println("ConversationBuffer.Len tracks message count.")
			fmt.Println("Type 'exit' or 'quit' to return.")
			fmt.Println()

			scanner := bufio.NewScanner(os.Stdin)
			scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
			for {
				line, err := tui.FPrompt(scanner, os.Stdout, tui.Bold("mchat>")+" ")
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return err
				}
				if line == "exit" || line == "quit" {
					break
				}

				turnNum++
				working.SetResult("turn", turnNum)

				chatCtx, cancel := context.WithTimeout(ctx, 180*time.Second)
				result, err := chatAgent.Chat(chatCtx, line)
				cancel()
				if err != nil {
					return err
				}

				lastResponse := result.Content
				if len(lastResponse) > 80 {
					lastResponse = lastResponse[:80] + "..."
				}
				working.SetResult("last_response", lastResponse)
				tracker.Append(core.NewUserMessage(line))
				tracker.Append(core.NewAssistantMessage(result.Content))

				fmt.Println()
				tui.FPrintUsage(os.Stdout, tui.UsageStats{
					PromptTokens:    result.Usage.PromptTokens,
					ReasoningTokens: result.Usage.ReasoningTokens,
					OutputTokens:    result.Usage.OutputTokens,
					ContextTokens:   result.Usage.ContextTokens,
					ContextWindow:   result.Usage.ContextWindow,
					TokensPerSecond: result.Usage.TokensPerSecond,
				})
				tui.Infof("Conversation messages: %d", tracker.Len())

				if value, ok := working.GetResult("turn"); ok {
					fmt.Println(tui.Dim(fmt.Sprintf("working.turn = %v", value)))
				}
				if value, ok := working.GetResult("last_response"); ok {
					fmt.Println(tui.Dim(fmt.Sprintf("working.last_response = %v", value)))
				}

				working.Reset()
			}

			tui.Infof("Back to main menu.")
			return nil
		},
	})

	repl.Run(ctx)
}
