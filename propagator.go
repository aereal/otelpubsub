package otelpubsub

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	keyTraceID = "otel.trace_id"
	keySpanID  = "otel.span_id"
)

var (
	emptySpanContext trace.SpanContext
	fields           = []string{keyTraceID, keySpanID}
)

type Propagator struct{}

var _ propagation.TextMapPropagator = Propagator{}

func (Propagator) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return
	}
	carrier.Set(keyTraceID, sc.TraceID().String())
	carrier.Set(keySpanID, sc.SpanID().String())
}

func (Propagator) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	sc, err := extractSpanContext(carrier)
	if err != nil {
		slog.WarnContext(ctx, "failed to extract span context", slog.Any("error", err))
		return ctx
	}
	return trace.ContextWithRemoteSpanContext(ctx, sc)
}

func (Propagator) Fields() []string { return slices.Clone(fields) }

func extractSpanContext(carrier propagation.TextMapCarrier) (trace.SpanContext, error) {
	traceID, err := trace.TraceIDFromHex(carrier.Get(keyTraceID))
	if err != nil {
		return emptySpanContext, fmt.Errorf("trace.TraceIDFromHex: %w", err)
	}
	spanID, err := trace.SpanIDFromHex(carrier.Get(keySpanID))
	if err != nil {
		return emptySpanContext, fmt.Errorf("trace.SpanIDFromHex: %w", err)
	}
	return trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	}), nil
}
