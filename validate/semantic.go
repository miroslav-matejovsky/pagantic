package validate

import "context"

// SemanticValidator uses LLM checks for issues like hallucination or
// incoherent output. Only interface lives here for now.
//
// TODO: implementation needs orchestration layer for LLM calls.
type SemanticValidator interface {
	// Validate says if output fits intent.
	Validate(ctx context.Context, output string, intent string) (valid bool, reason string, err error)
}
