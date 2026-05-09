package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	orchestrate "github.com/miroslav-matejovsky/pagantic/layers/02_orchestrate"
	tool "github.com/miroslav-matejovsky/pagantic/layers/04_tool"
)

// RunConfig configures a single-shot CLI execution.
type RunConfig struct {
	// Engine is the inference engine to use. Required.
	Engine inference.Engine
	// SystemPrompt defines the system message for the conversation.
	SystemPrompt string
	// Registry provides tools for the agent loop. May be nil.
	Registry *tool.Registry
	// Stream receives streaming tokens during inference. May be nil.
	Stream *inference.StreamHandler
	// Timeout limits total execution time. Zero means no timeout.
	Timeout time.Duration
	// Out is the output writer. Defaults to os.Stdout if nil.
	Out io.Writer
}

// Runner executes a single inference request and writes the result.
type Runner struct {
	cfg RunConfig
}

// NewRunner creates a Runner from config. Panics if Engine is nil.
func NewRunner(cfg RunConfig) *Runner {
	if cfg.Engine == nil {
		panic("cli: RunConfig.Engine must not be nil")
	}
	return &Runner{cfg: cfg}
}

// Run executes a single inference call with the given prompt and writes
// the response content to the configured output writer.
func (r *Runner) Run(ctx context.Context, prompt string) error {
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("cli: empty prompt")
	}

	if r.cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.cfg.Timeout)
		defer cancel()
	}

	agent := orchestrate.NewAgentLoop(orchestrate.LoopConfig{
		SystemPrompt: r.cfg.SystemPrompt,
		Engine:       r.cfg.Engine,
		Tools:        r.cfg.Registry,
		Stream:       r.cfg.Stream,
	})

	result, err := agent.Chat(ctx, prompt)
	if err != nil {
		return fmt.Errorf("cli: %w", err)
	}

	if r.cfg.Out != nil && r.cfg.Stream == nil {
		_, err = fmt.Fprintln(r.cfg.Out, result.Content)
		if err != nil {
			return fmt.Errorf("cli: write output: %w", err)
		}
	}

	return nil
}

// ReadPrompt builds a prompt from command-line arguments. If args is empty,
// reads all of stdin. Returns the trimmed prompt string.
func ReadPrompt(args []string, stdin io.Reader) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
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
