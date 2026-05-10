package memory

// MemoryPolicy governs eviction, persistence, and ephemeral context
// handling for each loop pattern.
type MemoryPolicy struct {
	MaxMessages         int
	EvictionStrategy    string // "fifo", "lru", or "none"
	PersistConversation bool
	PersistSession      bool
	EphemeralContext    bool // true means context is working memory, not persisted
}
