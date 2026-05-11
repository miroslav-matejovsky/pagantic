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

func TestNewRunner_PanicsOnNilEngine(t *testing.T) {
	require.Panics(t, func() {
		NewRunner(RunConfig{})
	})
}

func TestRunner_Run_EmptyPrompt(t *testing.T) {
	r := NewRunner(RunConfig{Engine: &stubEngine{}})

	err := r.Run(context.Background(), "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty prompt")
}

func TestRunner_Run_WhitespacePrompt(t *testing.T) {
	r := NewRunner(RunConfig{Engine: &stubEngine{}})

	err := r.Run(context.Background(), "   \n  ")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty prompt")
}

func TestRunner_Run_WritesOutput(t *testing.T) {
	var buf bytes.Buffer
	r := NewRunner(RunConfig{
		Engine: &stubEngine{response: "hello world"},
		Out:    &buf,
	})

	err := r.Run(context.Background(), "say hello")
	require.NoError(t, err)
	require.Equal(t, "hello world\n", buf.String())
}

func TestRunner_Run_NilOutDefaultsToStdout(t *testing.T) {
	// With nil Out, NewRunner should default to os.Stdout - no panic, no lost output.
	r := NewRunner(RunConfig{Engine: &stubEngine{response: "ok"}})
	require.Equal(t, os.Stdout, r.cfg.Out)
}

func TestRunner_Run_NoOutputWhenStreaming(t *testing.T) {
	var buf bytes.Buffer
	r := NewRunner(RunConfig{
		Engine: &stubEngine{response: "streamed"},
		Out:    &buf,
		Stream: &inference.StreamHandler{},
	})

	err := r.Run(context.Background(), "say hello")
	require.NoError(t, err)
	require.Empty(t, buf.String(), "should not write to Out when streaming")
}

func TestRunner_Run_AlwaysHasDeadline(t *testing.T) {
	// Engine requires a context with deadline - verify Run always sets one.
	eng := &stubEngine{response: "ok"}
	r := NewRunner(RunConfig{Engine: eng})

	err := r.Run(context.Background(), "hello")
	require.NoError(t, err)

	_, hasDeadline := eng.captCtx.Deadline()
	require.True(t, hasDeadline, "Run must always set a context deadline")
}

func TestRunner_Run_DefaultTimeoutApplied(t *testing.T) {
	eng := &stubEngine{response: "ok"}
	r := NewRunner(RunConfig{Engine: eng})

	before := time.Now().Add(DefaultTimeout)
	err := r.Run(context.Background(), "hello")
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
	r := NewRunner(RunConfig{Engine: eng, Timeout: custom})

	before := time.Now().Add(custom)
	err := r.Run(context.Background(), "hello")
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
