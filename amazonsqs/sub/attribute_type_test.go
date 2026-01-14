package sub_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aereal/otelpubsub/amazonsqs/sub"
	"github.com/google/go-cmp/cmp"
)

var (
	testCaseAttributeTypeConversion = []struct {
		attrType sub.AttributeType
		label    string
		json     []byte
		custom   bool
	}{
		{
			json:     []byte(`"String"`),
			attrType: sub.AttributeTypeString,
		},
		{
			json:     []byte(`"Number"`),
			attrType: sub.AttributeTypeNumber,
		},
		{
			json:     []byte(`"Binary"`),
			attrType: sub.AttributeTypeBinary,
		},
		{
			json:     []byte(`"Binary.png"`),
			attrType: sub.CustomType(sub.AttributeKindBinary, "png"),
			custom:   true,
			label:    "png",
		},
	}
)

func TestAttributeType_marshal(t *testing.T) {
	t.Parallel()

	for _, tc := range testCaseAttributeTypeConversion {
		t.Run(fmt.Sprintf("%#v", tc.attrType), func(t *testing.T) {
			t.Parallel()

			gotJSON, err := json.Marshal(tc.attrType)
			if err != nil {
				t.Fatal(err)
			}
			var got any
			if err := json.Unmarshal(gotJSON, &got); err != nil {
				t.Fatal(err)
			}
			var want any
			if err := json.Unmarshal(tc.json, &want); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("(-want, +got):\n%s", diff)
			}
		})
	}
}

func TestAttributeType_unmarshal(t *testing.T) {
	t.Parallel()

	for _, tc := range testCaseAttributeTypeConversion {
		t.Run(string(tc.json), func(t *testing.T) {
			t.Parallel()

			var got sub.AttributeType
			if err := json.Unmarshal(tc.json, &got); err != nil {
				t.Fatal(err)
			}
			if got != tc.attrType {
				t.Errorf("want=%#v got=%#v", tc.attrType, got)
			}
			if label := got.Label(); label != tc.label {
				t.Errorf("Label(): want=%q got=%q", tc.label, label)
			}
			if isCustom := got.IsCustom(); isCustom != tc.custom {
				t.Errorf("IsCustom(): want=%v got=%v", tc.custom, isCustom)
			}
		})
	}
}
