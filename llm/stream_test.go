package llm_test

import (
	"context"
	"testing"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
	"github.com/miroslav-matejovsky/pagantic/llm"
	"github.com/stretchr/testify/require"
)

// fakeEngine satisfies llm.Chat for testing.
type fakeEngine struct{}

func (e *fakeEngine) ChatStreaming(_ context.Context, _ model.D) (<-chan model.ChatResponse, error) {
	return nil, nil
}

func (e *fakeEngine) ModelConfig() model.Config { return model.Config{} }

func strPtr(s string) *string { return &s }

// makeChannel returns a closed channel pre-loaded with the given responses.
func makeChannel(responses ...model.ChatResponse) <-chan model.ChatResponse {
	ch := make(chan model.ChatResponse, len(responses))
	for _, r := range responses {
		ch <- r
	}
	close(ch)
	return ch
}

// TestStreamResponse_ContentInBodyChunk verifies that content tokens delivered
// in non-stop chunks are captured correctly.
func TestStreamResponse_ContentInBodyChunk(t *testing.T) {
	ch := makeChannel(
		model.ChatResponse{Choices: []model.Choice{
			{Delta: &model.ResponseMessage{Content: "hello"}},
		}},
		model.ChatResponse{
			Choices: []model.Choice{{FinishReasonPtr: strPtr(model.FinishReasonStop)}},
			Usage:   &model.Usage{},
		},
	)

	result, err := llm.StreamResponse(&fakeEngine{}, model.DocumentArray(), ch, nil)

	require.NoError(t, err)
	require.Equal(t, "hello", result.Content)
}

// TestStreamResponse_ContentInStopChunk is the regression test for the bug where
// the final token (e.g., closing `}` of JSON) arrives in the same chunk as
// FinishReasonStop and was silently dropped.
func TestStreamResponse_ContentInStopChunk(t *testing.T) {
	ch := makeChannel(
		model.ChatResponse{Choices: []model.Choice{
			{Delta: &model.ResponseMessage{Content: `{"value": 1`}},
		}},
		// Final chunk: content AND stop reason in the same message.
		model.ChatResponse{
			Choices: []model.Choice{{
				Delta:           &model.ResponseMessage{Content: "}"},
				FinishReasonPtr: strPtr(model.FinishReasonStop),
			}},
			Usage: &model.Usage{},
		},
	)

	result, err := llm.StreamResponse(&fakeEngine{}, model.DocumentArray(), ch, nil)

	require.NoError(t, err)
	require.Equal(t, `{"value": 1}`, result.Content)
}

// TestStreamResponse_EmptyStopChunk verifies stop with no content still works.
func TestStreamResponse_EmptyStopChunk(t *testing.T) {
	ch := makeChannel(
		model.ChatResponse{Choices: []model.Choice{
			{Delta: &model.ResponseMessage{Content: "done"}},
		}},
		model.ChatResponse{
			Choices: []model.Choice{{FinishReasonPtr: strPtr(model.FinishReasonStop)}},
			Usage:   &model.Usage{},
		},
	)

	result, err := llm.StreamResponse(&fakeEngine{}, model.DocumentArray(), ch, nil)

	require.NoError(t, err)
	require.Equal(t, "done", result.Content)
}

// is invoked for content arriving in the stop chunk.
func TestStreamResponse_OnTokenCalledForStopChunkContent(t *testing.T) {
	var tokens []string
	onToken := func(kind, text string) {
		if kind == "content" {
			tokens = append(tokens, text)
		}
	}

	ch := makeChannel(
		model.ChatResponse{Choices: []model.Choice{
			{Delta: &model.ResponseMessage{Content: "ab"}},
		}},
		model.ChatResponse{
			Choices: []model.Choice{{
				Delta:           &model.ResponseMessage{Content: "c"},
				FinishReasonPtr: strPtr(model.FinishReasonStop),
			}},
			Usage: &model.Usage{},
		},
	)

	result, err := llm.StreamResponse(&fakeEngine{}, model.DocumentArray(), ch, onToken)

	require.NoError(t, err)
	require.Equal(t, "abc", result.Content)
	require.Equal(t, []string{"ab", "c"}, tokens)
}
