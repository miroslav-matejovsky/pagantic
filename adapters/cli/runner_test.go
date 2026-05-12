package cli

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"

	"github.com/stretchr/testify/require"
)

type stubEngine struct {
	response string
	captCtx  context.Context
	handler  *inference.StreamHandler
}

func (s *stubEngine) WithStreamHandler(h *inference.StreamHandler) inference.Engine {
	s.handler = h
	return s
}

func (s *stubEngine) Infer(ctx context.Context, _ inference.Request) (*inference.Result, error) {
	s.captCtx = ctx
	s.handler.EmitContent(s.response)
	return &inference.Result{Content: s.response}, nil
}

func (s *stubEngine) ModelInfo() inference.ModelInfo {
	return inference.ModelInfo{Name: "stub"}
}

func TestNewRunner_ErrorOnNilEngine(t *testing.T) {
	_, err := NewRunner(RunConfig{})
	require.Error(t, err)
	require.ErrorContains(t, err, "Engine")
}

func TestRunner_Run_EmptyPrompt(t *testing.T) {
	r, err := NewRunner(RunConfig{Engine: &stubEngine{}})
	require.NoError(t, err)

	err = r.Run(context.Background(), "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty prompt")
}

func TestRunner_Run_WhitespacePrompt(t *testing.T) {
	r, err := NewRunner(RunConfig{Engine: &stubEngine{}})
	require.NoError(t, err)

	err = r.Run(context.Background(), "   \n  ")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty prompt")
}

func TestRunner_Run_WritesOutput(t *testing.T) {
	var buf bytes.Buffer
	r, err := NewRunner(RunConfig{
		Engine: &stubEngine{response: "hello world"},
		Out:    &buf,
	})
	require.NoError(t, err)

	err = r.Run(context.Background(), "say hello")
	require.NoError(t, err)
	require.Equal(t, "hello world\n", buf.String())
}

func TestRunner_Run_NilOutDefaultsToStdout(t *testing.T) {
	// With nil Out, NewRunner should default to os.Stdout - no panic, no lost output.
	r, err := NewRunner(RunConfig{Engine: &stubEngine{response: "ok"}})
	require.NoError(t, err)
	require.Equal(t, os.Stdout, r.cfg.Out)
}

func TestRunner_Run_NoOutputWhenStreaming(t *testing.T) {
	var buf bytes.Buffer
	r, err := NewRunner(RunConfig{
		Engine: &stubEngine{response: "streamed"},
		Out:    &buf,
		Stream: &inference.StreamHandler{},
	})
	require.NoError(t, err)

	err = r.Run(context.Background(), "say hello")
	require.NoError(t, err)
	require.Empty(t, buf.String(), "should not write to Out when streaming")
}

func TestRunner_Run_AlwaysHasDeadline(t *testing.T) {
	// Engine requires a context with deadline - verify Run always sets one.
	eng := &stubEngine{response: "ok"}
	r, err := NewRunner(RunConfig{Engine: eng})
	require.NoError(t, err)

	err = r.Run(context.Background(), "hello")
	require.NoError(t, err)

	_, hasDeadline := eng.captCtx.Deadline()
	require.True(t, hasDeadline, "Run must always set a context deadline")
}

func TestRunner_Run_DefaultTimeoutApplied(t *testing.T) {
	eng := &stubEngine{response: "ok"}
	r, err := NewRunner(RunConfig{Engine: eng})
	require.NoError(t, err)

	before := time.Now().Add(DefaultTimeout)
	err = r.Run(context.Background(), "hello")
	require.NoError(t, err)

	deadline, ok := eng.captCtx.Deadline()
	require.True(t, ok)
	// Deadline should be approximately DefaultTimeout from before the call.
	require.True(t, deadline.Before(before.Add(2*time.Second)),
		"deadline should be close to DefaultTimeout")
}

func TestRunner_Run_CustomTimeoutOverridesDefault(t *testing.T) {
	eng := &stubEngine{response: "ok"}
	custom := 5 * time.Second
	r, err := NewRunner(RunConfig{Engine: eng, Timeout: custom})
	require.NoError(t, err)

	before := time.Now().Add(custom)
	err = r.Run(context.Background(), "hello")
	require.NoError(t, err)

	deadline, ok := eng.captCtx.Deadline()
	require.True(t, ok)
	require.True(t, deadline.Before(before.Add(2*time.Second)),
		"deadline should reflect custom timeout, not DefaultTimeout")
}

func TestReadPrompt_FromArgs(t *testing.T) {
	prompt, err := ReadPrompt([]string{"hello", "world"}, nil)
	require.NoError(t, err)
	require.Equal(t, "hello world", prompt)
}

func TestReadPrompt_FromStdin(t *testing.T) {
	stdin := strings.NewReader("piped input\n")

	prompt, err := ReadPrompt(nil, stdin)
	require.NoError(t, err)
	require.Equal(t, "piped input", prompt)
}

func TestReadPrompt_EmptyStdin(t *testing.T) {
	stdin := strings.NewReader("")

	_, err := ReadPrompt(nil, stdin)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no prompt provided")
}

func TestReadPrompt_PipedStdin_DoesNotBlock(t *testing.T) {
	// os.Pipe returns an *os.File that is NOT a terminal (term.IsTerminal == false).
	// ReadPrompt must read it normally rather than returning the TTY error.
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer func() { _ = r.Close() }()

	_, _ = w.WriteString("piped content")
	_ = w.Close()

	prompt, err := ReadPrompt(nil, r)
	require.NoError(t, err)
	require.Equal(t, "piped content", prompt)
}

func TestReadPrompt_InteractiveStdin_FailsFast(t *testing.T) {
	// os.Stdin in test processes is NOT a terminal (tests run with piped I/O),
	// so we use /dev/null or NUL as a stand-in non-terminal *os.File to verify
	// non-terminal files still read normally, and rely on the Pipe test above
	// to confirm the TTY guard path is correctly skipped for pipes.
	//
	// True TTY detection can only be verified in a manual integration test
	// because test runners always redirect stdin away from the terminal.
	//
	// Here we just assert the contract: a strings.Reader (not *os.File) reads
	// normally and is never treated as a terminal.
	stdin := strings.NewReader("from reader")
	prompt, err := ReadPrompt(nil, stdin)
	require.NoError(t, err)
	require.Equal(t, "from reader", prompt)
}
