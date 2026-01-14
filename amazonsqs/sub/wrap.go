package sub

import (
	"context"

	"github.com/aereal/otelpubsub/amazonsqs/internal"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Processor is a function type that processes an SQS message.
type Processor func(context.Context, *Message) error

// WrapProcessor wraps a [Processor] to automatically start and end a span for each message processing.
// Errors returned from the wrapped function are recorded on the span.
func WrapProcessor(f Processor, opts ...StartProcessSpanOption) Processor {
	return func(ctx context.Context, msg *Message) (err error) {
		ctx, span := StartProcessSpan(ctx, msg, opts...)
		defer func() {
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "")
			}
			span.End()
		}()

		return f(ctx, msg)
	}
}

// Yielder is a generic function type that processes an SQS message and returns a value.
type Yielder[V any] func(context.Context, *Message) (V, error)

// WrapYielder wraps a [Yielder] to automatically start and end a span for each message processing.
// Errors returned from the wrapped function are recorded on the span.
func WrapYielder[V any](f Yielder[V], opts ...StartProcessSpanOption) Yielder[V] {
	return func(ctx context.Context, msg *Message) (_ V, err error) {
		ctx, span := StartProcessSpan(ctx, msg, opts...)
		defer func() {
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "")
			}
			span.End()
		}()

		return f(ctx, msg)
	}
}

// StartProcessSpan starts a new span for processing an SQS message.
// If the message contains trace context in its message attributes, the span is linked to the original trace.
// The caller is responsible for calling End on the returned span.
func StartProcessSpan(ctx context.Context, msg *Message, opts ...StartProcessSpanOption) (context.Context, trace.Span) {
	var cfg config
	for _, o := range opts {
		o.applyStartProcessSpanOption(&cfg)
	}
	if cfg.tracerProvider == nil {
		cfg.tracerProvider = otel.GetTracerProvider()
	}
	if msg != nil {
		remoteCtx := internal.Propagator.Extract(ctx, msg.MessageAttributes)
		link := trace.LinkFromContext(remoteCtx)
		if link.SpanContext.IsValid() {
			cfg.startSpanOptions = append(cfg.startSpanOptions, trace.WithLinks(link))
		}
	}
	return cfg.tracerProvider.Tracer("github.com/aereal/otelpubsub/amazonsqs/sub").Start(ctx, "process", cfg.startSpanOptions...)
}
