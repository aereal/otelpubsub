package semconv_test

import (
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/aereal/otelpubsub/amazonsns/sub"
	semconv "github.com/aereal/otelpubsub/amazonsns/sub/semconv/v1.39.0"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/google/go-cmp/cmp"
	"go.opentelemetry.io/otel/attribute"
)

func TestProcessSpanAttributeProducer(t *testing.T) {
	t.Parallel()

	ts := time.Date(2018, time.February, 3, 12, 34, 56, 789*1000*1000, time.UTC)
	topicARN := arn.ARN{
		Service:   "sns",
		Region:    "ap-northeast-1",
		AccountID: "123456789012",
		Resource:  "topic-01",
	}.String()
	arnWithoutResource := arn.ARN{
		Service:   "sns",
		Region:    "ap-northeast-1",
		AccountID: "123456789012",
	}.String()

	testCases := []struct {
		name   string
		entity *sub.Entity
		want   []attribute.KeyValue
	}{
		{
			name: "ok",
			entity: &sub.Entity{
				Timestamp: ts,
				MessageID: "msg-001",
				Message:   json.RawMessage(`{"body":{"ok":true}}`),
				TopicArn:  topicARN,
			},
			want: []attribute.KeyValue{
				attribute.String("messaging.system", "aws.sns"),
				attribute.String("messaging.operation.type", "process"),
				attribute.String("aws.sns.topic.arn", "arn::sns:ap-northeast-1:123456789012:topic-01"),
				attribute.String("messaging.message.id", "msg-001"),
				attribute.String("aws.sns.message.timestamp", "2018-02-03T12:34:56.789Z"),
				attribute.Int("messaging.message.body.size", 20),
				attribute.String("messaging.destination.name", "topic-01"),
			},
		},
		{
			name: "no resource ARN",
			entity: &sub.Entity{
				Timestamp: ts,
				MessageID: "msg-001",
				Message:   json.RawMessage(`{"body":{"ok":true}}`),
				TopicArn:  arnWithoutResource,
			},
			want: []attribute.KeyValue{
				attribute.String("messaging.system", "aws.sns"),
				attribute.String("messaging.operation.type", "process"),
				attribute.String("messaging.message.id", "msg-001"),
				attribute.String("aws.sns.message.timestamp", "2018-02-03T12:34:56.789Z"),
				attribute.Int("messaging.message.body.size", 20),
			},
		},
		{
			name: "corrupted ARN",
			entity: &sub.Entity{
				Timestamp: ts,
				MessageID: "msg-001",
				Message:   json.RawMessage(`{"body":{"ok":true}}`),
				TopicArn:  "topic-01",
			},
			want: []attribute.KeyValue{
				attribute.String("messaging.system", "aws.sns"),
				attribute.String("messaging.operation.type", "process"),
				attribute.String("messaging.message.id", "msg-001"),
				attribute.String("aws.sns.message.timestamp", "2018-02-03T12:34:56.789Z"),
				attribute.Int("messaging.message.body.size", 20),
			},
		},
		{
			name:   "nil entity",
			entity: nil,
			want:   nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			want := attribute.NewSet(tc.want...)
			got := attribute.NewSet(slices.Collect(semconv.ProcessSpanAttributeProducer{}.ProduceSNSProcessSpanAttributes(tc.entity))...)
			if diff := cmp.Diff(want, got, cmp.Comparer(func(a, b attribute.Set) bool { return a.Equals(&b) })); diff != "" {
				t.Errorf("attributes (-want, +got):\n%s", diff)
			}
		})
	}
}
