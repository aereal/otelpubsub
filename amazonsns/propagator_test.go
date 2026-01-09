package amazonsns_test

import (
	"context"
	"testing"

	"github.com/aereal/otelpubsub/amazonsns"
	"github.com/google/go-cmp/cmp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	hexTraceID = "abcdef121234567890abcdef12345678"
	hexSpanID  = "1234567890abcdef"

	traceID trace.TraceID
	spanID  trace.SpanID
)

func init() {
	var err error
	traceID, err = trace.TraceIDFromHex(hexTraceID)
	if err != nil {
		panic(err)
	}
	spanID, err = trace.SpanIDFromHex(hexSpanID)
	if err != nil {
		panic(err)
	}
}

func TestPropagator_Fields(t *testing.T) {
	t.Parallel()

	got := amazonsns.Propagator{}.Fields()
	want := []string{
		"otel.trace_id",
		"otel.span_id",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want, +got):\n%s", diff)
	}
}

func TestPropagator_Extract(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		carrier         propagation.MapCarrier
		wantSpanContext trace.SpanContext
	}{
		{
			name:    "ok",
			carrier: propagation.MapCarrier{"otel.trace_id": hexTraceID, "otel.span_id": hexSpanID},
			wantSpanContext: trace.NewSpanContext(trace.SpanContextConfig{
				TraceID: traceID,
				SpanID:  spanID,
				Remote:  true,
			}),
		},
		{
			name:            "no span bound",
			carrier:         propagation.MapCarrier{},
			wantSpanContext: trace.SpanContext{},
		},
		{
			name:            "no valid span ID",
			carrier:         propagation.MapCarrier{"otel.trace_id": hexTraceID},
			wantSpanContext: trace.SpanContext{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotSpanContext := trace.SpanContextFromContext(amazonsns.Propagator{}.Extract(context.Background(), tc.carrier))
			if !tc.wantSpanContext.Equal(gotSpanContext) {
				t.Errorf("want: %#v; got: %#v", tc.wantSpanContext, gotSpanContext)
			}
		})
	}
}

func TestPropagator_Inject(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		ctx  context.Context
		want propagation.MapCarrier
		name string
	}{
		{
			name: "ok",
			ctx: newContext(trace.SpanContextConfig{
				TraceID: traceID,
				SpanID:  spanID,
			}),
			want: propagation.MapCarrier{
				"otel.trace_id": hexTraceID,
				"otel.span_id":  hexSpanID,
			},
		},
		{
			name: "empty context",
			ctx:  context.Background(),
			want: propagation.MapCarrier{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := propagation.MapCarrier{}
			amazonsns.Propagator{}.Inject(tc.ctx, m)
			if diff := cmp.Diff(tc.want, m); diff != "" {
				t.Errorf("carrier (-want, +got):\n%s", diff)
			}
		})
	}
}

func newContext(cfg trace.SpanContextConfig) context.Context {
	return trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(cfg))
}
