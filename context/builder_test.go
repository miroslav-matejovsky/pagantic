package context

import (
	"context"
	"testing"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	"github.com/stretchr/testify/require"
)

// stubRetriever gives fixed chunks for builder tests.
type stubRetriever struct {
	chunks []Chunk
}

func (r stubRetriever) Retrieve(_ context.Context, _ string, _ int) ([]Chunk, error) {
	return r.chunks, nil
}

func TestContextBuilderBuildsContextMessages(t *testing.T) {
	t.Parallel()

	builder := ContextBuilder{
		Retriever: stubRetriever{chunks: []Chunk{{Content: "bounded knowledge", Source: "doc"}}},
		MaxChunks: 10,
	}

	messages, err := builder.Build(context.Background(), "knowledge")
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Equal(t, core.RoleSystem, messages[0].Role)
	require.Contains(t, messages[0].Content, "Relevant context:")
	require.Contains(t, messages[0].Content, "bounded knowledge")
}

func TestContextBuilderReturnsEmptyWhenNoChunksFound(t *testing.T) {
	t.Parallel()

	builder := ContextBuilder{
		Retriever: stubRetriever{},
	}

	messages, err := builder.Build(context.Background(), "knowledge")
	require.NoError(t, err)
	require.Empty(t, messages)
}

func TestContextBuilderFormatsChunksWithSourceAttribution(t *testing.T) {
	t.Parallel()

	builder := ContextBuilder{
		Retriever: stubRetriever{chunks: []Chunk{
			{Content: "first chunk", Source: "doc-one"},
			{Content: "second chunk", Source: "doc-two"},
		}},
	}

	messages, err := builder.Build(context.Background(), "chunk")
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Contains(t, messages[0].Content, "[1] (doc-one): first chunk")
	require.Contains(t, messages[0].Content, "[2] (doc-two): second chunk")
}
