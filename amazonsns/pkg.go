// Package amazonsns provides OpenTelemetry instrumentation for AWS SNS.
//
// Use the pub subpackage to inject trace context when publishing messages,
// and the sub subpackage to extract trace context when processing received messages.
package amazonsns
