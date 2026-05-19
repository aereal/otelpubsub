package sub

import (
	"iter"

	"go.opentelemetry.io/otel/attribute"
)

type SNSProcessSpanAttributeProducer interface {
	ProduceSNSProcessSpanAttributes(entity *Entity) iter.Seq[attribute.KeyValue]
}
