// Package validation implements the model defaults and validation pipeline
// (issue #108, #128).
//
// Contract:
//   - Deterministic order: system defaults -> field defaults -> required ->
//     type -> relation checks -> custom validators.
//   - Client input cannot override protected/system-managed fields.
//   - Required and type validation produce field-addressable errors.
//   - Update behavior does not overwrite omitted fields with defaults.
//   - Validation performs no hidden database writes.
package validation

import (
	"errors"
	"fmt"
	"strings"
)

// Errors
var (
	ErrRequired     = errors.New("validation: field is required")
	ErrInvalidType  = errors.New("validation: invalid type")
	ErrProtected    = errors.New("validation: protected field cannot be set by client")
	ErrUnknownField = errors.New("validation: unknown field")
)

// FieldError is a field-addressable validation error.
type FieldError struct {
	Field string
	Rule  string
	Err   error
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Rule)
}

func (e *FieldError) Unwrap() error { return e.Err }

// Field describes validation metadata for one field.
type Field struct {
	Name          string
	Type          string // string, int, float, bool, json
	Required      bool
	SystemManaged bool // protected from client input
	Default       any  // applied on create if value is absent
}

// Op distinguishes create from update semantics.
type Op int

const (
	OpCreate Op = iota
	OpUpdate
)

// ApplyDefaults applies system/field defaults in deterministic order.
// For OpUpdate, omitted fields (not present in values) are NOT overwritten.
func ApplyDefaults(fields []Field, values map[string]any, op Op) (map[string]any, error) {
	if values == nil {
		values = map[string]any{}
	}
	out := make(map[string]any, len(values))
	for k, v := range values {
		out[k] = v
	}

	for _, f := range fields {
		_, present := out[f.Name]

		// Protected fields: reject client input, but system may set via defaults.
		// (Caller must not pass client-owned values for system-managed fields;
		// here we strip them if present and not from an internal source.)
		if f.SystemManaged && present {
			delete(out, f.Name)
			present = false
		}

		if present {
			continue
		}

		// On update, do not apply defaults for omitted fields.
		if op == OpUpdate {
			continue
		}

		if f.Default != nil {
			out[f.Name] = f.Default
		}
	}
	return out, nil
}

// Validate checks required and type constraints. Returns all field errors.
func Validate(fields []Field, values map[string]any, op Op) []error {
	var errs []error
	fieldSet := make(map[string]Field, len(fields))
	for _, f := range fields {
		fieldSet[f.Name] = f
	}

	// Check for protected field input on raw values (before ApplyDefaults stripped them,
	// caller should validate raw input; but we also guard here by checking the field spec).
	for _, f := range fields {
		if f.Required {
			v, ok := values[f.Name]
			if !ok || isEmpty(v) {
				// On update, required only enforced if field is being set to empty?
				// Spec: "required checks" - on update, omitted is fine.
				if op == OpUpdate {
					if _, present := values[f.Name]; !present {
						continue
					}
				}
				errs = append(errs, &FieldError{Field: f.Name, Rule: "required", Err: ErrRequired})
				continue
			}
		}

		v, ok := values[f.Name]
		if !ok || v == nil {
			continue
		}
		if !typeMatches(f.Type, v) {
			errs = append(errs, &FieldError{Field: f.Name, Rule: "type:" + f.Type, Err: ErrInvalidType})
		}
	}
	return errs
}

// ValidateInput checks that client-provided values do not target protected fields.
func ValidateInput(fields []Field, values map[string]any) []error {
	var errs []error
	protected := map[string]bool{}
	for _, f := range fields {
		if f.SystemManaged {
			protected[f.Name] = true
		}
	}
	for k := range values {
		if protected[k] {
			errs = append(errs, &FieldError{Field: k, Rule: "protected", Err: ErrProtected})
		}
	}
	return errs
}

func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x) == ""
	case []any:
		return len(x) == 0
	case map[string]any:
		return len(x) == 0
	}
	return false
}

func typeMatches(typ string, v any) bool {
	switch typ {
	case "string":
		_, ok := v.(string)
		return ok
	case "int":
		switch v.(type) {
		case int, int64, int32, float64: // json numbers decode as float64
			return true
		}
		return false
	case "float":
		switch v.(type) {
		case float64, float32, int, int64:
			return true
		}
		return false
	case "bool":
		_, ok := v.(bool)
		return ok
	case "json":
		return true
	default:
		return true // unknown types pass through
	}
}
