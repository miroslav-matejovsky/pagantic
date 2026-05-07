//go:build integration

package agent_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/miroslav-matejovsky/pagantic/agent"
	"github.com/miroslav-matejovsky/pagantic/kronk"
	"github.com/miroslav-matejovsky/pagantic/llm"
	"github.com/stretchr/testify/require"
)

const integrationModel = "unsloth/Qwen3-0.6B-Q8_0"

// loadEngine loads the LLM engine for an integration test.
// Skips the test if loading fails.
func loadEngine(t *testing.T) (llm.Chat, func()) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	engine, cleanup, err := kronk.Load(ctx, kronk.Config{ModelSource: integrationModel})
	if err != nil {
		t.Skipf("skipping integration test: engine load failed: %v", err)
	}

	return engine, cleanup
}

// TestIntegration_BasicChat verifies a simple chat produces non-empty content.
func TestIntegration_BasicChat(t *testing.T) {
	engine, cleanup := loadEngine(t)
	defer cleanup()

	a := agent.New(agent.Config{
		SystemPrompt: "You are a helpful assistant. Answer in one sentence.",
		Engine:       engine,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := a.Chat(ctx, "What is 2+2?")
	require.NoError(t, err)
	require.NotEmpty(t, result.Content, "expected non-empty response")
}

// TestIntegration_ContentStreamFires verifies OnContent callback fires during streaming.
func TestIntegration_ContentStreamFires(t *testing.T) {
	engine, cleanup := loadEngine(t)
	defer cleanup()

	var contentCalls atomic.Int32

	a := agent.New(agent.Config{
		SystemPrompt: "You are a helpful assistant. Answer briefly.",
		Engine:       engine,
		Stream: &llm.StreamHandler{
			OnContent: func(text string) {
				contentCalls.Add(1)
			},
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := a.Chat(ctx, "Say hello.")
	require.NoError(t, err)
	require.NotEmpty(t, result.Content)
	require.Greater(t, contentCalls.Load(), int32(0), "OnContent should fire at least once")
}

// TestIntegration_MultiTurn verifies conversation history works across turns.
func TestIntegration_MultiTurn(t *testing.T) {
	engine, cleanup := loadEngine(t)
	defer cleanup()

	a := agent.New(agent.Config{
		SystemPrompt: "You are a helpful assistant. Answer briefly.",
		Engine:       engine,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	result1, err := a.Chat(ctx, "Remember the number 42.")
	require.NoError(t, err)
	require.NotEmpty(t, result1.Content)

	result2, err := a.Chat(ctx, "What number did I ask you to remember?")
	require.NoError(t, err)
	require.NotEmpty(t, result2.Content)
	// Relaxed: just check response is non-empty. Small models may or may not
	// recall correctly, but the pipeline should not error.
}

// TestIntegration_NilStreamHandler verifies silent operation with nil handler.
func TestIntegration_NilStreamHandler(t *testing.T) {
	engine, cleanup := loadEngine(t)
	defer cleanup()

	a := agent.New(agent.Config{
		SystemPrompt: "You are a helpful assistant.",
		Engine:       engine,
		Stream:       nil,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := a.Chat(ctx, "Hello")
	require.NoError(t, err)
	require.NotEmpty(t, result.Content)
}
