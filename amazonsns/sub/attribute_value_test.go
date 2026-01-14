package sub_test

import (
	"testing"

	"github.com/aereal/otelpubsub/amazonsns/sub"
)

func TestAttributeValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		av                     sub.AttributeValue
		wantStringValue        *attributeValueExpectation
		wantStringArrayValue   *attributeValueExpectation
		wantNumberValue        *attributeValueExpectation
		wantEncodedBinaryValue *attributeValueExpectation
		name                   string
		wantAttrType           sub.AttributeType
	}{
		{name: "string", av: sub.StringAttributeValue("s"), wantAttrType: sub.AttributeTypeString, wantStringValue: someAttrValue("s")},
		{name: "string array", av: sub.StringArrayAttributeValue("s,t,u"), wantAttrType: sub.AttributeTypeStringArray, wantStringArrayValue: someAttrValue("s,t,u")},
		{name: "number", av: sub.NumberAttributeValue("123"), wantAttrType: sub.AttributeTypeNumber, wantNumberValue: someAttrValue("123")},
		{name: "binary", av: sub.BinaryAttributeValue([]byte{1, 2, 3, 4, 5}), wantAttrType: sub.AttributeTypeBinary, wantEncodedBinaryValue: someAttrValue("AQIDBAU")},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.av.Type() != tc.wantAttrType {
				t.Errorf("Type:\n\twant: %s\n\t got: %s", tc.wantAttrType, tc.av.Type())
			}

			t.Run("String", assertValueGetter(tc.av.StringValue, tc.wantStringValue))
			t.Run("StringArray", assertValueGetter(tc.av.StringArrayValue, tc.wantStringArrayValue))
			t.Run("Number", assertValueGetter(tc.av.NumberValue, tc.wantNumberValue))
			t.Run("Binary", assertValueGetter(tc.av.Base64EncodedBinaryValue, tc.wantEncodedBinaryValue))
		})
	}
}

func assertValueGetter(meth func() (string, bool), want *attributeValueExpectation) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		if want == nil {
			want = &attributeValueExpectation{}
		}
		got, gotOK := meth()
		if got != want.value {
			t.Errorf("value:\n\twant: %q\n\t got: %q", want.value, got)
		}
		if gotOK != want.ok {
			t.Errorf("ok: want=%v got=%v", want.ok, gotOK)
		}
	}
}

func someAttrValue(v string) *attributeValueExpectation {
	return &attributeValueExpectation{ok: true, value: v}
}

type attributeValueExpectation struct {
	value string
	ok    bool
}
