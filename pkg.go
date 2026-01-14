// Package otelpubsub provides OpenTelemetry trace context propagation for AWS SNS and SQS messages.
//
// This package defines [Propagator], a [propagation.TextMapPropagator] implementation
// that injects and extracts trace context using message attributes.
package otelpubsub

import "go.opentelemetry.io/otel/propagation"

var _ propagation.TextMapPropagator
