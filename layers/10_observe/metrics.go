package observe

import (
	"time"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
)

// MetricsCollector stores counters and timings.
type MetricsCollector interface {
	// RecordLatency stores latency for layer.
	RecordLatency(layer string, duration time.Duration)
	// RecordTokens stores token usage.
	RecordTokens(usage core.TokenUsage)
	// IncrementCounter adds delta to counter.
	IncrementCounter(name string, delta int)
}
