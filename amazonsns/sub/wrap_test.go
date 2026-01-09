package sub_test

import (
	"context"
	"testing"

	"github.com/aereal/otelpubsub/amazonsns/sub"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestWrapProcessor(t *testing.T) {
	t.Parallel()

	testWrapper(t, func(opts ...sub.StartProcessSpanOption) func(context.Context, *sub.Entity) error {
		return sub.WrapProcessor(processorFunc, opts...)
	})
}

func TestWrapYielder(t *testing.T) {
	t.Parallel()

	testWrapper(t, func(opts ...sub.StartProcessSpanOption) func(context.Context, *sub.Entity) error {
		yielder := sub.WrapYielder(yielderFunc, opts...)
		return func(ctx context.Context, entity *sub.Entity) error {
			_, err := yielder(ctx, entity)
			return err
		}
	})
}

func testWrapper(t *testing.T, wrap func(o ...sub.StartProcessSpanOption) func(context.Context, *sub.Entity) error) {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	wantTraceIDHex := "abcdef121234567890abcdef12345678"
	wantSpanIDHex := "1234567890abcdef"
	entity := &sub.Entity{
		MessageAttributes: sub.MessageAttributes{
			"otel.trace_id": sub.StringAttributeValue(wantTraceIDHex),
			"otel.span_id":  sub.StringAttributeValue(wantSpanIDHex),
		},
	}
	if err := wrap(sub.WithTracerProvider(tp), sub.WithStartSpanOptions(trace.WithSpanKind(trace.SpanKindClient)))(t.Context(), entity); err != nil {
		t.Fatal(err)
	}
	if err := tp.ForceFlush(t.Context()); err != nil {
		t.Fatal(err)
	}
	gotSpans := exporter.GetSpans()
	t.Logf("%d spans got", len(gotSpans))
	var found bool
	for _, span := range gotSpans {
		for _, link := range span.Links {
			if link.SpanContext.TraceID().String() == wantTraceIDHex && link.SpanContext.SpanID().String() == wantSpanIDHex {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("no span link found: %#v", gotSpans)
	}
}

func processorFunc(ctx context.Context, entity *sub.Entity) error { return nil }

func yielderFunc(ctx context.Context, entity *sub.Entity) (bool, error) { return true, nil }
