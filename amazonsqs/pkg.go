// Package amazonsqs provides OpenTelemetry instrumentation for AWS SQS.
//
// Use the pub subpackage to inject trace context when sending messages,
// and the sub subpackage to extract trace context when processing received messages.
package amazonsqs
