package sub

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var _ = otel.GetTracerProvider

type config struct {
	tracerProvider   trace.TracerProvider
	startSpanOptions []trace.SpanStartOption
}

// StartProcessSpanOption configures [StartProcessSpan] behavior.
type StartProcessSpanOption interface {
	applyStartProcessSpanOption(*config)
}

// WithTracerProvider specifies the [trace.TracerProvider] to use for creating spans.
// If not specified, [otel.GetTracerProvider] is used.
func WithTracerProvider(tp trace.TracerProvider) StartProcessSpanOption {
	return &optionWithTracerProvider{tp: tp}
}

type optionWithTracerProvider struct{ tp trace.TracerProvider }

func (o *optionWithTracerProvider) applyStartProcessSpanOption(c *config) { c.tracerProvider = o.tp }

// WithStartSpanOptions appends additional [trace.SpanStartOption] to the span creation.
func WithStartSpanOptions(opts ...trace.SpanStartOption) StartProcessSpanOption {
	return &optionWithStartSpanOptions{opts: opts}
}

type optionWithStartSpanOptions struct{ opts []trace.SpanStartOption }

func (o *optionWithStartSpanOptions) applyStartProcessSpanOption(c *config) {
	c.startSpanOptions = append(c.startSpanOptions, o.opts...)
}
