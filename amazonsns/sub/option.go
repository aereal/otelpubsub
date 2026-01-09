package sub

import "go.opentelemetry.io/otel/trace"

type config struct {
	tracerProvider   trace.TracerProvider
	startSpanOptions []trace.SpanStartOption
}

type StartProcessSpanOption interface {
	applyStartProcessSpanOption(*config)
}

func WithTracerProvider(tp trace.TracerProvider) StartProcessSpanOption {
	return &optionWithTracerProvider{tp: tp}
}

type optionWithTracerProvider struct{ tp trace.TracerProvider }

func (o *optionWithTracerProvider) applyStartProcessSpanOption(c *config) { c.tracerProvider = o.tp }

func WithStartSpanOptions(opts ...trace.SpanStartOption) StartProcessSpanOption {
	return &optionWithStartSpanOptions{opts: opts}
}

type optionWithStartSpanOptions struct{ opts []trace.SpanStartOption }

func (o *optionWithStartSpanOptions) applyStartProcessSpanOption(c *config) {
	c.startSpanOptions = append(c.startSpanOptions, o.opts...)
}
