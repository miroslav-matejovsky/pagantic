package observe

import (
	"context"
	"errors"
	"testing"
	"time"

	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	"github.com/stretchr/testify/require"
)

var (
	_ TraceRecorder    = NoOpTracer{}
	_ Span             = NoOpSpan{}
	_ EventLog         = NoOpEventLog{}
	_ MetricsCollector = NoOpMetrics{}
	_ CostTracker      = NoOpCostTracker{}
)

func TestNoOpImplementations_DoNotPanic(t *testing.T) {
	require.NotPanics(t, func() {
		ctx, span := NoOpTracer{}.StartSpan(context.Background(), "work")
		require.NotNil(t, ctx)
		require.NotNil(t, span)

		span.SetAttribute("key", "value")
		span.RecordError(errors.New("boom"))
		span.End()

		NoOpEventLog{}.Record(Event{Timestamp: time.Now(), Action: "run"})
		require.Nil(t, NoOpEventLog{}.Events())

		NoOpMetrics{}.RecordLatency("llm", time.Second)
		NoOpMetrics{}.RecordTokens(core.TokenUsage{PromptTokens: 1})
		NoOpMetrics{}.IncrementCounter("runs", 1)

		NoOpCostTracker{}.RecordUsage("model", core.TokenUsage{OutputTokens: 2})
		require.Zero(t, NoOpCostTracker{}.TotalCost())
	})
}
