package pub_test

import (
	"testing"

	"github.com/aereal/otelpubsub/amazonsns/internal/utils"
	"github.com/aereal/otelpubsub/amazonsns/pub"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/google/go-cmp/cmp"
)

func TestMessageAttributesCarrier_Get(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		attrs map[string]types.MessageAttributeValue
		name  string
		want  string
	}{
		{
			name:  "string value",
			want:  "v",
			attrs: map[string]types.MessageAttributeValue{"k": utils.StringAttributeValue("v")},
		},
		{
			name:  "number value",
			want:  "",
			attrs: map[string]types.MessageAttributeValue{"k": utils.NumberAttributeValue("123")},
		},
		{
			name:  "DataType is a String but unexpectedly value is nil",
			want:  "",
			attrs: map[string]types.MessageAttributeValue{"k": {DataType: utils.Ptr(utils.DataTypeString), StringValue: nil}},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			carrier := pub.NewMessageAttributeCarrier(tc.attrs)
			got := carrier.Get("k")
			if got != tc.want {
				t.Errorf("want=%q got=%q", tc.want, got)
			}
		})
	}
}

func TestMessageAttributesCarrier_Keys(t *testing.T) {
	t.Parallel()

	carrier := pub.NewMessageAttributeCarrier(map[string]types.MessageAttributeValue{
		"s2": utils.StringAttributeValue("v2"),
		"n1": utils.NumberAttributeValue("1"),
		"n2": utils.NumberAttributeValue("2"),
		"s1": utils.StringAttributeValue("v1"),
	})
	want := []string{
		"s1",
		"s2",
	}
	if diff := cmp.Diff(want, carrier.Keys()); diff != "" {
		t.Errorf("(-want, +got):\n%s", diff)
	}
}
