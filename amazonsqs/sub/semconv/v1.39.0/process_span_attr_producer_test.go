package semconv_test

import (
	"encoding/json"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/aereal/otelpubsub/amazonsqs/sub"
	semconv "github.com/aereal/otelpubsub/amazonsqs/sub/semconv/v1.39.0"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/google/go-cmp/cmp"
	"go.opentelemetry.io/otel/attribute"
)

func TestProcessSpanAttributeProducer(t *testing.T) {
	t.Parallel()

	ts := time.Date(2018, time.February, 3, 12, 34, 56, 789*1000*1000, time.UTC)
	queueARN := arn.ARN{
		Partition: "aws",
		Service:   "sqs",
		Region:    "ap-northeast-1",
		AccountID: "123456789012",
		Resource:  "queue-01",
	}.String()
	arnWithoutResource := arn.ARN{
		Partition: "aws",
		Service:   "sqs",
		Region:    "ap-northeast-1",
		AccountID: "123456789012",
	}.String()

	testCases := []struct {
		name string
		msg  *sub.Message
		want []attribute.KeyValue
	}{
		{
			name: "ok",
			msg: &sub.Message{
				MessageID:      "msg-001",
				Body:           json.RawMessage(`{"body":{"ok":true}}`),
				EventSourceARN: queueARN,
				Attributes: map[string]string{
					"SentTimestamp": strconv.FormatInt(ts.UnixMilli(), 10),
				},
			},
			want: []attribute.KeyValue{
				attribute.String("messaging.system", "aws_sqs"),
				attribute.String("messaging.operation.type", "process"),
				attribute.String("messaging.message.id", "msg-001"),
				attribute.Int("messaging.message.body.size", 20),
				attribute.String("aws.sqs.queue.url", "https://sqs.ap-northeast-1.amazonaws.com/123456789012/queue-01"),
				attribute.String("messaging.destination.name", "queue-01"),
				attribute.String("aws.sqs.message.sent_timestamp", ts.Format(time.RFC3339Nano)),
			},
		},
		{
			name: "no SentTimestamp",
			msg: &sub.Message{
				MessageID:      "msg-001",
				Body:           json.RawMessage(`{"body":{"ok":true}}`),
				EventSourceARN: queueARN,
				Attributes:     map[string]string{},
			},
			want: []attribute.KeyValue{
				attribute.String("messaging.system", "aws_sqs"),
				attribute.String("messaging.operation.type", "process"),
				attribute.String("messaging.message.id", "msg-001"),
				attribute.Int("messaging.message.body.size", 20),
				attribute.String("aws.sqs.queue.url", "https://sqs.ap-northeast-1.amazonaws.com/123456789012/queue-01"),
				attribute.String("messaging.destination.name", "queue-01"),
			},
		},
		{
			name: "no resource ARN",
			msg: &sub.Message{
				MessageID:      "msg-001",
				Body:           json.RawMessage(`{"body":{"ok":true}}`),
				EventSourceARN: arnWithoutResource,
				Attributes:     map[string]string{},
			},
			want: []attribute.KeyValue{
				attribute.String("messaging.system", "aws_sqs"),
				attribute.String("messaging.operation.type", "process"),
				attribute.String("messaging.message.id", "msg-001"),
				attribute.Int("messaging.message.body.size", 20),
			},
		},
		{
			name: "corrupted ARN",
			msg: &sub.Message{
				MessageID:      "msg-001",
				Body:           json.RawMessage(`{"body":{"ok":true}}`),
				EventSourceARN: "queue-01",
				Attributes:     map[string]string{},
			},
			want: []attribute.KeyValue{
				attribute.String("messaging.system", "aws_sqs"),
				attribute.String("messaging.operation.type", "process"),
				attribute.String("messaging.message.id", "msg-001"),
				attribute.Int("messaging.message.body.size", 20),
			},
		},
		{
			name: "nil message",
			msg:  nil,
			want: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			want := attribute.NewSet(tc.want...)
			got := attribute.NewSet(slices.Collect(semconv.ProcessSpanAttributeProducer{}.ProduceSQSProcessSpanAttributes(tc.msg))...)
			if diff := cmp.Diff(want, got, cmp.Comparer(func(a, b attribute.Set) bool { return a.Equals(&b) })); diff != "" {
				t.Errorf("attributes (-want, +got):\n%s", diff)
			}
		})
	}
}
