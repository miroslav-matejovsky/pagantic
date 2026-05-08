package observe

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInMemoryTracer_StartAndEndSpan(t *testing.T) {
	tracer := &InMemoryTracer{}

	ctx, span := tracer.StartSpan(context.Background(), "work")
	require.NotNil(t, ctx)
	require.NotNil(t, span)

	span.End()

	require.Len(t, tracer.spans, 1)
	require.Equal(t, "work", tracer.spans[0].Name)
	require.False(t, tracer.spans[0].Start.IsZero())
	require.False(t, tracer.spans[0].End.IsZero())
}

func TestInMemoryTracer_SetAttributesOnSpan(t *testing.T) {
	tracer := &InMemoryTracer{}
	_, span := tracer.StartSpan(context.Background(), "work")

	span.SetAttribute("model", "gpt")
	span.SetAttribute("tokens", 42)

	require.Len(t, tracer.spans, 1)
	require.Equal(t, "gpt", tracer.spans[0].Attributes["model"])
	require.Equal(t, 42, tracer.spans[0].Attributes["tokens"])
}

func TestInMemoryTracer_RecordErrorOnSpan(t *testing.T) {
	tracer := &InMemoryTracer{}
	_, span := tracer.StartSpan(context.Background(), "work")
	wantErr := errors.New("boom")

	span.RecordError(wantErr)

	require.Len(t, tracer.spans, 1)
	require.Equal(t, wantErr, tracer.spans[0].Error)
}
