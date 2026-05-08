package rerank

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSimpleScorer_ScoresHigherForMoreKeywordMatches(t *testing.T) {
	scorer := &SimpleScorer{}

	result, err := scorer.Score(context.Background(), "Go scorer", []Candidate{
		{Content: "Go scorer package"},
		{Content: "Go package"},
	})

	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Greater(t, result[0].Score, result[1].Score)
	require.Equal(t, 1.0, result[0].Score)
	require.Equal(t, 0.5, result[1].Score)
}

func TestSimpleScorer_ZeroScoreForNoMatches(t *testing.T) {
	scorer := &SimpleScorer{}

	result, err := scorer.Score(context.Background(), "go scorer", []Candidate{
		{Content: "python package"},
	})

	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Zero(t, result[0].Score)
}

func TestSimpleScorer_HandlesEmptyQuery(t *testing.T) {
	scorer := &SimpleScorer{}

	result, err := scorer.Score(context.Background(), "", []Candidate{
		{Content: "Go scorer package", Score: 9},
	})

	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Zero(t, result[0].Score)
}

func TestSimpleScorer_HandlesEmptyCandidates(t *testing.T) {
	scorer := &SimpleScorer{}

	result, err := scorer.Score(context.Background(), "go scorer", nil)

	require.NoError(t, err)
	require.Empty(t, result)
}
