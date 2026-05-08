package context

import (
	"context"
	"sort"
	"strings"
)

// Document is stored document for in-memory retriever.
type Document struct {
	Content  string
	Source   string
	Metadata map[string]any
}

// InMemoryRetriever is simple keyword-matching retriever for tests.
type InMemoryRetriever struct {
	Documents []Document
}

// NewInMemoryRetriever builds in-memory retriever with docs.
func NewInMemoryRetriever(docs ...Document) *InMemoryRetriever {
	copied := make([]Document, len(docs))
	copy(copied, docs)

	return &InMemoryRetriever{Documents: copied}
}

// Retrieve finds matching documents by query words.
func (r *InMemoryRetriever) Retrieve(ctx context.Context, query string, limit int) ([]Chunk, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	words := strings.Fields(strings.ToLower(query))
	if len(words) == 0 || len(r.Documents) == 0 {
		return []Chunk{}, nil
	}

	results := make([]Chunk, 0, len(r.Documents))
	for _, doc := range r.Documents {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		content := strings.ToLower(doc.Content)
		matches := 0
		for _, word := range words {
			if strings.Contains(content, word) {
				matches++
			}
		}
		if matches == 0 {
			continue
		}

		results = append(results, Chunk{
			Content:  doc.Content,
			Source:   doc.Source,
			Score:    float64(matches) / float64(len(words)),
			Metadata: doc.Metadata,
		})
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}
