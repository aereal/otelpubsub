package sub

import (
	"encoding/base64"
	"encoding/json"
)

func StringAttributeValue(v string) AttributeValue {
	return newAttrValue(AttributeTypeString, v)
}

func BinaryAttributeValue(raw []byte) AttributeValue {
	dst := make([]byte, base64.RawStdEncoding.EncodedLen(len(raw)))
	base64.RawStdEncoding.Encode(dst, raw)
	return newAttrValue(AttributeTypeBinary, string(dst))
}

func NumberAttributeValue(v string) AttributeValue {
	return newAttrValue(AttributeTypeNumber, v)
}

func StringArrayAttributeValue(v string) AttributeValue {
	return newAttrValue(AttributeTypeStringArray, v)
}

func newAttrValue(t AttributeType, v string) AttributeValue {
	return &attributeValue{payload: &attributeValuePayload{Value: v, Type: t}}
}

type AttributeValue interface {
	json.Marshaler
	json.Unmarshaler

	Type() AttributeType
	StringValue() (string, bool)
	StringArrayValue() (string, bool)
	NumberValue() (string, bool)
	Base64EncodedBinaryValue() (string, bool)
}

type attributeValuePayload struct {
	Value string
	Type  AttributeType
}

type attributeValue struct {
	payload *attributeValuePayload
}

var _ AttributeValue = (*attributeValue)(nil)

func (av *attributeValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(av.payload)
}

func (av *attributeValue) UnmarshalJSON(b []byte) error {
	var payload attributeValuePayload
	if err := json.Unmarshal(b, &payload); err != nil {
		return err
	}
	if av == nil {
		av = &attributeValue{}
	}
	av.payload = &payload
	return nil
}

func (av *attributeValue) Type() AttributeType { return av.payload.Type }

func (av *attributeValue) StringValue() (string, bool) {
	if av.payload.Type != AttributeTypeString {
		return "", false
	}
	return av.payload.Value, true
}

func (av *attributeValue) StringArrayValue() (string, bool) {
	if av.payload.Type != AttributeTypeStringArray {
		return "", false
	}
	return av.payload.Value, true
}

func (av *attributeValue) NumberValue() (string, bool) {
	if av.payload.Type != AttributeTypeNumber {
		return "", false
	}
	return av.payload.Value, true
}

func (av *attributeValue) Base64EncodedBinaryValue() (string, bool) {
	if av.payload.Type != AttributeTypeBinary {
		return "", false
	}
	return av.payload.Value, true
}
