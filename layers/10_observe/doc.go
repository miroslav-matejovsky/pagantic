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
// All events should carry request_id, session_id (when in AgentLoop),
// step_name (when in PlanExecutor), and tool_call_id (when tool-related).
// This enables full request timeline reconstruction and causality tracing.
//
// Package observe provides in-memory implementations for local runs and tests.
// It also provides no-op implementations for cases where observability is not
// needed.
package observe
