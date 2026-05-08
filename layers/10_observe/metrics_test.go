package observe

import (
	"testing"
	"time"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	"github.com/stretchr/testify/require"
)

func TestInMemoryMetrics_RecordLatency(t *testing.T) {
	metrics := &InMemoryMetrics{}

	metrics.RecordLatency("llm", 25*time.Millisecond)

	require.Contains(t, metrics.latencies, "llm")
	require.Equal(t, []time.Duration{25 * time.Millisecond}, metrics.latencies["llm"])
}

func TestInMemoryMetrics_RecordTokens(t *testing.T) {
	metrics := &InMemoryMetrics{}
	usage := core.TokenUsage{PromptTokens: 10, OutputTokens: 5}

	metrics.RecordTokens(usage)

	require.Equal(t, []core.TokenUsage{usage}, metrics.tokenUsages)
}

func TestInMemoryMetrics_IncrementCounter(t *testing.T) {
	metrics := &InMemoryMetrics{}

	metrics.IncrementCounter("runs", 2)
	metrics.IncrementCounter("runs", 3)

	require.Equal(t, 5, metrics.counters["runs"])
}
