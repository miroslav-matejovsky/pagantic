package context

import (
	"context"
	"fmt"
	"strings"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
)

// ContextBuilder assembles retrieved chunks into messages for model.
type ContextBuilder struct {
	Retriever Retriever
	MaxChunks int // max chunks to include; 0 or negative means use all retrieved
}

// Build retrieves relevant content and assembles it into context messages.
func (cb *ContextBuilder) Build(ctx context.Context, query string) ([]core.Message, error) {
	if cb == nil || cb.Retriever == nil {
		return nil, fmt.Errorf("context: nil retriever")
	}

	limit := cb.MaxChunks
	if limit < 0 {
		limit = 0
	}

	chunks, err := cb.Retriever.Retrieve(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	if len(chunks) == 0 {
		return []core.Message{}, nil
	}

	var builder strings.Builder
	builder.WriteString("Relevant context:\n\n")
	for i, chunk := range chunks {
		fmt.Fprintf(&builder, "[%d] (%s): %s", i+1, chunk.Source, chunk.Content)
		if i < len(chunks)-1 {
			builder.WriteString("\n")
		}
	}

	return []core.Message{core.NewSystemMessage(builder.String())}, nil
}
