package tui

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/miroslav-matejovsky/pagantic/core"
	"github.com/miroslav-matejovsky/pagantic/inference"
	"github.com/miroslav-matejovsky/pagantic/orchestrate"
	"github.com/miroslav-matejovsky/pagantic/tool"
)

// AgentConfig controls AgentREPL creation.
type AgentConfig struct {
	// Title is printed as the header when Run starts.
	Title string
	// Banner is passed to the underlying REPL as a startup message.
	Banner string
	// SystemPrompt defines the built-in chat loop's role.
	SystemPrompt string
	// EngineLoader is called once on first use to load the inference engine.
	// It returns the engine, a cleanup func (may be nil), and an error.
	// Required.
	EngineLoader func(context.Context) (inference.Engine, func(), error)
	// Registry provides tools for the chat loop and tool-status listing.
	// Required.
	Registry *tool.Registry
	// LocalDir is the working directory for file I/O (cloned repos, etc.).
	LocalDir string
}

// AgentREPL is a generic agent-harness REPL. It provides:
//   - a tools command that lists tool availability
//   - a chat command backed by a streaming orchestration loop
//   - lazy inference engine loading via EngineLoader
//   - an extensible command set via AddCommand
//
// Application-specific commands are registered from outside via AddCommand,
// keeping this type engine- and domain-agnostic.
//
// Input and output use the underlying REPL's In/Out/ErrOut fields, which
// default to stdin/stdout/stderr and can be replaced for testing.
type AgentREPL struct {
	cfg          AgentConfig
	engine       inference.Engine
	engineLoaded bool
	cleanup      func()
	repl         *REPL
}

// NewAgentREPL creates an AgentREPL from the given config.
// Defaults LocalDir to ".local" if empty.
// Panics if EngineLoader or Registry is nil.
func NewAgentREPL(cfg AgentConfig) *AgentREPL {
	if cfg.EngineLoader == nil {
		panic("tui: AgentConfig.EngineLoader must not be nil")
	}
	if cfg.Registry == nil {
		panic("tui: AgentConfig.Registry must not be nil")
	}
	if cfg.LocalDir == "" {
		cfg.LocalDir = ".local"
	}

	t := &AgentREPL{cfg: cfg}

	prompt := Bold(cfg.Title+">") + " "
	t.repl = NewREPL(prompt)
	if cfg.Banner != "" {
		t.repl.SetBanner(cfg.Banner)
	}

	t.repl.AddCommand(Command{
		Name:        "tools",
		Description: "List tools with descriptions and availability",
		Run: func(_ context.Context, _ []string) error {
			t.printToolStatus()
			return nil
		},
	})

	t.repl.AddCommand(Command{
		Name:        "chat",
		Description: "Interactive inference chat",
		Run: func(ctx context.Context, _ []string) error {
			return t.runChat(ctx)
		},
	})

	return t
}

// AddCommand registers an additional command in the REPL.
func (t *AgentREPL) AddCommand(cmd Command) {
	t.repl.AddCommand(cmd)
}

// Engine returns the lazily-loaded inference engine, loading it on first call.
// Use this in custom commands registered via AddCommand.
func (t *AgentREPL) Engine(ctx context.Context) (inference.Engine, error) {
	if err := t.ensureEngine(ctx); err != nil {
		return nil, err
	}
	return t.engine, nil
}

// LocalDir returns the working directory configured for this REPL.
func (t *AgentREPL) LocalDir() string {
	return t.cfg.LocalDir
}

// Run prints the title header, lists tools, then starts the REPL loop.
// Blocks until the user types quit/exit, input is exhausted, or ctx is cancelled.
func (t *AgentREPL) Run(ctx context.Context) {
	_, _ = fmt.Fprintln(t.repl.Out, Bold("=== "+t.cfg.Title+" ==="))
	t.printToolStatus()
	_, _ = fmt.Fprintln(t.repl.Out)

	t.repl.Run(ctx)
	t.shutdown()
}

// ensureEngine loads the inference engine once via EngineLoader.
// A failed load does not set engineLoaded, allowing retry on the next call.
func (t *AgentREPL) ensureEngine(ctx context.Context) error {
	if t.engineLoaded {
		return nil
	}
	FWarn(t.repl.ErrOut, "Loading inference engine (first use)...")
	eng, cleanup, err := t.cfg.EngineLoader(ctx)
	if err != nil {
		return fmt.Errorf("engine load: %w", err)
	}
	t.engine = eng
	t.cleanup = cleanup
	t.engineLoaded = true
	return nil
}

func (t *AgentREPL) shutdown() {
	if t.cleanup != nil {
		t.cleanup()
	}
	_, _ = fmt.Fprintln(t.repl.Out, Grey("Bye."))
}

func (t *AgentREPL) printToolStatus() {
	statuses := t.cfg.Registry.CheckAvailability()
	if len(statuses) == 0 {
		return
	}
	_, _ = fmt.Fprintln(t.repl.Out, "\nTools:")
	for _, s := range statuses {
		if s.Available {
			_, _ = fmt.Fprintf(t.repl.Out, "  %s %-15s - %s\n", Green("[OK]"), s.Name, s.Description)
		} else {
			_, _ = fmt.Fprintf(t.repl.Out, "  %s %-15s - %s (%s)\n", Red("[--]"), s.Name, s.Description, s.Reason)
		}
	}
}

func (t *AgentREPL) runChat(ctx context.Context) error {
	if err := t.ensureEngine(ctx); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(t.repl.Out, "\n"+Cyan("=== Chat Mode ==="))
	_, _ = fmt.Fprintln(t.repl.Out, "Type messages to chat with model. Tools are available.")
	_, _ = fmt.Fprintln(t.repl.Out, "Type 'exit' or 'quit' to return to main menu.")
	_, _ = fmt.Fprintln(t.repl.Out)

	chatAgent := orchestrate.NewAgentLoop(orchestrate.LoopConfig{
		SystemPrompt: t.cfg.SystemPrompt,
		Engine:       t.engine,
		Tools:        t.cfg.Registry,
		Stream:       TerminalRenderer(t.repl.Out),
		OnToolResult: func(name, output string) {
			_, _ = fmt.Fprintf(t.repl.Out, "\n%s\n%s\n", Green("Tool: "+name), SanitizeOutput(output))
		},
	})

	scanner := bufio.NewScanner(t.repl.In)

	for {
		line, err := FPrompt(scanner, t.repl.Out, Bold("chat>")+" ")
		if err != nil {
			if !errors.Is(err, io.EOF) {
				FError(t.repl.ErrOut, err.Error())
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
			FErrorf(t.repl.ErrOut, "Chat error: %v", chatErr)
			continue
		}

		_, _ = fmt.Fprintln(t.repl.Out)
		FPrintUsage(t.repl.Out, usageStats(result.Usage))
	}

	FInfo(t.repl.Out, "Back to main menu.")
	return nil
}

// usageStats converts core.TokenUsage to UsageStats for display.
func usageStats(u core.TokenUsage) UsageStats {
	return UsageStats{
		PromptTokens:    u.PromptTokens,
		ReasoningTokens: u.ReasoningTokens,
		OutputTokens:    u.OutputTokens,
		ContextTokens:   u.ContextTokens,
		ContextWindow:   u.ContextWindow,
		TokensPerSecond: u.TokensPerSecond,
	}
}
