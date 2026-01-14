package sub

import (
	"context"

	"github.com/aereal/otelpubsub"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Processor func(context.Context, *Message) error

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

type Yielder[V any] func(context.Context, *Message) (V, error)

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

func StartProcessSpan(ctx context.Context, msg *Message, opts ...StartProcessSpanOption) (context.Context, trace.Span) {
	var cfg config
	for _, o := range opts {
		o.applyStartProcessSpanOption(&cfg)
	}
	if cfg.tracerProvider == nil {
		cfg.tracerProvider = otel.GetTracerProvider()
	}
	if msg != nil {
		remoteCtx := otelpubsub.Propagator{}.Extract(ctx, msg.MessageAttributes)
		link := trace.LinkFromContext(remoteCtx)
		if link.SpanContext.IsValid() {
			cfg.startSpanOptions = append(cfg.startSpanOptions, trace.WithLinks(link))
		}
	}
	return cfg.tracerProvider.Tracer("github.com/aereal/amazonsqs/sub").Start(ctx, "process", cfg.startSpanOptions...)
}
