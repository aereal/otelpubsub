package sub_test

import (
	"testing"

	"github.com/aereal/otelpubsub/amazonsqs/sub"
	"github.com/google/go-cmp/cmp"
)

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
