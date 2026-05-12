package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	orchestrate "github.com/miroslav-matejovsky/pagantic/layers/02_orchestrate"
	tool "github.com/miroslav-matejovsky/pagantic/layers/04_tool"
	"golang.org/x/term"
)

// DefaultTimeout is applied when RunConfig.Timeout is zero.
// The kronk inference engine requires a context with a deadline.
const DefaultTimeout = 120 * time.Second

// ErrNoPrompt is returned by ReadPrompt when no prompt is provided.
// Callers may check for this error with errors.Is to distinguish "nothing
// provided" from real I/O failures.
var ErrNoPrompt = errors.New("cli: no prompt provided")

// ANSI escape codes for default stream rendering.
const (
	ansiReset = "\033[0m"
	ansiGrey  = "\033[90m"
	ansiGreen = "\033[92m"
	ansiCyan  = "\033[96m"
)

// RunConfig configures a single-shot CLI execution.
type RunConfig struct {
	// Engine is the inference engine to use. Required.
	Engine inference.Engine
	// SystemPrompt defines the system message for the conversation.
	SystemPrompt string
	// Registry provides tools for the agent loop. May be nil.
	Registry *tool.Registry
	// ContextProvider retrieves context before each inference call. May be nil.
	ContextProvider core.ContextProvider
	// Stream receives streaming tokens during inference. May be nil.
	Stream *inference.StreamHandler
	// Timeout limits total execution time. Zero uses DefaultTimeout (120s).
	Timeout time.Duration
	// Out is the output writer. Defaults to os.Stdout when nil.
	Out io.Writer
	// MaxTokens limits response length. Required; must be > 0.
	MaxTokens int
	// MaxToolIterations limits tool-call loop rounds. Required; must be > 0.
	MaxToolIterations int
}

// Runner executes a single inference request and writes the result.
type Runner struct {
	cfg RunConfig
}

// NewRunner creates a Runner from config.
// Returns error if Engine is nil, MaxTokens <= 0, or MaxToolIterations <= 0.
// Out defaults to os.Stdout when nil.
func NewRunner(cfg RunConfig) (*Runner, error) {
	if cfg.Engine == nil {
		return nil, fmt.Errorf("cli: RunConfig.Engine must not be nil")
	}
	if cfg.MaxTokens <= 0 {
		return nil, fmt.Errorf("cli: RunConfig.MaxTokens must be > 0")
	}
	if cfg.MaxToolIterations <= 0 {
		return nil, fmt.Errorf("cli: RunConfig.MaxToolIterations must be > 0")
	}
	if cfg.Out == nil {
		cfg.Out = os.Stdout
	}
	return &Runner{cfg: cfg}, nil
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
	// streams reasoning (grey), content, and tool calls (green) to Out.
	// ANSI colors are only used when Out is an interactive terminal.
	stream := r.cfg.Stream
	var wroteAnyStream, streamedContent bool
	var onToolResult func(string, string)
	if stream == nil {
		info := r.cfg.Engine.ModelInfo()
		fmt.Fprintf(os.Stderr, "model: %s  context: %d tokens\n\n", info.Name, info.ContextWindow)
		out := r.cfg.Out
		var grey, green, cyan, rst string
		if f, ok := out.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
			grey, green, cyan, rst = ansiGrey, ansiGreen, ansiCyan, ansiReset
		}
		stream = &inference.StreamHandler{
			OnReasoning: func(text string) {
				wroteAnyStream = true
				_, _ = fmt.Fprint(out, grey+text+rst)
			},
			OnContent: func(text string) {
				wroteAnyStream = true
				streamedContent = true
				_, _ = fmt.Fprint(out, text)
			},
			OnToolCall: func(name, argsJSON string) {
				_, _ = fmt.Fprintf(out, "\n%s[TOOL] %s(%s)%s\n", green, name, argsJSON, rst)
			},
		}
		onToolResult = func(name, output string) {
			_, _ = fmt.Fprintf(out, "%s[RESULT] %s: %s%s\n", cyan, name, output, rst)
		}
	}

	agent, err := orchestrate.NewAgentLoop(orchestrate.LoopConfig{
		SystemPrompt:      r.cfg.SystemPrompt,
		Engine:            r.cfg.Engine,
		Tools:             r.cfg.Registry,
		Stream:            stream,
		ContextProvider:   r.cfg.ContextProvider,
		MaxTokens:         r.cfg.MaxTokens,
		MaxToolIterations: r.cfg.MaxToolIterations,
		OnToolResult:      onToolResult,
	})
	if err != nil {
		return fmt.Errorf("cli: %w", err)
	}

	result, err := agent.Chat(ctx, prompt)
	if err != nil {
		return fmt.Errorf("cli: %w", err)
	}

	if r.cfg.Stream == nil {
		if streamedContent {
			// Content was streamed incrementally; add trailing newline.
			_, _ = fmt.Fprintln(r.cfg.Out)
		} else {
			if wroteAnyStream {
				// Reasoning was streamed but content was not; add separator newline.
				_, _ = fmt.Fprintln(r.cfg.Out)
			}
			if result.Content != "" {
				// No content streaming (stub engine or reasoning-only); print full content.
				_, err = fmt.Fprintln(r.cfg.Out, result.Content)
				if err != nil {
					return fmt.Errorf("cli: write output: %w", err)
				}
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
		return "", fmt.Errorf("%w (pass as arguments or pipe to stdin)", ErrNoPrompt)
	}

	data, err := io.ReadAll(stdin)
	if err != nil {
		return "", fmt.Errorf("cli: read stdin: %w", err)
	}

	prompt := strings.TrimSpace(string(data))
	if prompt == "" {
		return "", fmt.Errorf("%w (pass as arguments or pipe to stdin)", ErrNoPrompt)
	}

	return prompt, nil
}
