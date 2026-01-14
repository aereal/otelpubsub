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
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
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
	gotSpans := exporter.GetSpans()
	t.Logf("%d spans got", len(gotSpans))
	for i, span := range gotSpans {
		t.Logf("span: index=%d name=%s", i, span.Name)
	}
	t.Logf("%d message attributes", len(gotMsgAttrs))
	wantMsgAttrs := map[string]messageAttributeValue{
		"traceparent": {DataType: utils.DataTypeString, Value: fmt.Sprintf("00-%s-%s-01", span.SpanContext().TraceID(), span.SpanContext().SpanID())},
	}
	if diff := cmp.Diff(wantMsgAttrs, gotMsgAttrs); diff != "" {
		t.Errorf("message attributes (-want, +got):\n%s", diff)
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
	gotSpans := exporter.GetSpans()
	t.Logf("%d spans got", len(gotSpans))
	for i, span := range gotSpans {
		t.Logf("span: index=%d name=%s", i, span.Name)
	}
	t.Logf("%d message attributes", len(gotMsgAttrs))
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
