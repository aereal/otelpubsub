package sub_test

import (
	"embed"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/aereal/otelpubsub/amazonsqs/sub"
	"github.com/google/go-cmp/cmp"
)

//go:embed testdata/*
var testdata embed.FS

func TestMessage_marshal(t *testing.T) {
	t.Parallel()

	wantMap := map[string]any{
		"messageId":              "MessageID_1",
		"receiptHandle":          "MessageReceiptHandle",
		"body":                   "Message Body",
		"md5OfBody":              "fce0ea8dd236ccb3ed9b37dae260836f",
		"md5OfMessageAttributes": "582c92c5c5b6ac403040a4f3ab3115c9",
		"eventSourceARN":         "arn:aws:sqs:us-west-2:123456789012:SQSQueue",
		"eventSource":            "aws:sqs",
		"awsRegion":              "us-west-2",
		"attributes": map[string]string{
			"ApproximateReceiveCount":          "2",
			"SentTimestamp":                    "1520621625029",
			"SenderId":                         "AROAIWPX5BD2BHG722MW4:sender",
			"ApproximateFirstReceiveTimestamp": "1520621634884",
		},
		"messageAttributes": sub.MessageAttributes{
			"Attribute3": sub.BinaryAttributeValue([]byte{1, 1, 0, 0}),
			"Attribute2": sub.NumberAttributeValue("123"),
			"Attribute1": sub.StringAttributeValue("AttributeValue1"),
		},
	}
	want, err := json.Marshal(wantMap)
	if err != nil {
		t.Fatal(err)
	}

	msg := &sub.Message{
		MessageAttributes: sub.MessageAttributes{
			"Attribute1": sub.StringAttributeValue("AttributeValue1"),
			"Attribute2": sub.NumberAttributeValue("123"),
			"Attribute3": sub.BinaryAttributeValue([]byte{1, 1, 0, 0}),
		},
		Attributes: map[string]string{
			"ApproximateFirstReceiveTimestamp": "1520621634884",
			"ApproximateReceiveCount":          "2",
			"SenderId":                         "AROAIWPX5BD2BHG722MW4:sender",
			"SentTimestamp":                    "1520621625029",
		},
		MessageID:              "MessageID_1",
		MD5OfBody:              "fce0ea8dd236ccb3ed9b37dae260836f",
		MD5OfMessageAttributes: "582c92c5c5b6ac403040a4f3ab3115c9",
		ReceiptHandle:          "MessageReceiptHandle",
		EventSourceARN:         "arn:aws:sqs:us-west-2:123456789012:SQSQueue",
		EventSource:            "aws:sqs",
		AWSRegion:              "us-west-2",
		Body:                   json.RawMessage(`"Message Body"`),
	}
	got, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	if err := diffJSONMessage(want, got); err != nil {
		t.Error(err)
	}
}

func TestMessage_unmarshal(t *testing.T) {
	t.Parallel()

	f, err := testdata.Open("testdata/event.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })
	var ev sub.Event
	if err := json.NewDecoder(f).Decode(&ev); err != nil {
		t.Fatal(err)
	}
	msg := ev.Records[0]
	t.Logf("%d message attributes", len(msg.MessageAttributes))
	for _, key := range slices.Sorted(maps.Keys(msg.MessageAttributes)) {
		value := msg.MessageAttributes[key]
		t.Logf("key=%s type=%s", key, value.Type())
	}
}

var transformJSONMessage = cmp.Transformer("JSON", func(v json.RawMessage) map[string]any {
	var m map[string]any
	_ = json.Unmarshal(v, &m) //nolint:errcheck
	return m
})

func diffJSONMessage(want, got json.RawMessage) error {
	if diff := cmp.Diff(want, got, transformJSONMessage); diff != "" {
		return fmt.Errorf("(-want, +got):\n%s", diff) //nolint:err113
	}
	return nil
}
