package rerank

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type recordingScorer struct {
	called bool
	fn     func(ctx context.Context, query string, candidates []Candidate) ([]Candidate, error)
}

func (s *recordingScorer) Score(ctx context.Context, query string, candidates []Candidate) ([]Candidate, error) {
	s.called = true
	return s.fn(ctx, query, candidates)
}

func TestReranker_ReranksCandidatesByScore(t *testing.T) {
	scorer := &recordingScorer{
		fn: func(ctx context.Context, query string, candidates []Candidate) ([]Candidate, error) {
			require.Equal(t, "go query", query)
			require.Len(t, candidates, 3)
			return append([]Candidate(nil), candidates...), nil
		},
	}

	r := &Reranker{Scorer: scorer}
	result, err := r.Rerank(context.Background(), CandidateSet{
		Query: "go query",
		Candidates: []Candidate{
			{Content: "low", Score: 0.2},
			{Content: "high", Score: 0.9},
			{Content: "mid", Score: 0.5},
		},
	})

	require.NoError(t, err)
	require.True(t, scorer.called)
	require.Equal(t, []Candidate{
		{Content: "high", Score: 0.9},
		{Content: "mid", Score: 0.5},
		{Content: "low", Score: 0.2},
	}, result)
}

func TestReranker_AppliesTopKFilter(t *testing.T) {
	r := &Reranker{
		Scorer: &recordingScorer{
			fn: func(ctx context.Context, query string, candidates []Candidate) ([]Candidate, error) {
				return append([]Candidate(nil), candidates...), nil
			},
		},
		Policy: SelectionPolicy{TopK: 2},
	}

	result, err := r.Rerank(context.Background(), CandidateSet{
		Candidates: []Candidate{
			{Content: "third", Score: 0.2},
			{Content: "first", Score: 0.9},
			{Content: "second", Score: 0.7},
		},
	})

	require.NoError(t, err)
	require.Equal(t, []Candidate{
		{Content: "first", Score: 0.9},
		{Content: "second", Score: 0.7},
	}, result)
}

func TestReranker_AppliesMinScoreFilter(t *testing.T) {
	r := &Reranker{
		Scorer: &recordingScorer{
			fn: func(ctx context.Context, query string, candidates []Candidate) ([]Candidate, error) {
				return append([]Candidate(nil), candidates...), nil
			},
		},
		Policy: SelectionPolicy{MinScore: 0.5},
	}

	result, err := r.Rerank(context.Background(), CandidateSet{
		Candidates: []Candidate{
			{Content: "keep-high", Score: 0.9},
			{Content: "drop", Score: 0.2},
			{Content: "keep-edge", Score: 0.5},
		},
	})

	require.NoError(t, err)
	require.Equal(t, []Candidate{
		{Content: "keep-high", Score: 0.9},
		{Content: "keep-edge", Score: 0.5},
	}, result)
}

func TestReranker_HandlesEmptyCandidateSet(t *testing.T) {
	scorer := &recordingScorer{
		fn: func(ctx context.Context, query string, candidates []Candidate) ([]Candidate, error) {
			require.Empty(t, candidates)
			return append([]Candidate(nil), candidates...), nil
		},
	}

	r := &Reranker{Scorer: scorer}
	result, err := r.Rerank(context.Background(), CandidateSet{Query: "empty"})

	require.NoError(t, err)
	require.True(t, scorer.called)
	require.Empty(t, result)
}

func TestReranker_CombinesTopKAndMinScore(t *testing.T) {
	r := &Reranker{
		Scorer: &recordingScorer{
			fn: func(ctx context.Context, query string, candidates []Candidate) ([]Candidate, error) {
				return append([]Candidate(nil), candidates...), nil
			},
		},
		Policy: SelectionPolicy{TopK: 2, MinScore: 0.75},
	}

	result, err := r.Rerank(context.Background(), CandidateSet{
		Candidates: []Candidate{
			{Content: "best", Score: 0.95},
			{Content: "good", Score: 0.8},
			{Content: "too-low", Score: 0.7},
			{Content: "great", Score: 0.9},
		},
	})

	require.NoError(t, err)
	require.Equal(t, []Candidate{
		{Content: "best", Score: 0.95},
		{Content: "great", Score: 0.9},
	}, result)
}
