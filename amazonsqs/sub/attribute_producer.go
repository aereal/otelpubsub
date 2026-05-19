package sub

import (
	"iter"

	"go.opentelemetry.io/otel/attribute"
)

type SQSProcessSpanAttributeProducer interface {
	ProduceSQSProcessSpanAttributes(msg *Message) iter.Seq[attribute.KeyValue]
}
