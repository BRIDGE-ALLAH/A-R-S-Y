package persist

import "fmt"

// ErrNotFound is returned when a requested record is not found.
type ErrNotFound struct {
	ModelID  string
	RecordID string
}

func (e *ErrNotFound) Error() string {
	return "persist: model " + e.ModelID + " record " + e.RecordID + " not found"
}

// ErrConstraintViolation is returned when a constraint violation occurs (e.g., unique key).
type ErrConstraintViolation struct {
	ModelID    string
	Field      string
	Value      interface{}
	Constraint string // e.g., "uq_email"
}

func (e *ErrConstraintViolation) Error() string {
	return "persist: model " + e.ModelID + " constraint violation on field " + e.Field + " with value " + fmt.Sprintf("%v", e.Value)
}

// ErrFieldConversion is returned when a field value cannot be converted to the expected type.
type ErrFieldConversion struct {
	ModelID    string
	Field      string
	Value      interface{}
	TargetType string // e.g., "number", "bool"
}

func (e *ErrFieldConversion) Error() string {
	return "persist: model " + e.ModelID + " cannot convert field " + e.Field + " value " + fmt.Sprintf("%v", e.Value) + " to type " + e.TargetType
}

// ErrUnknownModel is returned when the model ID is not registered in the registry.
type ErrUnknownModel struct {
	ModelID string
}

func (e *ErrUnknownModel) Error() string {
	return "persist: model " + e.ModelID + " is not registered"
}

// ErrUnknownField is returned when a field name is not recognized for the given model.
type ErrUnknownField struct {
	ModelID  string
	Field    string
}

func (e *ErrUnknownField) Error() string {
	return "persist: model " + e.ModelID + " has no field " + e.Field
}
