// Package rerank handles layer 6 reranking and evaluation.
//
// Layer 6 checks candidate output after first pass. It scores retrieved or
// generated candidates, reorders them by semantic relevance, and picks best
// subset.
//
// Role: secondary reasoning layer correcting initial approximations.
//
// Key types:
//   - Candidate
//   - CandidateSet
//   - RelevanceScorer
//   - Reranker
//   - SelectionPolicy
//
// TODO:
//   - Cross-encoder reranker. Needs reranking model.
//   - LLM-based scoring. Needs inference integration.
package rerank
