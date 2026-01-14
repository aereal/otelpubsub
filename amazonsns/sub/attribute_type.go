package sub

import (
	"encoding/json"
	"fmt"
	"maps"

	"github.com/aereal/iter/seq2"
)

var (
	at2str = map[AttributeType]string{
		AttributeTypeString:      "String",
		AttributeTypeStringArray: "String.Array",
		AttributeTypeNumber:      "Number",
		AttributeTypeBinary:      "Binary",
	}
	str2at = maps.Collect(seq2.Flip(maps.All(at2str)))
)

// Attribute type constants corresponding to SNS message attribute data types.
// See: https://docs.aws.amazon.com/sns/latest/dg/sns-message-attributes.html
const (
	AttributeTypeString AttributeType = iota
	AttributeTypeStringArray
	AttributeTypeNumber
	AttributeTypeBinary
)

// AttributeType represents the data type of an SNS message attribute.
type AttributeType int

var (
	_ fmt.Stringer     = AttributeType(0)
	_ json.Marshaler   = AttributeType(0)
	_ json.Unmarshaler = (*AttributeType)(nil)
)

func (t AttributeType) String() string {
	s, ok := at2str[t]
	if !ok {
		return fmt.Sprintf("INVALID.AttributeType(%d)", t)
	}
	return s
}

func (t AttributeType) MarshalJSON() ([]byte, error) {
	s, ok := at2str[t]
	if !ok {
		return nil, ErrInvalidAttributeType
	}
	return json.Marshal(s)
}

func (t *AttributeType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	at, ok := str2at[s]
	if !ok {
		return &UnknownAttributeTypeError{AttributeType: s}
	}
	*t = at
	return nil
}
