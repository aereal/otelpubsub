package semconv

import (
	"iter"
	"log/slog"

	"github.com/aereal/otelpubsub/amazonsns/sub"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
)

type ProcessSpanAttributeProducer struct{}

var _ sub.SNSProcessSpanAttributeProducer = ProcessSpanAttributeProducer{}

func (p ProcessSpanAttributeProducer) ProduceSNSProcessSpanAttributes(entity *sub.Entity) iter.Seq[attribute.KeyValue] {
	return func(yield func(attribute.KeyValue) bool) {
		if entity == nil {
			return
		}

		if !yield(semconv.MessagingSystemAWSSNS) {
			return
		}
		if !yield(semconv.MessagingOperationTypeProcess) {
			return
		}
		if !yield(semconv.MessagingMessageID(entity.MessageID)) {
			return
		}
		if !yield(AttrAWSSNSMessageTimestamp(entity.Timestamp)) {
			return
		}
		if !yield(semconv.MessagingMessageBodySize(len(entity.Message))) {
			return
		}

		for kv := range p.topicARNAttrs(arn.Parse(entity.TopicArn)) {
			if !yield(kv) {
				return
			}
		}
	}
}

func (ProcessSpanAttributeProducer) topicARNAttrs(topicARN arn.ARN, err error) iter.Seq[attribute.KeyValue] {
	return func(yield func(attribute.KeyValue) bool) {
		if err != nil {
			slog.Warn("failed to parse Amazon SNS topic ARN", slog.String("error", err.Error()))
			return
		}
		if topicARN.Resource == "" {
			slog.Warn("parsed ARN's resource part is empty")
			return
		}
		if !yield(semconv.AWSSNSTopicARN(topicARN.String())) {
			return
		}
		if !yield(semconv.MessagingDestinationName(topicARN.Resource)) {
			return
		}
	}
}
