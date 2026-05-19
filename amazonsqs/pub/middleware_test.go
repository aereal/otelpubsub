package pub_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aereal/otelpubsub/amazonsqs/internal/utils"
	"github.com/aereal/otelpubsub/amazonsqs/pub"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
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

type messageAttributeValue struct {
	DataType string
	Value    string
}

func TestMiddleware_sendMessage(t *testing.T) {
	t.Parallel()

	var gotMsgAttrs map[string]messageAttributeValue
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		input := new(sqs.SendMessageInput)
		if err := json.NewDecoder(r.Body).Decode(input); err != nil {
			t.Errorf("failed to decode request body: %s", err)
			return
		}
		gotMsgAttrs = map[string]messageAttributeValue{}
		for k, v := range input.MessageAttributes {
			gotMsgAttrs[k] = messageAttributeValue{
				DataType: *v.DataType,
				Value:    *v.StringValue,
			}
		}
	}))
	t.Cleanup(srv.Close)
	cfg := aws.Config{
		Region:       "us-east-1",
		Credentials:  staticCredentials("id", "secret", "token"),
		BaseEndpoint: &srv.URL,
	}
	pub.AppendMiddlewares(&cfg.APIOptions)
	client := sqs.NewFromConfig(cfg)

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))

	input := &sqs.SendMessageInput{
		QueueUrl:    utils.Ptr("arn:aws:sqs:us-east-1:1234567890123:queue-1"),
		MessageBody: utils.Ptr("ptr"),
	}

	ctx, span := tp.Tracer("test").Start(t.Context(), "parent")
	if _, err := client.SendMessage(ctx, input); err != nil {
		t.Fatal(err)
	}
	span.End()
	if err := tp.ForceFlush(t.Context()); err != nil {
		t.Fatal(err)
	}
	wantMsgAttrs := map[string]messageAttributeValue{
		"traceparent": {DataType: utils.DataTypeString, Value: fmt.Sprintf("00-%s-%s-01", span.SpanContext().TraceID(), span.SpanContext().SpanID())},
	}
	if diff := cmp.Diff(wantMsgAttrs, gotMsgAttrs); diff != "" {
		t.Errorf("message attributes (-want, +got):\n%s", diff)
	}
	wantSpans := []tracetest.SpanStub{
		{
			Name:        "parent",
			SpanContext: trace.NewSpanContext(trace.SpanContextConfig{TraceFlags: trace.FlagsSampled, TraceID: dummyTraceID, SpanID: dummySpanID}),
			SpanKind:    trace.SpanKindInternal,
			Resource: resource.NewSchemaless(
				attribute.String("service.name", "unknown_service:pub.test"),
				attribute.String("telemetry.sdk.language", "go"),
				attribute.String("telemetry.sdk.name", "opentelemetry"),
				attribute.String("telemetry.sdk.version", "1.43.0"),
			),
			InstrumentationScope: instrumentation.Scope{
				Name: "test",
			},
		},
	}
	if diff := diffSpans(wantSpans, exporter.GetSpans()); diff != "" {
		t.Errorf("spans (-want, +got):\n%s", diff)
	}
}

func TestMiddleware_sendMessageBatch(t *testing.T) {
	t.Parallel()

	gotMsgAttrs := map[string]map[string]messageAttributeValue{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		input := new(sqs.SendMessageBatchInput)
		if err := json.NewDecoder(r.Body).Decode(input); err != nil {
			t.Errorf("failed to decode request body: %s", err)
			return
		}
		gotMsgAttrs = map[string]map[string]messageAttributeValue{}
		for _, entry := range input.Entries {
			entryAttrs := map[string]messageAttributeValue{}
			for k, v := range entry.MessageAttributes {
				entryAttrs[k] = messageAttributeValue{
					DataType: *v.DataType,
					Value:    *v.StringValue,
				}
			}
			gotMsgAttrs[*entry.Id] = entryAttrs
		}
	}))
	t.Cleanup(srv.Close)
	cfg := aws.Config{
		Region:       "us-east-1",
		Credentials:  staticCredentials("id", "secret", "token"),
		BaseEndpoint: &srv.URL,
	}
	pub.AppendMiddlewares(&cfg.APIOptions)
	client := sqs.NewFromConfig(cfg)

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))

	input := &sqs.SendMessageBatchInput{
		QueueUrl: utils.Ptr("arn:aws:sqs:us-east-1:1234567890123:queue-1"),
		Entries: []types.SendMessageBatchRequestEntry{
			{
				Id:          utils.Ptr("1"),
				MessageBody: utils.Ptr("msg-1"),
			},
			{
				Id:          utils.Ptr("2"),
				MessageBody: utils.Ptr("msg-2"),
			},
		},
	}
	ctx, span := tp.Tracer("test").Start(t.Context(), "parent")
	if _, err := client.SendMessageBatch(ctx, input); err != nil {
		t.Fatal(err)
	}
	span.End()
	if err := tp.ForceFlush(t.Context()); err != nil {
		t.Fatal(err)
	}
	wantMsgAttrs := map[string]map[string]messageAttributeValue{
		"1": {
			"traceparent": {DataType: utils.DataTypeString, Value: fmt.Sprintf("00-%s-%s-01", span.SpanContext().TraceID(), span.SpanContext().SpanID())},
		},
		"2": {
			"traceparent": {DataType: utils.DataTypeString, Value: fmt.Sprintf("00-%s-%s-01", span.SpanContext().TraceID(), span.SpanContext().SpanID())},
		},
	}
	if diff := cmp.Diff(wantMsgAttrs, gotMsgAttrs); diff != "" {
		t.Errorf("message attributes (-want, +got):\n%s", diff)
	}
	wantSpans := []tracetest.SpanStub{
		{
			Name:        "parent",
			SpanContext: trace.NewSpanContext(trace.SpanContextConfig{TraceFlags: trace.FlagsSampled, TraceID: dummyTraceID, SpanID: dummySpanID}),
			SpanKind:    trace.SpanKindInternal,
			Resource: resource.NewSchemaless(
				attribute.String("service.name", "unknown_service:pub.test"),
				attribute.String("telemetry.sdk.language", "go"),
				attribute.String("telemetry.sdk.name", "opentelemetry"),
				attribute.String("telemetry.sdk.version", "1.43.0"),
			),
			InstrumentationScope: instrumentation.Scope{
				Name: "test",
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

func staticCredentials(keyID, secret, sessionToken string) *awsCredentials {
	return &awsCredentials{Credentials: aws.Credentials{
		AccessKeyID:     keyID,
		SecretAccessKey: secret,
		SessionToken:    sessionToken,
	}}
}

type awsCredentials struct {
	aws.Credentials
}

var _ aws.CredentialsProvider = (*awsCredentials)(nil)

func (c *awsCredentials) Retrieve(_ context.Context) (aws.Credentials, error) {
	return c.Credentials, nil
}
