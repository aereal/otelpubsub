package sub

import (
	"encoding/json"
	"sort"

	"go.opentelemetry.io/otel/propagation"
)

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
		if !v.Type().IsString() {
			continue
		}
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return ret
}
