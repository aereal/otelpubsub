package sub

import (
	"encoding/json"
	"sort"
	"time"

	"go.opentelemetry.io/otel/propagation"
)

// Entity represents an SNS notification message delivered via HTTP/S endpoints.
// This structure matches the JSON format documented at:
// https://docs.aws.amazon.com/sns/latest/dg/sns-message-and-json-formats.html#http-notification-json
type Entity struct {
	Timestamp         time.Time         `json:"Timestamp"`
	MessageAttributes MessageAttributes `json:"MessageAttributes"`
	Signature         string            `json:"Signature"`
	MessageID         string            `json:"MessageId"`
	Type              string            `json:"Type"`
	TopicArn          string            `json:"TopicArn"`
	SignatureVersion  string            `json:"SignatureVersion"`
	SigningCertURL    string            `json:"SigningCertUrl"`
	UnsubscribeURL    string            `json:"UnsubscribeUrl"`
	Subject           string            `json:"Subject"`
	Message           json.RawMessage   `json:"Message"`
}

// MessageAttributes is a map of attribute names to values, implementing [propagation.TextMapCarrier].
type MessageAttributes map[string]AttributeValue

var (
	_ propagation.TextMapCarrier = MessageAttributes{}
	_ json.Unmarshaler           = (*MessageAttributes)(nil)
)

func (ma *MessageAttributes) UnmarshalJSON(b []byte) error {
	concrete := map[string]*attributeValue{}
	if err := json.Unmarshal(b, &concrete); err != nil {
		return err
	}
	if ma == nil || *ma == nil {
		*ma = MessageAttributes{}
	}
	for k, v := range concrete {
		(*ma)[k] = v
	}
	return nil
}

func (ma MessageAttributes) Get(key string) string {
	av, ok := ma[key]
	if !ok {
		return ""
	}
	sv, ok := av.StringValue()
	if !ok {
		return ""
	}
	return sv
}

func (ma MessageAttributes) Set(key, value string) {
	ma[key] = StringAttributeValue(value)
}

func (ma MessageAttributes) Keys() []string {
	ret := make([]string, 0, len(ma))
	for k, v := range ma {
		if v.Type() != AttributeTypeString {
			continue
		}
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return ret
}
