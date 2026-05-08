package tui

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"
	"github.com/miroslav-matejovsky/pagantic/tool"
)

type fakeEngine struct{}

func (*fakeEngine) Infer(context.Context, inference.Request) (*inference.Result, error) {
	return &inference.Result{}, nil
}

func (*fakeEngine) ModelInfo() inference.ModelInfo {
	return inference.ModelInfo{Name: "fake"}
}

// stubRegistry is a minimal *tool.Registry substitute for testing.
// We build a real registry with no tools to satisfy the type requirement.
var stubRegistry = tool.NewRegistry()

func stubLoader(engine inference.Engine) func(context.Context) (inference.Engine, func(), error) {
	return func(_ context.Context) (inference.Engine, func(), error) {
		return engine, nil, nil
	}
}

func errLoader(err error) func(context.Context) (inference.Engine, func(), error) {
	return func(_ context.Context) (inference.Engine, func(), error) {
		return nil, nil, err
	}
}

func newTestREPL(loader func(context.Context) (inference.Engine, func(), error)) (*AgentREPL, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	t := NewAgentREPL(AgentConfig{
		Title:        "Test",
		SystemPrompt: "You are a test agent.",
		EngineLoader: loader,
		Registry:     stubRegistry,
	})
	t.repl.Out = out
	t.repl.ErrOut = errOut
	return t, out, errOut
}

func TestNewAgentREPL_PanicsOnNilEngineLoader(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil EngineLoader")
		}
	}()
	NewAgentREPL(AgentConfig{Registry: stubRegistry})
}

func TestNewAgentREPL_PanicsOnNilRegistry(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil Registry")
		}
	}()
	NewAgentREPL(AgentConfig{
		EngineLoader: stubLoader(nil),
	})
}

func TestAgentREPL_DefaultLocalDir(t *testing.T) {
	ar, _, _ := newTestREPL(stubLoader(nil))
	if ar.LocalDir() != ".local" {
		t.Errorf("LocalDir() = %q, want %q", ar.LocalDir(), ".local")
	}
}

func TestAgentREPL_CustomLocalDir(t *testing.T) {
	ar := NewAgentREPL(AgentConfig{
		Title:        "Test",
		SystemPrompt: "x",
		EngineLoader: stubLoader(&fakeEngine{}),
		Registry:     stubRegistry,
		LocalDir:     "C:\\work\\mydir",
	})
	if ar.LocalDir() != "C:\\work\\mydir" {
		t.Errorf("LocalDir() = %q, want C:\\work\\mydir", ar.LocalDir())
	}
}

func TestAgentREPL_EngineCalledOnce(t *testing.T) {
	calls := 0
	loader := func(_ context.Context) (inference.Engine, func(), error) {
		calls++
		return nil, nil, nil
	}
	ar, _, _ := newTestREPL(loader)

	_, _ = ar.Engine(context.Background())
	_, _ = ar.Engine(context.Background())
	_, _ = ar.Engine(context.Background())

	if calls != 1 {
		t.Errorf("EngineLoader called %d times, want 1", calls)
	}
}

func TestAgentREPL_EngineErrorPropagation(t *testing.T) {
	want := errors.New("engine failed")
	ar, _, _ := newTestREPL(errLoader(want))

	_, err := ar.Engine(context.Background())
	if !errors.Is(err, want) {
		t.Errorf("Engine() error = %v, want %v", err, want)
	}
}

func TestAgentREPL_EngineErrorNotCached(t *testing.T) {
	// A failed load should allow retry on next call.
	calls := 0
	loader := func(_ context.Context) (inference.Engine, func(), error) {
		calls++
		if calls == 1 {
			return nil, nil, errors.New("transient error")
		}
		return nil, nil, nil
	}
	ar, _, _ := newTestREPL(loader)

	_, err := ar.Engine(context.Background())
	if err == nil {
		t.Fatal("expected error on first call")
	}
	_, err = ar.Engine(context.Background())
	if err != nil {
		t.Errorf("expected nil error on retry, got %v", err)
	}
	if calls != 2 {
		t.Errorf("loader called %d times, want 2", calls)
	}
}

func TestAgentREPL_CommandsRegistered(t *testing.T) {
	ar, out, _ := newTestREPL(stubLoader(nil))
	ar.repl.In = strings.NewReader("help\nquit\n")

	ar.Run(context.Background())

	got := SanitizeOutput(out.String())
	if !strings.Contains(got, "tools") {
		t.Error("help output missing 'tools' command")
	}
	if !strings.Contains(got, "chat") {
		t.Error("help output missing 'chat' command")
	}
}

func TestAgentREPL_AddCommand(t *testing.T) {
	var called bool
	ar, _, _ := newTestREPL(stubLoader(nil))
	ar.repl.In = strings.NewReader("ping\nquit\n")

	ar.AddCommand(Command{
		Name: "ping",
		Run: func(_ context.Context, _ []string) error {
			called = true
			return nil
		},
	})

	ar.Run(context.Background())

	if !called {
		t.Error("custom command was not dispatched")
	}
}

func TestAgentREPL_RunPassesContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	ar, _, _ := newTestREPL(stubLoader(nil))
	ar.repl.In = strings.NewReader("ping\nquit\n")

	// Should return immediately without dispatching due to cancelled context.
	ar.Run(ctx)
}

func TestAgentREPL_LoadingWarningToErrOut(t *testing.T) {
	ar, _, errOut := newTestREPL(stubLoader(nil))

	_, _ = ar.Engine(context.Background())

	got := SanitizeOutput(errOut.String())
	if !strings.Contains(got, "Loading inference engine") {
		t.Errorf("expected loading warning in errOut, got: %q", got)
	}
}

// -- Convenience wrapper smoke tests --

func TestInfof_Smoke(t *testing.T) { Infof("test %d", 1) }
func TestWarnf_Smoke(t *testing.T) { Warnf("test %d", 2) }
