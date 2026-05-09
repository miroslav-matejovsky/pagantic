package tui

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestREPL_DispatchesCommand(t *testing.T) {
	var called bool
	r := NewREPL("> ")
	r.In = strings.NewReader("greet\nquit\n")
	r.Out = &bytes.Buffer{}
	r.ErrOut = &bytes.Buffer{}

	r.AddCommand(Command{
		Name:        "greet",
		Description: "Say hello",
		Run: func(_ context.Context, _ []string) error {
			called = true
			return nil
		},
	})

	r.Run(context.Background())

	if !called {
		t.Error("command was not dispatched")
	}
}

func TestREPL_CommandArgs(t *testing.T) {
	var gotArgs []string
	r := NewREPL("> ")
	r.In = strings.NewReader("do foo bar\nquit\n")
	r.Out = &bytes.Buffer{}
	r.ErrOut = &bytes.Buffer{}

	r.AddCommand(Command{
		Name: "do",
		Run: func(_ context.Context, args []string) error {
			gotArgs = args
			return nil
		},
	})

	r.Run(context.Background())

	if len(gotArgs) != 2 || gotArgs[0] != "foo" || gotArgs[1] != "bar" {
		t.Errorf("args = %v, want [foo bar]", gotArgs)
	}
}

func TestREPL_Aliases(t *testing.T) {
	var count int
	r := NewREPL("> ")
	r.In = strings.NewReader("g\ngreet\nquit\n")
	r.Out = &bytes.Buffer{}
	r.ErrOut = &bytes.Buffer{}

	r.AddCommand(Command{
		Name:    "greet",
		Aliases: []string{"g"},
		Run: func(_ context.Context, _ []string) error {
			count++
			return nil
		},
	})

	r.Run(context.Background())

	if count != 2 {
		t.Errorf("command called %d times, want 2", count)
	}
}

func TestREPL_AutoHelp(t *testing.T) {
	var out bytes.Buffer
	r := NewREPL("> ")
	r.In = strings.NewReader("help\nquit\n")
	r.Out = &out
	r.ErrOut = &bytes.Buffer{}

	r.AddCommand(Command{
		Name:        "test",
		Description: "Run tests",
		Run:         func(_ context.Context, _ []string) error { return nil },
	})

	r.Run(context.Background())

	if !strings.Contains(out.String(), "test") {
		t.Error("help output missing registered command")
	}
	if !strings.Contains(out.String(), "quit") {
		t.Error("help output missing quit command")
	}
}

func TestREPL_UnknownCommand(t *testing.T) {
	var out bytes.Buffer
	r := NewREPL("> ")
	r.In = strings.NewReader("nope\nquit\n")
	r.Out = &out
	r.ErrOut = &bytes.Buffer{}

	r.Run(context.Background())

	if !strings.Contains(out.String(), "Unknown command: nope") {
		t.Errorf("expected unknown command message, got: %s", out.String())
	}
}

func TestREPL_QuitVariants(t *testing.T) {
	for _, cmd := range []string{"quit", "exit", "q"} {
		var out bytes.Buffer
		r := NewREPL("> ")
		r.In = strings.NewReader(cmd + "\n")
		r.Out = &out
		r.ErrOut = &bytes.Buffer{}

		r.Run(context.Background())
		// Should exit without error - test passes if Run returns.
	}
}

func TestREPL_EOF(t *testing.T) {
	r := NewREPL("> ")
	r.In = strings.NewReader("")
	r.Out = &bytes.Buffer{}
	r.ErrOut = &bytes.Buffer{}

	r.Run(context.Background())
	// Should exit cleanly on EOF.
}

func TestREPL_CommandError(t *testing.T) {
	var errOut bytes.Buffer
	r := NewREPL("> ")
	r.In = strings.NewReader("fail\nquit\n")
	r.Out = &bytes.Buffer{}
	r.ErrOut = &errOut

	r.AddCommand(Command{
		Name: "fail",
		Run: func(_ context.Context, _ []string) error {
			return context.DeadlineExceeded
		},
	})

	r.Run(context.Background())

	got := SanitizeOutput(errOut.String())
	if !strings.Contains(got, "context deadline exceeded") {
		t.Errorf("expected error in errOut, got: %q", got)
	}
}

func TestREPL_CaseInsensitive(t *testing.T) {
	var called bool
	r := NewREPL("> ")
	r.In = strings.NewReader("GREET\nquit\n")
	r.Out = &bytes.Buffer{}
	r.ErrOut = &bytes.Buffer{}

	r.AddCommand(Command{
		Name: "greet",
		Run: func(_ context.Context, _ []string) error {
			called = true
			return nil
		},
	})

	r.Run(context.Background())

	if !called {
		t.Error("case-insensitive dispatch failed")
	}
}

func TestREPL_Banner(t *testing.T) {
	var out bytes.Buffer
	r := NewREPL("> ")
	r.In = strings.NewReader("quit\n")
	r.Out = &out
	r.ErrOut = &bytes.Buffer{}
	r.SetBanner("=== My Tool ===")

	r.Run(context.Background())

	if !strings.Contains(out.String(), "=== My Tool ===") {
		t.Errorf("banner not printed, got: %s", out.String())
	}
}

func TestREPL_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r := NewREPL("> ")
	r.In = strings.NewReader("greet\nquit\n")
	r.Out = &bytes.Buffer{}
	r.ErrOut = &bytes.Buffer{}

	r.Run(ctx)
	// Should return immediately without dispatching.
}

func TestREPL_OnUnknown(t *testing.T) {
	var got string
	r := NewREPL("> ")
	r.In = strings.NewReader("xyz\nquit\n")
	r.Out = &bytes.Buffer{}
	r.ErrOut = &bytes.Buffer{}
	r.OnUnknown = func(cmd string) {
		got = cmd
	}

	r.Run(context.Background())

	if got != "xyz" {
		t.Errorf("OnUnknown got %q, want %q", got, "xyz")
	}
}

func TestREPL_AddCommand_PanicsOnEmptyName(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty Command.Name")
		}
	}()
	NewREPL("> ").AddCommand(Command{
		Name: "",
		Run:  func(_ context.Context, _ []string) error { return nil },
	})
}

func TestREPL_AddCommand_PanicsOnNilRun(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil Command.Run")
		}
	}()
	NewREPL("> ").AddCommand(Command{Name: "cmd"})
}

func TestREPL_HelpWithArgs(t *testing.T) {
	var out bytes.Buffer
	r := NewREPL("> ")
	r.In = strings.NewReader("help\nquit\n")
	r.Out = &out
	r.ErrOut = &bytes.Buffer{}

	r.AddCommand(Command{
		Name:        "analyze",
		Args:        "<repo-url>",
		Description: "Analyze a repo",
		Run:         func(_ context.Context, _ []string) error { return nil },
	})

	r.Run(context.Background())

	if !strings.Contains(out.String(), "<repo-url>") {
		t.Error("help output missing Args hint")
	}
}
