package memory

import (
	"testing"

	"github.com/miroslav-matejovsky/pagantic/core"
	"github.com/stretchr/testify/require"
)

func TestConversationBufferAppendAndRetrieve(t *testing.T) {
	t.Parallel()

	cb := NewConversationBuffer(0)
	want := []core.Message{
		core.NewSystemMessage("sys"),
		core.NewUserMessage("user"),
		core.NewAssistantMessage("assistant"),
	}

	for _, msg := range want {
		cb.Append(msg)
	}

	require.Equal(t, want, cb.Messages())
}

func TestConversationBufferUnlimited(t *testing.T) {
	t.Parallel()

	cb := NewConversationBuffer(0)
	for i := range 10 {
		cb.Append(core.NewUserMessage(string(rune('a' + i))))
	}

	require.Len(t, cb.Messages(), 10)
}

func TestConversationBufferLimitedDropsOldestNonSystem(t *testing.T) {
	t.Parallel()

	cb := NewConversationBuffer(3)
	cb.Append(core.NewSystemMessage("sys"))
	cb.Append(core.NewUserMessage("one"))
	cb.Append(core.NewAssistantMessage("two"))
	cb.Append(core.NewUserMessage("three"))

	require.Equal(t, []core.Message{
		core.NewSystemMessage("sys"),
		core.NewAssistantMessage("two"),
		core.NewUserMessage("three"),
	}, cb.Messages())
}

func TestConversationBufferSystemMessagesNeverDropped(t *testing.T) {
	t.Parallel()

	cb := NewConversationBuffer(1)
	cb.Append(core.NewSystemMessage("sys-one"))
	cb.Append(core.NewSystemMessage("sys-two"))

	require.Equal(t, []core.Message{
		core.NewSystemMessage("sys-one"),
		core.NewSystemMessage("sys-two"),
	}, cb.Messages())
}

func TestConversationBufferClear(t *testing.T) {
	t.Parallel()

	cb := NewConversationBuffer(0)
	cb.Append(core.NewUserMessage("user"))
	cb.Append(core.NewAssistantMessage("assistant"))

	cb.Clear()

	require.Empty(t, cb.Messages())
}

func TestConversationBufferLen(t *testing.T) {
	t.Parallel()

	cb := NewConversationBuffer(0)
	require.Zero(t, cb.Len())

	cb.Append(core.NewUserMessage("one"))
	cb.Append(core.NewAssistantMessage("two"))

	require.Equal(t, 2, cb.Len())
}

func TestConversationBufferMessagesReturnsCopy(t *testing.T) {
	t.Parallel()

	cb := NewConversationBuffer(0)
	cb.Append(core.NewUserMessage("user"))

	msgs := cb.Messages()
	msgs[0].Content = "changed"
	msgs = append(msgs, core.NewAssistantMessage("extra"))

	stored := cb.Messages()
	require.Len(t, stored, 1)
	require.Equal(t, "user", stored[0].Content)
}
