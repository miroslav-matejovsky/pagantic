package tui

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Command defines a single REPL command.
type Command struct {
	// Name is the primary command name used for matching input.
	Name string
	// Aliases are alternate names that also trigger this command.
	Aliases []string
	// Description is a short summary shown in help output.
	Description string
	// Args is an argument hint shown after the command name in help
	// (e.g. "<repo-url>"). Leave empty for commands that take no arguments.
	Args string
	// Run executes the command. args contains everything after the command
	// name, split by whitespace. Return a non-nil error to print it and
	// continue the REPL loop.
	Run func(ctx context.Context, args []string) error
}

// matches reports whether input matches this command's name or any alias.
func (c Command) matches(input string) bool {
	if strings.EqualFold(c.Name, input) {
		return true
	}
	for _, a := range c.Aliases {
		if strings.EqualFold(a, input) {
			return true
		}
	}
	return false
}

// REPL is a generic command-dispatch read-eval-print loop.
//
// Commands are registered with AddCommand. The loop reads lines from In,
// dispatches to matching commands, and prints errors to ErrOut. The built-in
// quit/exit/q commands stop the loop. If no explicit "help" command is
// registered, an auto-generated help listing is provided.
//
// Input parsing uses strings.Fields, so quoted arguments are not supported.
type REPL struct {
	commands []Command
	prompt   string
	banner   string

	// OnUnknown is called when input does not match any command.
	// When nil, a default "Unknown command" message is printed.
	OnUnknown func(cmd string)

	// In is the input source. Defaults to os.Stdin.
	In io.Reader
	// Out is the output destination. Defaults to os.Stdout.
	Out io.Writer
	// ErrOut is the error output destination. Defaults to os.Stderr.
	ErrOut io.Writer
}

// NewREPL creates a REPL with the given prompt string.
// Input and output default to stdin/stdout/stderr.
func NewREPL(prompt string) *REPL {
	return &REPL{
		prompt: prompt,
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
}

// SetBanner sets text printed once at startup before the first prompt.
func (r *REPL) SetBanner(banner string) {
	r.banner = banner
}

// AddCommand registers a command. Commands are matched in registration order.
// Panics if cmd.Name is empty or cmd.Run is nil.
func (r *REPL) AddCommand(cmd Command) {
	if cmd.Name == "" {
		panic("tui: Command.Name must not be empty")
	}
	if cmd.Run == nil {
		panic("tui: Command.Run must not be nil")
	}
	r.commands = append(r.commands, cmd)
}

// Run starts the REPL loop. It blocks until the user types quit/exit/q,
// the input reaches EOF, or the context is cancelled.
func (r *REPL) Run(ctx context.Context) {
	if r.banner != "" {
		_, _ = fmt.Fprintln(r.Out, r.banner)
	}

	if !r.hasCommand("help") {
		r.printHelp()
		_, _ = fmt.Fprintln(r.Out)
	}

	scanner := bufio.NewScanner(r.In)

	for {
		if ctx.Err() != nil {
			return
		}

		line, err := FPrompt(scanner, r.Out, r.prompt)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				FErrorf(r.ErrOut, "input error: %v", err)
			}
			return
		}

		parts := strings.Fields(line)
		cmd := strings.ToLower(parts[0])
		args := parts[1:]

		if cmd == "quit" || cmd == "exit" || cmd == "q" {
			return
		}

		if cmd == "help" || cmd == "h" || cmd == "?" {
			if r.hasCommand("help") {
				r.dispatch(ctx, "help", args)
			} else {
				r.printHelp()
			}
			continue
		}

		r.dispatch(ctx, cmd, args)
	}
}

func (r *REPL) dispatch(ctx context.Context, cmd string, args []string) {
	for _, c := range r.commands {
		if c.matches(cmd) {
			if err := c.Run(ctx, args); err != nil {
				FErrorf(r.ErrOut, "%v", err)
			}
			return
		}
	}

	if r.OnUnknown != nil {
		r.OnUnknown(cmd)
	} else {
		FWarnf(r.Out, "Unknown command: %s. Type 'help' for commands.", cmd)
	}
}

func (r *REPL) hasCommand(name string) bool {
	for _, c := range r.commands {
		if c.matches(name) {
			return true
		}
	}
	return false
}

func (r *REPL) printHelp() {
	_, _ = fmt.Fprintln(r.Out, "Commands:")
	for _, c := range r.commands {
		suffix := ""
		if c.Args != "" {
			suffix = " " + c.Args
		}
		_, _ = fmt.Fprintf(r.Out, "  %-20s %s\n", c.Name+suffix, c.Description)
	}
	_, _ = fmt.Fprintf(r.Out, "  %-20s %s\n", "quit", "Exit")
}
