package semconv

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
)

var AttrKeyAWSSNSMessageTimestamp = attribute.Key("aws.sns.message.timestamp")

func AttrAWSSNSMessageTimestamp(t time.Time) attribute.KeyValue {
	return AttrKeyAWSSNSMessageTimestamp.String(t.Format(time.RFC3339Nano))
}
