package fields

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidSource       = errors.New("invalid field metadata source")
	ErrInvalidFieldName    = errors.New("invalid field name")
	ErrInvalidFieldType    = errors.New("invalid field type")
	ErrInvalidFieldLabel   = errors.New("invalid field label")
	ErrDuplicateField      = errors.New("duplicate field name")
	ErrInvalidFieldOption  = errors.New("invalid field option")
	ErrIncompatibleOption  = errors.New("field option is not supported by type")
	ErrInvalidDefault      = errors.New("default value does not match field type")
	ErrWritableSystemField = errors.New("system-managed field must be read-only")
	ErrInvalidFieldSet     = errors.New("invalid field set")
)

// ValidationError identifies the field and property that failed validation.
type ValidationError struct {
	Field  string
	Rule   error
	Detail string
}

func (e *ValidationError) Error() string {
	if e.Field == "" {
		return fmt.Sprintf("%v: %s", e.Rule, e.Detail)
	}
	return fmt.Sprintf("%v: field %q: %s", e.Rule, e.Field, e.Detail)
}

func (e *ValidationError) Unwrap() error { return e.Rule }
