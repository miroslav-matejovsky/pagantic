package memory

import (
	"sort"
	"sync"
)

// SessionState stores key-value pairs across execution steps.
// Thread-safe via sync.RWMutex.
type SessionState struct {
	values map[string]any
	mu     sync.RWMutex
}

// NewSessionState creates empty session state.
func NewSessionState() *SessionState {
	return &SessionState{values: make(map[string]any)}
}

// Set stores value for key.
func (ss *SessionState) Set(key string, value any) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.values[key] = value
}

// Get loads value for key.
func (ss *SessionState) Get(key string) (any, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	value, ok := ss.values[key]
	return value, ok
}

// Delete removes key from state.
func (ss *SessionState) Delete(key string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	delete(ss.values, key)
}

// Keys returns sorted list of keys.
func (ss *SessionState) Keys() []string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	keys := make([]string, 0, len(ss.values))
	for key := range ss.values {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}
