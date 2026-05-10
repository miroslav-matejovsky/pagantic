// Package observe holds layer 10 observability and control plane pieces.
//
// Layer 10 makes system measurable and debuggable. No execution without
// traceability.
//
// Key abstractions are TraceRecorder, EventLog, MetricsCollector, and
// CostTracker.
//
// # Correlation
//
// CorrelationContext ties events to requests, sessions, and execution spans.
// All events carry RequestID and SessionID. Causal context (step name, tool
// call id, or any other cause) is recorded in the CausedBy string field.
// SpanID and ParentSpanID link events into a span tree for timeline
// reconstruction.
//
// Package observe provides in-memory implementations for local runs and tests.
// It also provides no-op implementations for cases where observability is not
// needed.
package observe
