package context

import "context"

// Chunk is piece of retrieved content with relevance metadata.
type Chunk struct {
	Content  string
	Source   string
	Score    float64
	Metadata map[string]any
}

// Retriever finds relevant content for query.
type Retriever interface {
	Retrieve(ctx context.Context, query string, limit int) ([]Chunk, error)
}
