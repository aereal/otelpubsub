package sub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aereal/otelpubsub/amazonsns/sub"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

var (
	dummyTraceID = trace.TraceID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x21, 0xdc, 0x18, 0x7, 0x52, 0x47, 0x85}
	dummySpanID  = trace.SpanID{0x0, 0x21, 0x97, 0xec, 0x5d, 0x8a, 0x25, 0xe}
)

func TestWrapProcessor(t *testing.T) {
	t.Parallel()

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
				Name: "github.com/aereal/otelpubsub/amazonsns/sub",
			},
		},
	}
	testWrapper(t, func(opts ...sub.StartProcessSpanOption) func(context.Context, *sub.Entity) error {
		return sub.WrapProcessor(processorFunc, opts...)
	}, wantSpans)
}

func TestWrapYielder(t *testing.T) {
	t.Parallel()

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
				Name: "github.com/aereal/otelpubsub/amazonsns/sub",
			},
		},
	}
	testWrapper(t, func(opts ...sub.StartProcessSpanOption) func(context.Context, *sub.Entity) error {
		yielder := sub.WrapYielder(yielderFunc, opts...)
		return func(ctx context.Context, entity *sub.Entity) error {
			_, err := yielder(ctx, entity)
			return err
		}
	}, wantSpans)
}

func testWrapper(t *testing.T, wrap func(o ...sub.StartProcessSpanOption) func(context.Context, *sub.Entity) error, wantSpans []tracetest.SpanStub) {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	wantTraceIDHex := "abcdef121234567890abcdef12345678"
	wantSpanIDHex := "1234567890abcdef"
	entity := &sub.Entity{
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
	if diff := diffSpans(wantSpans, gotSpans); diff != "" {
		t.Errorf("spans (-want, +got):\n%s", diff)
	}
}

func processorFunc(ctx context.Context, entity *sub.Entity) error { return nil }

func yielderFunc(ctx context.Context, entity *sub.Entity) (bool, error) { return true, nil }

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
