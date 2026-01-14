package sub_test

import (
	"embed"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/aereal/otelpubsub/amazonsns/sub"
	"github.com/google/go-cmp/cmp"
)

//go:embed testdata/*
var testdata embed.FS

func TestEntity_marshal(t *testing.T) {
	t.Parallel()

	wantMap := map[string]any{
		"Signature": "EXAMPLE",
		"MessageId": "95df01b4-ee98-5cb9-9903-4c221d41eb5e",
		"Type":      "Notification",
		"TopicArn":  "arn:aws:sns:EXAMPLE",
		"MessageAttributes": map[string]any{
			"Test": map[string]string{
				"Type":  "String",
				"Value": "TestString",
			},
			"TestBinary": map[string]string{
				"Type":  "Binary",
				"Value": "AQIDBAU",
			},
		},
		"SignatureVersion": "1",
		"Timestamp":        "2015-06-03T17:43:27.123Z",
		"SigningCertUrl":   "EXAMPLE",
		"Message":          "Hello from SNS!",
		"UnsubscribeUrl":   "EXAMPLE",
		"Subject":          "TestInvoke",
	}
	want, err := json.Marshal(wantMap)
	if err != nil {
		t.Fatal(err)
	}

	entity := &sub.Entity{
		Timestamp: time.Date(2015, time.June, 3, 17, 43, 27, int(time.Duration(time.Millisecond*123).Nanoseconds()), time.UTC),
		MessageAttributes: sub.MessageAttributes{
			"Test":       sub.StringAttributeValue("TestString"),
			"TestBinary": sub.BinaryAttributeValue([]byte{1, 2, 3, 4, 5}),
		},
		Signature:        "EXAMPLE",
		MessageID:        "95df01b4-ee98-5cb9-9903-4c221d41eb5e",
		Type:             "Notification",
		TopicArn:         "arn:aws:sns:EXAMPLE",
		SignatureVersion: "1",
		SigningCertURL:   "EXAMPLE",
		UnsubscribeURL:   "EXAMPLE",
		Subject:          "TestInvoke",
		Message:          json.RawMessage(`"Hello from SNS!"`),
	}
	got, err := json.Marshal(entity)
	if err != nil {
		t.Fatal(err)
	}
	if err := diffJSONMessage(want, got); err != nil {
		t.Error(err)
	}
}

func TestEntity_unmarshal(t *testing.T) {
	t.Parallel()

	f, err := testdata.Open("testdata/entity.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })
	var entity sub.Entity
	if err := json.NewDecoder(f).Decode(&entity); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d message attributes", len(entity.MessageAttributes))
	for _, key := range slices.Sorted(maps.Keys(entity.MessageAttributes)) {
		value := entity.MessageAttributes[key]
		t.Logf("key=%s type=%s", key, value.Type())
	}
}

func TestMessageAttributes_carrier(t *testing.T) {
	t.Parallel()

	attrs := sub.MessageAttributes{
		"s": sub.StringAttributeValue("abc"),
		"n": sub.NumberAttributeValue("123"),
	}
	if got := attrs.Get("s"); got != "abc" {
		t.Errorf("got value: %s", got)
	}
	if got := attrs.Get("not-found"); got != "" {
		t.Errorf("got value: %s", got)
	}
	if got := attrs.Get("n"); got != "" {
		t.Errorf("got value: %s", got)
	}

	attrs.Set("s-2", "def")
	if got := attrs.Get("s-2"); got != "def" {
		t.Errorf("got value: %s", got)
	}

	keys := attrs.Keys()
	wantKeys := []string{"s", "s-2"}
	if diff := cmp.Diff(wantKeys, keys); diff != "" {
		t.Errorf("Keys() (-want, +got):\n%s", diff)
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
