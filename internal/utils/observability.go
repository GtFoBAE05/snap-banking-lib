package utils

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func StartSpan(ctx context.Context, tracer trace.Tracer, bank, operation string) (context.Context, trace.Span) {
	return tracer.Start(ctx, "snap."+operation,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("snap.bank", bank),
			attribute.String("snap.operation", operation),
		),
	)
}
