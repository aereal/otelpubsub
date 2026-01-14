package sub

import "fmt"

// ErrInvalidAttributeKind is the sentinel error for [InvalidAttributeKindError].
var ErrInvalidAttributeKind InvalidAttributeKindError

// InvalidAttributeKindError indicates an [AttributeKind] value is out of the valid range.
type InvalidAttributeKindError struct{}

var _ error = InvalidAttributeKindError{}

func (InvalidAttributeKindError) Error() string { return "invalid AttributeType" }

// UnknownAttributeKindError indicates an unrecognized attribute type string was encountered during unmarshaling.
type UnknownAttributeKindError struct {
	Kind string
}

var _ error = (*UnknownAttributeKindError)(nil) //nolint:errcheck

func (e *UnknownAttributeKindError) Error() string {
	return fmt.Sprintf("unknown attribute kind: %q", e.Kind)
}
