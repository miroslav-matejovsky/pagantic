package observe

import "context"

// TraceRecorder starts spans for work units.
type TraceRecorder interface {
	// StartSpan opens span. Returns ctx with span.
	StartSpan(ctx context.Context, name string) (context.Context, Span)
}

// Span records timing, attrs, and errors.
type Span interface {
	// End closes span.
	End()
	// SetAttribute stores attr on span.
	SetAttribute(key string, value any)
	// RecordError stores error on span.
	RecordError(err error)
}
