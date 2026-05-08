package rerank

import (
	"context"
	"errors"
	"sort"
)

// Reranker combines a scorer with a selection policy to produce
// a refined, filtered candidate list.
type Reranker struct {
	Scorer RelevanceScorer
	Policy SelectionPolicy
}

// Rerank scores candidates, applies selection policy, and returns
// filtered set sorted by score descending.
func (r *Reranker) Rerank(ctx context.Context, cs CandidateSet) ([]Candidate, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errors.New("rerank: nil reranker")
	}
	if r.Scorer == nil {
		return nil, errors.New("rerank: nil scorer")
	}

	scored, err := r.Scorer.Score(ctx, cs.Query, cs.Candidates)
	if err != nil {
		return nil, err
	}

	filtered := make([]Candidate, 0, len(scored))
	for _, candidate := range scored {
		if r.Policy.MinScore > 0 && candidate.Score < r.Policy.MinScore {
			continue
		}
		filtered = append(filtered, candidate)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].Score > filtered[j].Score
	})

	if r.Policy.TopK > 0 && len(filtered) > r.Policy.TopK {
		filtered = filtered[:r.Policy.TopK]
	}

	return filtered, nil
}
