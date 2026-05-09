package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	inference "github.com/miroslav-matejovsky/pagantic/layers/01_inference"

	"github.com/stretchr/testify/require"
)

type stubEngine struct {
	response string
}

func (s *stubEngine) Infer(_ context.Context, _ inference.Request) (*inference.Result, error) {
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
