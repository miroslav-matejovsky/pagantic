package observe

// CorrelationContext ties observability events to requests, sessions,
// and execution spans. All events should carry these identifiers.
type CorrelationContext struct {
	RequestID    string
	SessionID    string
	TraceID      string
	SpanID       string
	ParentSpanID string
	CausedBy     string // causal link: tool_call_id, step_name, etc.
}
