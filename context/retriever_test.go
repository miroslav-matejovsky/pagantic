package context

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInMemoryRetrieverFindsRelevantDocuments(t *testing.T) {
	t.Parallel()

	r := NewInMemoryRetriever(
		Document{Content: "Go context builder for query handling", Source: "builder"},
		Document{Content: "Context stays bounded and verified", Source: "policy"},
		Document{Content: "Tracing and events", Source: "observe"},
	)

	chunks, err := r.Retrieve(context.Background(), "context builder", 10)
	require.NoError(t, err)
	require.Len(t, chunks, 2)
	require.Equal(t, "builder", chunks[0].Source)
	require.Equal(t, "policy", chunks[1].Source)
}

func TestInMemoryRetrieverReturnsEmptyForNoMatches(t *testing.T) {
	t.Parallel()

	r := NewInMemoryRetriever(
		Document{Content: "Tracing and events", Source: "observe"},
	)

	chunks, err := r.Retrieve(context.Background(), "schema prompt", 10)
	require.NoError(t, err)
	require.Empty(t, chunks)
}

func TestInMemoryRetrieverRespectsLimit(t *testing.T) {
	t.Parallel()

	r := NewInMemoryRetriever(
		Document{Content: "context alpha", Source: "one"},
		Document{Content: "context beta", Source: "two"},
		Document{Content: "context gamma", Source: "three"},
	)

	chunks, err := r.Retrieve(context.Background(), "context", 2)
	require.NoError(t, err)
	require.Len(t, chunks, 2)
}

func TestInMemoryRetrieverScoresHigherForMoreKeywordMatches(t *testing.T) {
	t.Parallel()

	r := NewInMemoryRetriever(
		Document{Content: "context builder retrieval", Source: "full"},
		Document{Content: "context only", Source: "partial"},
	)

	chunks, err := r.Retrieve(context.Background(), "context builder retrieval", 10)
	require.NoError(t, err)
	require.Len(t, chunks, 2)
	require.Equal(t, "full", chunks[0].Source)
	require.Equal(t, "partial", chunks[1].Source)
	require.Greater(t, chunks[0].Score, chunks[1].Score)
}

func TestInMemoryRetrieverHandlesEmptyDocumentList(t *testing.T) {
	t.Parallel()

	r := NewInMemoryRetriever()

	chunks, err := r.Retrieve(context.Background(), "context", 10)
	require.NoError(t, err)
	require.Empty(t, chunks)
}
