package sub

import (
	"context"

	"github.com/aereal/otelpubsub/amazonsns"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Processor func(context.Context, *Entity) error

func WrapProcessor(f Processor, opts ...StartProcessSpanOption) Processor {
	return func(ctx context.Context, entity *Entity) (err error) {
		ctx, span := StartProcessSpan(ctx, entity, opts...)
		defer func() {
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "")
			}
			span.End()
		}()

		return f(ctx, entity)
	}
}

type Yielder[V any] func(context.Context, *Entity) (V, error)

func WrapYielder[V any](f Yielder[V], opts ...StartProcessSpanOption) Yielder[V] {
	return func(ctx context.Context, entity *Entity) (_ V, err error) {
		ctx, span := StartProcessSpan(ctx, entity, opts...)
		defer func() {
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "")
			}
			span.End()
		}()

		return f(ctx, entity)
	}
}

func StartProcessSpan(ctx context.Context, entity *Entity, opts ...StartProcessSpanOption) (context.Context, trace.Span) {
	var cfg config
	for _, o := range opts {
		o.applyStartProcessSpanOption(&cfg)
	}
	if cfg.tracerProvider == nil {
		cfg.tracerProvider = otel.GetTracerProvider()
	}
	if entity != nil {
		remoteCtx := amazonsns.Propagator{}.Extract(ctx, entity.MessageAttributes)
		link := trace.LinkFromContext(remoteCtx)
		if link.SpanContext.IsValid() {
			cfg.startSpanOptions = append(cfg.startSpanOptions, trace.WithLinks(link))
		}
	}
	return cfg.tracerProvider.Tracer("github.com/aereal/amazonsns/sub").Start(ctx, "process", cfg.startSpanOptions...)
}
