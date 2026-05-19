package sub_test

import (
	"context"
	"fmt"
	"iter"
	"testing"

	"github.com/aereal/otelpubsub/amazonsqs/sub"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

type producerFunc func(msg *sub.Message) iter.Seq[attribute.KeyValue]

func (f producerFunc) ProduceSQSProcessSpanAttributes(msg *sub.Message) iter.Seq[attribute.KeyValue] {
	return f(msg)
}

var (
	dummyTraceID = trace.TraceID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x21, 0xdc, 0x18, 0x7, 0x52, 0x47, 0x85}
	dummySpanID  = trace.SpanID{0x0, 0x21, 0x97, 0xec, 0x5d, 0x8a, 0x25, 0xe}
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
	wantSpans := []tracetest.SpanStub{
		{
			Name:        "process",
			SpanContext: trace.NewSpanContext(trace.SpanContextConfig{TraceFlags: trace.FlagsSampled, TraceID: dummyTraceID, SpanID: dummySpanID}),
			SpanKind:    trace.SpanKindClient,
			Links: []sdktrace.Link{
				{
					SpanContext: trace.NewSpanContext(trace.SpanContextConfig{Remote: true, TraceFlags: trace.FlagsSampled, TraceID: dummyTraceID, SpanID: dummySpanID}),
				},
			},
			Resource: resource.NewSchemaless(
				attribute.String("service.name", "unknown_service:sub.test"),
				attribute.String("telemetry.sdk.language", "go"),
				attribute.String("telemetry.sdk.name", "opentelemetry"),
				attribute.String("telemetry.sdk.version", "1.43.0"),
			),
			InstrumentationScope: instrumentation.Scope{
				Name: "github.com/aereal/otelpubsub/amazonsqs/sub",
			},
		},
	}
	if diff := diffSpans(wantSpans, exporter.GetSpans()); diff != "" {
		t.Errorf("spans (-want, +got):\n%s", diff)
	}
}

func diffSpans(want, got []tracetest.SpanStub) string {
	return cmp.Diff(want, got,
		cmpopts.IgnoreFields(
			tracetest.SpanStub{},
			"StartTime", "EndTime",
			"InstrumentationLibrary",
		),
		cmp.Comparer(func(a, b attribute.Set) bool {
			return a.Equals(&b)
		}),
		cmp.Comparer(func(a, b *resource.Resource) bool {
			return a.Equal(b)
		}),
		cmp.Comparer(func(a, b trace.SpanContext) bool {
			return a.HasTraceID() == b.HasTraceID() &&
				a.HasSpanID() == b.HasSpanID() &&
				a.IsSampled() == b.IsSampled() &&
				a.IsRemote() == b.IsRemote() &&
				a.IsValid() == b.IsValid()
		}),
	)
}

func TestStartProcessSpan_WithAttributeProducers(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	msg := &sub.Message{}
	producer := producerFunc(func(_ *sub.Message) iter.Seq[attribute.KeyValue] {
		return func(yield func(attribute.KeyValue) bool) {
			yield(attribute.String("test.key", "test.value"))
		}
	})

	_, span := sub.StartProcessSpan(t.Context(), msg,
		sub.WithTracerProvider(tp),
		sub.WithAttributeProducers(producer),
	)
	span.End()

	if err := tp.ForceFlush(t.Context()); err != nil {
		t.Fatal(err)
	}

	gotSpans := exporter.GetSpans()
	if len(gotSpans) != 1 {
		t.Fatalf("got %d spans, want 1", len(gotSpans))
	}
	wantAttrs := attribute.NewSet(attribute.String("test.key", "test.value"))
	gotAttrs := attribute.NewSet(gotSpans[0].Attributes...)
	if diff := cmp.Diff(wantAttrs, gotAttrs, cmp.Comparer(func(a, b attribute.Set) bool { return a.Equals(&b) })); diff != "" {
		t.Errorf("attributes (-want, +got):\n%s", diff)
	}
}

func processorFunc(ctx context.Context, entity *sub.Message) error { return nil }

func yielderFunc(ctx context.Context, entity *sub.Message) (bool, error) { return true, nil }
