package observe

import (
	"context"
	"sync"
	"time"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
)

// InMemoryTracer stores span records in memory.
type InMemoryTracer struct {
	spans []SpanRecord
	mu    sync.Mutex
}

// SpanRecord is one stored span.
type SpanRecord struct {
	Name       string
	Attributes map[string]any
	Error      error
	Start      time.Time
	End        time.Time
}

// InMemorySpan updates parent tracer state.
type InMemorySpan struct {
	tracer *InMemoryTracer
	index  int
}

// InMemoryEventLog stores events in memory.
type InMemoryEventLog struct {
	events []Event
	mu     sync.Mutex
}

// InMemoryMetrics stores timings, usage, and counters in memory.
type InMemoryMetrics struct {
	latencies   map[string][]time.Duration
	tokenUsages []core.TokenUsage
	counters    map[string]int
	mu          sync.Mutex
}

// InMemoryCostTracker stores usage records in memory.
type InMemoryCostTracker struct {
	usages []usageRecord
	mu     sync.Mutex
}

type usageRecord struct {
	Model string
	Usage core.TokenUsage
}

type spanContextKey struct{}

// StartSpan opens span and stores record.
func (t *InMemoryTracer) StartSpan(ctx context.Context, name string) (context.Context, Span) {
	if ctx == nil {
		ctx = context.Background()
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	record := SpanRecord{
		Name:       name,
		Attributes: make(map[string]any),
		Start:      time.Now(),
	}
	t.spans = append(t.spans, record)

	span := &InMemorySpan{tracer: t, index: len(t.spans) - 1}
	return context.WithValue(ctx, spanContextKey{}, span), span
}

// End closes span.
func (s *InMemorySpan) End() {
	s.tracer.mu.Lock()
	defer s.tracer.mu.Unlock()

	s.tracer.spans[s.index].End = time.Now()
}

// SetAttribute stores attr on span.
func (s *InMemorySpan) SetAttribute(key string, value any) {
	s.tracer.mu.Lock()
	defer s.tracer.mu.Unlock()

	s.tracer.spans[s.index].Attributes[key] = value
}

// RecordError stores error on span.
func (s *InMemorySpan) RecordError(err error) {
	s.tracer.mu.Lock()
	defer s.tracer.mu.Unlock()

	s.tracer.spans[s.index].Error = err
}

// Record stores event.
func (l *InMemoryEventLog) Record(event Event) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.events = append(l.events, cloneEvent(event))
}

// Events returns recorded events in order.
func (l *InMemoryEventLog) Events() []Event {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.events) == 0 {
		return []Event{}
	}

	events := make([]Event, len(l.events))
	for i, event := range l.events {
		events[i] = cloneEvent(event)
	}

	return events
}

// RecordLatency stores latency for layer.
func (m *InMemoryMetrics) RecordLatency(layer string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.latencies == nil {
		m.latencies = make(map[string][]time.Duration)
	}
	m.latencies[layer] = append(m.latencies[layer], duration)
}

// RecordTokens stores token usage.
func (m *InMemoryMetrics) RecordTokens(usage core.TokenUsage) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tokenUsages = append(m.tokenUsages, usage)
}

// IncrementCounter adds delta to counter.
func (m *InMemoryMetrics) IncrementCounter(name string, delta int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.counters == nil {
		m.counters = make(map[string]int)
	}
	m.counters[name] += delta
}

// RecordUsage stores usage for model.
func (c *InMemoryCostTracker) RecordUsage(model string, usage core.TokenUsage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.usages = append(c.usages, usageRecord{Model: model, Usage: usage})
}

// TotalCost returns total cost.
func (c *InMemoryCostTracker) TotalCost() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	return 0
}

func cloneEvent(event Event) Event {
	event.Data = cloneData(event.Data)
	return event
}

func cloneData(data map[string]any) map[string]any {
	if data == nil {
		return nil
	}

	cloned := make(map[string]any, len(data))
	for key, value := range data {
		cloned[key] = value
	}

	return cloned
}
