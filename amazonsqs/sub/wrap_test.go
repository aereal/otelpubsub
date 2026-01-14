package sub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aereal/otelpubsub/amazonsqs/sub"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestWrapProcessor(t *testing.T) {
	t.Parallel()

	testWrapper(t, func(opts ...sub.StartProcessSpanOption) func(context.Context, *sub.Message) error {
		return sub.WrapProcessor(processorFunc, opts...)
	})
}

func TestWrapYielder(t *testing.T) {
	t.Parallel()

	testWrapper(t, func(opts ...sub.StartProcessSpanOption) func(context.Context, *sub.Message) error {
		yielder := sub.WrapYielder(yielderFunc, opts...)
		return func(ctx context.Context, entity *sub.Message) error {
			_, err := yielder(ctx, entity)
			return err
		}
	})
}

func testWrapper(t *testing.T, wrap func(o ...sub.StartProcessSpanOption) func(context.Context, *sub.Message) error) {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	wantTraceIDHex := "abcdef121234567890abcdef12345678"
	wantSpanIDHex := "1234567890abcdef"
	entity := &sub.Message{
		MessageAttributes: sub.MessageAttributes{
			"traceparent": sub.StringAttributeValue(fmt.Sprintf("00-%s-%s-01", wantTraceIDHex, wantSpanIDHex)),
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

func processorFunc(ctx context.Context, entity *sub.Message) error { return nil }

func yielderFunc(ctx context.Context, entity *sub.Message) (bool, error) { return true, nil }
