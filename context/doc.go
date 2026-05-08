// Package context holds layer 3 context and knowledge pieces.
//
// Layer 3 gives model relevant, bounded knowledge.
// Input is user query.
// Output is structured context block.
//
// Invariant: model must never act on unconstrained or unverified knowledge.
//
// Key types are Retriever, Chunk, ContextBuilder, and Document.
//
// TODO:
//   - EmbeddingService: needs embedding model or API.
//   - VectorIndex: needs vector database.
package context
