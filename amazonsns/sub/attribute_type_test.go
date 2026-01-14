package sub_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/aereal/otelpubsub/amazonsns/sub"
)

var testCaseAttributeTypeConversion = []struct {
	jv []byte
	at sub.AttributeType
}{
	{at: sub.AttributeTypeString, jv: []byte(`"String"`)},
	{at: sub.AttributeTypeStringArray, jv: []byte(`"String.Array"`)},
	{at: sub.AttributeTypeNumber, jv: []byte(`"Number"`)},
	{at: sub.AttributeTypeBinary, jv: []byte(`"Binary"`)},
}

func TestAttributeType_unmarshal(t *testing.T) {
	t.Parallel()

	for _, tc := range testCaseAttributeTypeConversion {
		t.Run(tc.at.String(), func(t *testing.T) {
			t.Parallel()

			var got sub.AttributeType
			if err := json.Unmarshal(tc.jv, &got); err != nil {
				t.Fatal(err)
			}
			if got != tc.at {
				t.Errorf("want=%v got=%v", tc.at, got)
			}
		})
	}
}

func TestAttributeType_marshal(t *testing.T) {
	t.Parallel()

	for _, tc := range testCaseAttributeTypeConversion {
		t.Run(tc.at.String(), func(t *testing.T) {
			t.Parallel()

			jv, err := json.Marshal(tc.at)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(jv, tc.jv) {
				t.Errorf("String() mismatch:\n\twant: %q\n\t got: %q", tc.jv, string(jv))
			}
		})
	}
}
