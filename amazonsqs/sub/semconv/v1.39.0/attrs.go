package semconv

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
)

var AttrKeyAWSSQSMessageSentTimestamp = attribute.Key("aws.sqs.message.sent_timestamp")

func AttrAWSSQSMessageSentTimestamp(t time.Time) attribute.KeyValue {
	return AttrKeyAWSSQSMessageSentTimestamp.String(t.Format(time.RFC3339Nano))
}
