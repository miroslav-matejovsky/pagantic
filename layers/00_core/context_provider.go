package core

import "context"

// ContextProvider retrieves context messages for a given query.
// Implementations assemble relevant domain knowledge (e.g. from a vector
// store or in-memory index) into messages that are injected ephemerally
// before each inference call without being stored in the conversation buffer.
//
// This interface lives in core so both the orchestrate layer (consumer) and
// the context layer (implementor) can reference it without creating a direct
// dependency between those layers. Implementations satisfy this interface via
// Go structural typing — no import of this package is required.
type ContextProvider interface {
	Build(ctx context.Context, query string) ([]Message, error)
}
