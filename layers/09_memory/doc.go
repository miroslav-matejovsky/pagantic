// Package memory manages explicit, versionable state across execution steps.
//
// Layer 9 keeps state outside prompt text. State stays explicit so other
// layers can inspect it, reset it, persist it, and version it.
//
// Key types:
//   - ConversationBuffer stores message history for multi-turn work.
//   - SessionState stores thread-safe key-value data across steps.
//   - WorkingMemory stores transient per-step context and results.
//   - MemoryPolicy governs eviction, persistence, and ephemeral context handling.
//
// Each orchestration pattern uses memory differently. AgentLoop persists
// ConversationMemory and SessionState across turns but treats context as
// ephemeral WorkingMemory. SpecializedLoop is fully stateless. MemoryPolicy
// formalizes these rules so behavior is explicit, not a special case.
package memory
