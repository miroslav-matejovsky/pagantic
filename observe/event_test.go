package observe

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInMemoryEventLog_RecordAndRetrieve(t *testing.T) {
	log := &InMemoryEventLog{}
	event := Event{
		Timestamp: time.Now(),
		Layer:     "llm",
		Action:    "chat",
		Data:      map[string]any{"step": 1},
		Duration:  time.Second,
	}

	log.Record(event)

	events := log.Events()
	require.Len(t, events, 1)
	require.Equal(t, event.Timestamp, events[0].Timestamp)
	require.Equal(t, event.Layer, events[0].Layer)
	require.Equal(t, event.Action, events[0].Action)
	require.Equal(t, event.Data, events[0].Data)
	require.Equal(t, event.Duration, events[0].Duration)
}

func TestInMemoryEventLog_EventsReturnedInOrder(t *testing.T) {
	log := &InMemoryEventLog{}
	first := Event{Timestamp: time.Now(), Action: "first"}
	second := Event{Timestamp: time.Now().Add(time.Second), Action: "second"}

	log.Record(first)
	log.Record(second)

	events := log.Events()
	require.Len(t, events, 2)
	require.Equal(t, "first", events[0].Action)
	require.Equal(t, "second", events[1].Action)
}

func TestInMemoryEventLog_EmptyLogReturnsEmptySlice(t *testing.T) {
	log := &InMemoryEventLog{}

	events := log.Events()
	require.NotNil(t, events)
	require.Len(t, events, 0)
}
