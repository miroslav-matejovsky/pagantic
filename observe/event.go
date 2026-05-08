package observe

import "time"

// Event is one observable action.
type Event struct {
	Timestamp time.Time
	Layer     string
	Action    string
	Data      map[string]any
	Duration  time.Duration
	Error     error
}

// EventLog stores events.
type EventLog interface {
	// Record stores event.
	Record(event Event)
	// Events returns recorded events.
	Events() []Event
}
