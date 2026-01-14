package sub

import "fmt"

// ErrInvalidAttributeType is the sentinel error for [InvalidAttributeTypeError].
var ErrInvalidAttributeType InvalidAttributeTypeError

// InvalidAttributeTypeError indicates an [AttributeKind] value is out of the valid range.
type InvalidAttributeTypeError struct{}

var _ error = InvalidAttributeTypeError{}

func (InvalidAttributeTypeError) Error() string { return "invalid AttributeType" }

// UnknownAttributeTypeError indicates an unrecognized attribute type string was encountered during unmarshaling.
type UnknownAttributeTypeError struct {
	AttributeType string
}

var _ error = (*UnknownAttributeTypeError)(nil) //nolint:errcheck

func (e *UnknownAttributeTypeError) Error() string {
	return fmt.Sprintf("unknown AttributeType: %q", e.AttributeType)
}
