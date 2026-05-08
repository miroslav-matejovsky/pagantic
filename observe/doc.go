// Package observe holds layer 10 observability and control plane pieces.
//
// Layer 10 makes system measurable and debuggable. No execution without
// traceability.
//
// Key abstractions are TraceRecorder, EventLog, MetricsCollector, and
// CostTracker.
//
// Package observe provides in-memory implementations for local runs and tests.
// It also provides no-op implementations for cases where observability is not
// needed.
package observe
