package sub

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"maps"
	"strings"

	"github.com/aereal/iter/seq2"
)

// CustomType creates an [AttributeType] with a custom label (e.g., "String.MyCustomType").
// SQS supports custom types for String and Binary kinds.
// See: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-message-metadata.html
func CustomType(kind AttributeKind, label string) AttributeType {
	return AttributeType{kind: kind, label: label}
}

// Predefined attribute types corresponding to SQS message attribute data types.
var (
	AttributeTypeString = AttributeType{kind: AttributeKindString}
	AttributeTypeNumber = AttributeType{kind: AttributeKindNumber}
	AttributeTypeBinary = AttributeType{kind: AttributeKindBinary}
)

// AttributeType represents the data type of an SQS message attribute.
// Unlike SNS, SQS allows custom type labels (e.g., "String.UUID" or "Binary.png").
type AttributeType struct {
	label string
	kind  AttributeKind
}

var (
	_ json.Marshaler   = AttributeType{}
	_ json.Unmarshaler = (*AttributeType)(nil)
)

func (a AttributeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *AttributeType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	kind, label, _ := strings.Cut(s, ".")
	var at AttributeKind
	if err := (&at).UnmarshalText([]byte(kind)); err != nil {
		return err
	}
	*a = AttributeType{kind: at, label: label}
	return nil
}

func (a AttributeType) Kind() AttributeKind { return a.kind }

func (a AttributeType) IsString() bool { return a.kind == AttributeKindString }

func (a AttributeType) IsNumber() bool { return a.kind == AttributeKindNumber }

func (a AttributeType) IsBinary() bool { return a.kind == AttributeKindBinary }

func (a AttributeType) Label() string { return a.label }

func (a AttributeType) IsCustom() bool { return a.label != "" }

func (a AttributeType) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString(a.kind.String())
	if a.IsCustom() {
		buf.WriteByte('.')
		buf.WriteString(a.label)
	}
	return buf.String()
}

var (
	ak2str = map[AttributeKind]string{
		AttributeKindString: "String",
		AttributeKindNumber: "Number",
		AttributeKindBinary: "Binary",
	}
	str2at = maps.Collect(seq2.Flip(maps.All(ak2str)))
)

const (
	AttributeKindString AttributeKind = iota
	AttributeKindNumber
	AttributeKindBinary
)

// AttributeKind represents the base data type category of an SQS message attribute (String, Number, or Binary).
type AttributeKind int

var (
	_ fmt.Stringer             = AttributeKind(0)
	_ encoding.TextMarshaler   = AttributeKind(0)
	_ encoding.TextUnmarshaler = (*AttributeKind)(nil)
)

func (k AttributeKind) String() string {
	s, ok := ak2str[k]
	if !ok {
		return fmt.Sprintf("INVALID.AttributeType(%d)", k)
	}
	return s
}

func (k AttributeKind) MarshalText() ([]byte, error) {
	s, ok := ak2str[k]
	if !ok {
		return nil, ErrInvalidAttributeType
	}
	return []byte(s), nil
}

func (k *AttributeKind) UnmarshalText(b []byte) error {
	s := string(b)
	at, ok := str2at[s]
	if !ok {
		return &UnknownAttributeTypeError{AttributeType: s}
	}
	*k = at
	return nil
}
