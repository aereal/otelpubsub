package semconv

import (
	"fmt"
	"iter"
	"log/slog"
	"strconv"
	"time"

	"github.com/aereal/otelpubsub/amazonsqs/sub"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
)

type ProcessSpanAttributeProducer struct{}

var _ sub.SQSProcessSpanAttributeProducer = ProcessSpanAttributeProducer{}

func (p ProcessSpanAttributeProducer) ProduceSQSProcessSpanAttributes(msg *sub.Message) iter.Seq[attribute.KeyValue] {
	return func(yield func(attribute.KeyValue) bool) {
		if msg == nil {
			return
		}

		if !yield(semconv.MessagingSystemAWSSQS) {
			return
		}
		if !yield(semconv.MessagingOperationTypeProcess) {
			return
		}
		if !yield(semconv.MessagingMessageID(msg.MessageID)) {
			return
		}
		if !yield(semconv.MessagingMessageBodySize(len(msg.Body))) {
			return
		}

		for kv := range p.queueARNAttrs(arn.Parse(msg.EventSourceARN)) {
			if !yield(kv) {
				return
			}
		}

		if ts, ok := p.parseSentTimestamp(msg.Attributes); ok {
			if !yield(AttrAWSSQSMessageSentTimestamp(ts)) {
				return
			}
		}
	}
}

func (ProcessSpanAttributeProducer) queueARNAttrs(queueARN arn.ARN, err error) iter.Seq[attribute.KeyValue] {
	return func(yield func(attribute.KeyValue) bool) {
		if err != nil {
			slog.Warn("failed to parse Amazon SQS queue ARN", slog.String("error", err.Error()))
			return
		}
		if queueARN.Resource == "" {
			slog.Warn("parsed ARN's resource part is empty")
			return
		}
		queueURL := fmt.Sprintf("https://sqs.%s.amazonaws.com/%s/%s", queueARN.Region, queueARN.AccountID, queueARN.Resource)
		if !yield(semconv.AWSSQSQueueURL(queueURL)) {
			return
		}
		if !yield(semconv.MessagingDestinationName(queueARN.Resource)) {
			return
		}
	}
}

func (ProcessSpanAttributeProducer) parseSentTimestamp(attrs map[string]string) (time.Time, bool) {
	raw, ok := attrs["SentTimestamp"]
	if !ok {
		return time.Time{}, false
	}
	ms, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		slog.Warn("failed to parse SentTimestamp", slog.String("value", raw), slog.String("error", err.Error()))
		return time.Time{}, false
	}
	return time.UnixMilli(ms).UTC(), true
}
