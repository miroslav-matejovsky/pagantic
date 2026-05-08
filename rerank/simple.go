package rerank

import (
	"context"
	"strings"
	"unicode"
)

// SimpleScorer scores candidates by keyword overlap with query.
// For testing and development only.
type SimpleScorer struct{}

// Score assigns overlap scores to candidates.
func (s *SimpleScorer) Score(ctx context.Context, query string, candidates []Candidate) ([]Candidate, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return []Candidate{}, nil
	}

	queryWords := splitWords(query)
	scored := make([]Candidate, len(candidates))
	for i, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		candidate.Score = overlapScore(queryWords, candidate.Content)
		scored[i] = candidate
	}

	return scored, nil
}

// splitWords breaks text into lowercase word tokens.
func splitWords(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

// overlapScore computes keyword overlap for one candidate.
func overlapScore(queryWords []string, content string) float64 {
	if len(queryWords) == 0 {
		return 0
	}

	contentWords := splitWords(content)
	if len(contentWords) == 0 {
		return 0
	}

	wordSet := make(map[string]struct{}, len(contentWords))
	for _, word := range contentWords {
		wordSet[word] = struct{}{}
	}

	matches := 0
	for _, word := range queryWords {
		if _, ok := wordSet[word]; ok {
			matches++
		}
	}

	return float64(matches) / float64(len(queryWords))
}
