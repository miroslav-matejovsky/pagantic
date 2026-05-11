// CLI example: inference with full observability wiring.
//
// Demonstrates layer 10 (observe) with all four observer types:
//
//   - TraceRecorder (InMemoryTracer): records span lifecycle - start, set
//     attributes, record errors, end. Spans are work units with timing.
//
//   - EventLog (InMemoryEventLog): records discrete events with timestamp,
//     layer, action, data, duration, and error. AgentLoop records events
//     through LoopConfig.Observer field.
//
//   - MetricsCollector (InMemoryMetrics): records latencies per layer,
//     token usage, and named counters.
//
//   - CostTracker (InMemoryCostTracker): records model usage for cost
//     calculation. TotalCost() currently returns 0 (pricing not wired).
//
//   - NoOp implementations: NoOpTracer, NoOpSpan, NoOpEventLog, NoOpMetrics,
//     NoOpCostTracker are do-nothing defaults. Used when observability is
//     disabled or in tests.
//
// Wiring:
//   - Only EventLog is wired through LoopConfig.Observer (orchestrate layer)
//   - TraceRecorder, MetricsCollector, CostTracker are used directly in this
//     example around the inference call
//   - In production, these would be wired into middleware or interceptors
//
// Key pagantic concept: observability is a cross-cutting concern. The observe
// layer provides interfaces and implementations; wiring is up to the adapter.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/miroslav-matejovsky/pagantic/adapters/cli"
	"github.com/miroslav-matejovsky/pagantic/kronk"
	core "github.com/miroslav-matejovsky/pagantic/layers/00_core"
	orchestrate "github.com/miroslav-matejovsky/pagantic/layers/02_orchestrate"
	observe "github.com/miroslav-matejovsky/pagantic/layers/10_observe"
)

const llmModel = "unsloth/gemma-4-E4B-it"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	prompt, err := cli.ReadPrompt(os.Args[1:], os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create all observer types.
	tracer := &observe.InMemoryTracer{}
	eventLog := &observe.InMemoryEventLog{}
	metrics := &observe.InMemoryMetrics{}
	costTracker := &observe.InMemoryCostTracker{}

	// NoOp implementations exist as safe defaults.
	_ = observe.NoOpTracer{}
	_ = observe.NoOpSpan{}
	_ = observe.NoOpEventLog{}
	_ = observe.NoOpMetrics{}
	_ = observe.NoOpCostTracker{}

	// Demonstrate NoOp: calling methods is safe and does nothing.
	noopTracer := observe.NoOpTracer{}
	noopCtx, noopSpan := noopTracer.StartSpan(ctx, "noop-test")
	noopSpan.SetAttribute("key", "value")
	noopSpan.RecordError(fmt.Errorf("test error"))
	noopSpan.End()
	_ = noopCtx

	noopEvents := observe.NoOpEventLog{}
	noopEvents.Record(observe.Event{Layer: "test", Action: "noop"})
	_ = noopEvents.Events() // returns nil

	noopMetrics := observe.NoOpMetrics{}
	noopMetrics.RecordLatency("test", time.Second)
	noopMetrics.RecordTokens(core.TokenUsage{})
	noopMetrics.IncrementCounter("test", 1)

	noopCost := observe.NoOpCostTracker{}
	noopCost.RecordUsage("model", core.TokenUsage{})
	_ = noopCost.TotalCost() // returns 0

	// Start tracing span for the entire query.
	ctx, span := tracer.StartSpan(ctx, "observe-query")
	span.SetAttribute("prompt", prompt)

	krn, cleanup, err := kronk.Load(ctx, kronk.Config{ModelSource: llmModel})
	if err != nil {
		span.RecordError(err)
		span.End()
		fmt.Fprintf(os.Stderr, "Failed to load engine: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	engine := kronk.NewAdapter(krn, nil)

	// Record inference latency.
	inferStart := time.Now()

	// EventLog wired through LoopConfig.Observer.
	agent := orchestrate.NewAgentLoop(orchestrate.LoopConfig{
		SystemPrompt: "You are a helpful assistant. Be concise.",
		Engine:       engine,
		Observer:     eventLog,
	})

	callCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	result, err := agent.Chat(callCtx, prompt)
	inferDuration := time.Since(inferStart)

	if err != nil {
		span.RecordError(err)
		span.End()
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Record metrics.
	metrics.RecordLatency("inference", inferDuration)
	metrics.RecordTokens(result.Usage)
	metrics.IncrementCounter("inference_calls", 1)

	// Record cost.
	costTracker.RecordUsage(llmModel, result.Usage)

	span.SetAttribute("tokens", result.Usage.PromptTokens+result.Usage.OutputTokens)
	span.End()

	// Print result.
	fmt.Println(result.Content)

	// Print observability data.
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "--- Observability ---\n")
	fmt.Fprintf(os.Stderr, "Inference latency: %v\n", inferDuration)
	fmt.Fprintf(os.Stderr, "Prompt tokens:     %d\n", result.Usage.PromptTokens)
	fmt.Fprintf(os.Stderr, "Output tokens:     %d\n", result.Usage.OutputTokens)
	fmt.Fprintf(os.Stderr, "Total cost:        $%.4f\n", costTracker.TotalCost())

	events := eventLog.Events()
	fmt.Fprintf(os.Stderr, "\nEvents (%d):\n", len(events))
	for _, ev := range events {
		fmt.Fprintf(os.Stderr, "  [%s] %s.%s\n", ev.Timestamp.Format("15:04:05.000"), ev.Layer, ev.Action)
	}
}

// Usage:
//
//	go run examples/cli/observe-query/main.go "What is the capital of France?"
//	echo "Explain Go interfaces" | go run examples/cli/observe-query/main.go
