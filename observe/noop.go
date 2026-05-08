package observe

import (
	"context"
	"time"

	"github.com/miroslav-matejovsky/pagantic/core"
)

// NoOpTracer does nothing.
type NoOpTracer struct{}

// NoOpSpan does nothing.
type NoOpSpan struct{}

// NoOpEventLog does nothing.
type NoOpEventLog struct{}

// NoOpMetrics does nothing.
type NoOpMetrics struct{}

// NoOpCostTracker does nothing.
type NoOpCostTracker struct{}

// StartSpan opens no-op span.
func (NoOpTracer) StartSpan(ctx context.Context, _ string) (context.Context, Span) {
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx, NoOpSpan{}
}

// End does nothing.
func (NoOpSpan) End() {}

// SetAttribute does nothing.
func (NoOpSpan) SetAttribute(_ string, _ any) {}

// RecordError does nothing.
func (NoOpSpan) RecordError(_ error) {}

// Record does nothing.
func (NoOpEventLog) Record(_ Event) {}

// Events returns nil.
func (NoOpEventLog) Events() []Event { return nil }

// RecordLatency does nothing.
func (NoOpMetrics) RecordLatency(_ string, _ time.Duration) {}

// RecordTokens does nothing.
func (NoOpMetrics) RecordTokens(_ core.TokenUsage) {}

// IncrementCounter does nothing.
func (NoOpMetrics) IncrementCounter(_ string, _ int) {}

// RecordUsage does nothing.
func (NoOpCostTracker) RecordUsage(_ string, _ core.TokenUsage) {}

// TotalCost returns zero.
func (NoOpCostTracker) TotalCost() float64 { return 0 }
