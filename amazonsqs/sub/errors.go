package sub

import "fmt"

var ErrInvalidAttributeType InvalidAttributeTypeError

type InvalidAttributeTypeError struct{}

var _ error = InvalidAttributeTypeError{}

func (InvalidAttributeTypeError) Error() string { return "invalid AttributeType" }

type UnknownAttributeTypeError struct {
	AttributeType string
}

var _ error = (*UnknownAttributeTypeError)(nil) //nolint:errcheck

func (e *UnknownAttributeTypeError) Error() string {
	return fmt.Sprintf("unknown AttributeType: %q", e.AttributeType)
}
