package memory

import "github.com/miroslav-matejovsky/pagantic/core"

// ConversationBuffer stores message history with optional size limit.
// It is primary state container for multi-turn conversations.
type ConversationBuffer struct {
	messages []core.Message
	maxSize  int // 0 means unlimited
}

// NewConversationBuffer creates a buffer. maxSize 0 means unlimited.
func NewConversationBuffer(maxSize int) *ConversationBuffer {
	if maxSize < 0 {
		maxSize = 0
	}

	return &ConversationBuffer{maxSize: maxSize}
}

// Append adds a message. If maxSize > 0 and buffer is full, oldest
// non-system message is dropped.
func (cb *ConversationBuffer) Append(msg core.Message) {
	cb.messages = append(cb.messages, msg)
	if cb.maxSize <= 0 {
		return
	}

	for len(cb.messages) > cb.maxSize {
		idx := cb.oldestNonSystemIndex()
		if idx < 0 {
			return
		}

		cb.messages = append(cb.messages[:idx], cb.messages[idx+1:]...)
	}
}

// Messages returns a copy of all messages.
func (cb *ConversationBuffer) Messages() []core.Message {
	msgs := make([]core.Message, len(cb.messages))
	copy(msgs, cb.messages)
	return msgs
}

// Len returns message count.
func (cb *ConversationBuffer) Len() int {
	return len(cb.messages)
}

// Clear removes all messages.
func (cb *ConversationBuffer) Clear() {
	cb.messages = nil
}

// oldestNonSystemIndex finds oldest message safe to drop.
func (cb *ConversationBuffer) oldestNonSystemIndex() int {
	for i, msg := range cb.messages {
		if msg.Role != core.RoleSystem {
			return i
		}
	}

	return -1
}
