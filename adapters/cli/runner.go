package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	orchestrate "github.com/miroslav-matejovsky/pagantic/layers/02_orchestrate"
	tool "github.com/miroslav-matejovsky/pagantic/layers/04_tool"
	"golang.org/x/term"
)

// DefaultTimeout is applied when RunConfig.Timeout is zero.
// The kronk inference engine requires a context with a deadline.
const DefaultTimeout = 120 * time.Second

// RunConfig configures a single-shot CLI execution.
type RunConfig struct {
	// Engine is the inference engine to use. Required.
	Engine inference.Engine
	// SystemPrompt defines the system message for the conversation.
	SystemPrompt string
	// Registry provides tools for the agent loop. May be nil.
	Registry *tool.Registry
	// ContextProvider retrieves context before each inference call. May be nil.
	ContextProvider orchestrate.ContextProvider
	// Stream receives streaming tokens during inference. May be nil.
	Stream *inference.StreamHandler
	// Timeout limits total execution time. Zero uses DefaultTimeout (120s).
	Timeout time.Duration
	// Out is the output writer. Defaults to os.Stdout when nil.
	Out io.Writer
}

// Runner executes a single inference request and writes the result.
type Runner struct {
	cfg RunConfig
}

// NewRunner creates a Runner from config. Panics if Engine is nil.
// Out defaults to os.Stdout when nil.
func NewRunner(cfg RunConfig) *Runner {
	if cfg.Engine == nil {
		panic("cli: RunConfig.Engine must not be nil")
	}
	if cfg.Out == nil {
		cfg.Out = os.Stdout
	}
	return &Runner{cfg: cfg}
}

// Run executes a single inference call with the given prompt and writes
// the response content to the configured output writer.
// Always applies a timeout: uses cfg.Timeout if set, otherwise DefaultTimeout.
// When Stream is nil, Run prints model info to stderr and streams tokens to Out
// as they arrive instead of buffering the full response.
func (r *Runner) Run(ctx context.Context, prompt string) error {
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("cli: empty prompt")
	}

	timeout := r.cfg.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	// Use the caller-provided stream handler, or build a default one that
	// prints model info and streams tokens to Out as they arrive.
	stream := r.cfg.Stream
	var streamed bool
	if stream == nil {
		info := r.cfg.Engine.ModelInfo()
		fmt.Fprintf(os.Stderr, "model: %s  context: %d tokens\n\n", info.Name, info.ContextWindow)
		out := r.cfg.Out
		stream = &inference.StreamHandler{
			OnContent: func(text string) {
				streamed = true
				_, _ = fmt.Fprint(out, text)
			},
			OnToolCall: func(name, argsJSON string) {
				fmt.Fprintf(os.Stderr, "[tool] %s %s\n", name, argsJSON)
			},
		}
	}

	agent := orchestrate.NewAgentLoop(orchestrate.LoopConfig{
		SystemPrompt:    r.cfg.SystemPrompt,
		Engine:          r.cfg.Engine,
		Tools:           r.cfg.Registry,
		Stream:          stream,
		ContextProvider: r.cfg.ContextProvider,
	})

	result, err := agent.Chat(ctx, prompt)
	if err != nil {
		return fmt.Errorf("cli: %w", err)
	}

	if r.cfg.Stream == nil {
		if streamed {
			// Default stream printed tokens incrementally; add a trailing newline.
			_, _ = fmt.Fprintln(r.cfg.Out)
		} else {
			// Engine did not emit streaming tokens (e.g. test stub); print full content.
			_, err = fmt.Fprintln(r.cfg.Out, result.Content)
			if err != nil {
				return fmt.Errorf("cli: write output: %w", err)
			}
		}
	}
	// When r.cfg.Stream != nil the caller's handler dealt with output; write nothing.

	return nil
}

// ReadPrompt builds a prompt from command-line arguments. If args is empty,
// reads all of stdin. Returns an error immediately when stdin is an interactive
// terminal (no pipe), to avoid blocking the caller.
func ReadPrompt(args []string, stdin io.Reader) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}

	if f, ok := stdin.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		return "", fmt.Errorf("cli: no prompt provided (pass as arguments or pipe to stdin)")
	}

	data, err := io.ReadAll(stdin)
	if err != nil {
		return "", fmt.Errorf("cli: read stdin: %w", err)
	}

	prompt := strings.TrimSpace(string(data))
	if prompt == "" {
		return "", fmt.Errorf("cli: no prompt provided (pass as arguments or pipe to stdin)")
	}

	return prompt, nil
}
