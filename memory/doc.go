// Package memory manages explicit, versionable state across execution steps.
//
// Layer 9 keeps state outside prompt text. State stays explicit so other
// layers can inspect it, reset it, persist it, and version it.
//
// Key types:
//   - ConversationBuffer stores message history for multi-turn work.
//   - SessionState stores thread-safe key-value data across steps.
//   - WorkingMemory stores transient per-step context and results.
package memory
