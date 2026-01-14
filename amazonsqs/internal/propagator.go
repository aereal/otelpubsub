package internal

import "go.opentelemetry.io/otel/propagation"

var Propagator = propagation.TraceContext{}
