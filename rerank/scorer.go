package rerank

import "context"

// Candidate represents a scored item for ranking.
type Candidate struct {
	Content  string
	Score    float64
	Source   string
	Metadata map[string]any
}

// CandidateSet groups candidates with original query.
type CandidateSet struct {
	Query      string
	Candidates []Candidate
}

// RelevanceScorer assigns relevance scores to candidates.
type RelevanceScorer interface {
	Score(ctx context.Context, query string, candidates []Candidate) ([]Candidate, error)
}
