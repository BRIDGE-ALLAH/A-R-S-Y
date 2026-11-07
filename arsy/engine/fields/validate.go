package fields

import (
	"encoding/json"
	"fmt"
	"math"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	namePattern    = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	decimalPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?$`)
	systemFields   = map[string]struct{}{
		"id": {}, "created": {}, "updated": {},
		"collectionId": {}, "collectionName": {},
	}
)

func fieldError(name string, rule error, detail string) error {
	return &ValidationError{Field: name, Rule: rule, Detail: detail}
}

// Validate validates and normalizes a field set. The result is sorted by field
// name, contains detached option slices/pointers, and is safe to serialize.
// Relation targets are deliberately not interpreted here.
func Validate(input []Descriptor, source Source) ([]Descriptor, error) {
	if source != SourceStaticGo && source != SourceMetadata {
		return nil, &ValidationError{Rule: ErrInvalidSource, Detail: "source must be static_go or metadata"}
	}
	out := make([]Descriptor, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for _, original := range input {
		d, err := normalize(original, source)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[d.Name]; exists {
			return nil, fieldError(d.Name, ErrDuplicateField, "field names must be unique within a model")
		}
		seen[d.Name] = struct{}{}
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func normalize(in Descriptor, source Source) (Descriptor, error) {
	d := in
	_, reservedName := systemFields[d.Name]
	if !namePattern.MatchString(d.Name) && !reservedName {
		return d, fieldError(d.Name, ErrInvalidFieldName, "use lowercase letters, digits, and underscores; the first character must be a letter")
	}
	if strings.TrimSpace(d.Label) == "" {
		return d, fieldError(d.Name, ErrInvalidFieldLabel, "label must not be empty")
	}
	d.Label = strings.TrimSpace(d.Label)
	d.HelpText = strings.TrimSpace(d.HelpText)

	if _, reserved := systemFields[d.Name]; reserved {
		if source == SourceMetadata && !d.ReadOnly {
			return d, fieldError(d.Name, ErrWritableSystemField, "untrusted metadata cannot declare a system-managed field writable")
		}
		d.ReadOnly = true
		d.SystemManaged = true
	} else {
		// Never allow a manifest to grant itself system-managed privileges.
		d.SystemManaged = false
	}

	if err := validateOptions(d); err != nil {
		return d, err
	}
	if d.Default != nil {
		if err := validateDefault(d); err != nil {
			return d, err
		}
	}
	if d.Unique && !d.Index {
		// Unique constraints are index-backed in the supported storage engines.
		d.Index = true
	}
	d.Options = cloneOptions(d.Options)
	return d, nil
}

func validateOptions(d Descriptor) error {
	o := d.Options
	if o == nil {
		if d.Type == TypeSelect {
			return fieldError(d.Name, ErrInvalidFieldOption, "select fields require at least one choice")
		}
		return validateType(d)
	}
	switch d.Type {
	case TypeString, TypeText:
		if o.MinLength != nil && *o.MinLength < 0 {
			return fieldError(d.Name, ErrInvalidFieldOption, "min_length cannot be negative")
		}
		if o.MaxLength != nil && *o.MaxLength < 1 {
			return fieldError(d.Name, ErrInvalidFieldOption, "max_length must be positive")
		}
		if o.MinLength != nil && o.MaxLength != nil && *o.MinLength > *o.MaxLength {
			return fieldError(d.Name, ErrInvalidFieldOption, "min_length cannot exceed max_length")
		}
		if o.Pattern != "" {
			if _, err := regexp.Compile(o.Pattern); err != nil {
				return fieldError(d.Name, ErrInvalidFieldOption, "pattern must be a valid Go regular expression")
			}
		}
		if o.Min != nil || o.Max != nil || o.Precision != nil || o.Scale != nil || len(o.Choices) > 0 {
			return incompatible(d.Name, d.Type)
		}
	case TypeInteger, TypeNumber:
		if (o.Min != nil && (math.IsNaN(*o.Min) || math.IsInf(*o.Min, 0))) || (o.Max != nil && (math.IsNaN(*o.Max) || math.IsInf(*o.Max, 0))) {
			return fieldError(d.Name, ErrInvalidFieldOption, "numeric bounds must be finite")
		}
		if o.Min != nil && o.Max != nil && *o.Min > *o.Max {
			return fieldError(d.Name, ErrInvalidFieldOption, "min cannot exceed max")
		}
		if d.Type == TypeInteger && ((o.Min != nil && math.Trunc(*o.Min) != *o.Min) || (o.Max != nil && math.Trunc(*o.Max) != *o.Max)) {
			return fieldError(d.Name, ErrInvalidFieldOption, "integer bounds must be whole numbers")
		}
		if o.MinLength != nil || o.MaxLength != nil || o.Pattern != "" || o.Precision != nil || o.Scale != nil || len(o.Choices) > 0 {
			return incompatible(d.Name, d.Type)
		}
	case TypeDecimal:
		if o.Precision == nil || o.Scale == nil || *o.Precision < 1 || *o.Scale < 0 || *o.Scale > *o.Precision {
			return fieldError(d.Name, ErrInvalidFieldOption, "decimal requires precision >= 1 and scale between 0 and precision")
		}
		if o.MinLength != nil || o.MaxLength != nil || o.Pattern != "" || o.Min != nil || o.Max != nil || len(o.Choices) > 0 {
			return incompatible(d.Name, d.Type)
		}
	case TypeSelect:
		if len(o.Choices) == 0 {
			return fieldError(d.Name, ErrInvalidFieldOption, "select fields require at least one choice")
		}
		seen := make(map[string]struct{}, len(o.Choices))
		for _, choice := range o.Choices {
			if strings.TrimSpace(choice) == "" {
				return fieldError(d.Name, ErrInvalidFieldOption, "select choices cannot be empty")
			}
			if _, ok := seen[choice]; ok {
				return fieldError(d.Name, ErrInvalidFieldOption, "select choices must be unique")
			}
			seen[choice] = struct{}{}
		}
		if o.MinLength != nil || o.MaxLength != nil || o.Pattern != "" || o.Min != nil || o.Max != nil || o.Precision != nil || o.Scale != nil {
			return incompatible(d.Name, d.Type)
		}
	default:
		if o.MinLength != nil || o.MaxLength != nil || o.Pattern != "" || o.Min != nil || o.Max != nil || o.Precision != nil || o.Scale != nil || len(o.Choices) > 0 {
			return incompatible(d.Name, d.Type)
		}
	}
	return validateType(d)
}

func validateType(d Descriptor) error {
	switch d.Type {
	case TypeString, TypeText, TypeEmail, TypeURL, TypeBoolean, TypeInteger, TypeNumber, TypeDecimal, TypeDate, TypeDateTime, TypeJSON, TypeSelect:
		return nil
	default:
		return fieldError(d.Name, ErrInvalidFieldType, fmt.Sprintf("unsupported field type %q", d.Type))
	}
}

func incompatible(name string, typ Type) error {
	return fieldError(name, ErrIncompatibleOption, fmt.Sprintf("options supplied are not supported for type %q", typ))
}

func validateDefault(d Descriptor) error {
	fail := func(detail string) error { return fieldError(d.Name, ErrInvalidDefault, detail) }
	switch d.Type {
	case TypeString, TypeText, TypeEmail, TypeURL, TypeDate, TypeDateTime, TypeDecimal, TypeSelect:
		value, ok := d.Default.(string)
		if !ok {
			return fail("expected a string value")
		}
		switch d.Type {
		case TypeEmail:
			addr, err := mail.ParseAddress(value)
			if err != nil || addr.Address != value {
				return fail("expected a valid email address")
			}
		case TypeURL:
			u, err := url.ParseRequestURI(value)
			if err != nil || u.Scheme == "" || u.Host == "" {
				return fail("expected an absolute URL with scheme and host")
			}
		case TypeDate:
			if _, err := time.Parse("2006-01-02", value); err != nil {
				return fail("expected date format YYYY-MM-DD")
			}
		case TypeDateTime:
			if _, err := time.Parse(time.RFC3339, value); err != nil {
				return fail("expected an RFC3339 datetime")
			}
		case TypeDecimal:
			if !decimalPattern.MatchString(value) {
				return fail("expected a decimal string to avoid floating-point precision loss")
			}
			if d.Options != nil {
				parts := strings.Split(strings.TrimPrefix(value, "-"), ".")
				integerDigits := len(strings.TrimLeft(parts[0], "0"))
				if integerDigits == 0 {
					integerDigits = 1
				}
				fractionDigits := 0
				if len(parts) == 2 {
					fractionDigits = len(parts[1])
				}
				if integerDigits+fractionDigits > *d.Options.Precision || fractionDigits > *d.Options.Scale {
					return fail("value exceeds decimal precision or scale")
				}
			}
		case TypeSelect:
			found := false
			if d.Options != nil {
				for _, choice := range d.Options.Choices {
					if value == choice {
						found = true
						break
					}
				}
			}
			if !found {
				return fail("default must match one of the declared choices")
			}
		}
		if d.Options != nil && (d.Type == TypeString || d.Type == TypeText) {
			length := len([]rune(value))
			if d.Options.MinLength != nil && length < *d.Options.MinLength {
				return fail("default is shorter than min_length")
			}
			if d.Options.MaxLength != nil && length > *d.Options.MaxLength {
				return fail("default is longer than max_length")
			}
			if d.Options.Pattern != "" && !regexp.MustCompile(d.Options.Pattern).MatchString(value) {
				return fail("default does not match pattern")
			}
		}
	case TypeBoolean:
		if _, ok := d.Default.(bool); !ok {
			return fail("expected a boolean value")
		}
	case TypeInteger:
		if !isInteger(d.Default) {
			return fail("expected an integer value")
		}
		if d.Options != nil {
			n, _ := numericValue(d.Default)
			if d.Options.Min != nil && n < *d.Options.Min {
				return fail("default is below min")
			}
			if d.Options.Max != nil && n > *d.Options.Max {
				return fail("default is above max")
			}
		}
	case TypeNumber:
		n, ok := numericValue(d.Default)
		if !ok || math.IsNaN(n) || math.IsInf(n, 0) {
			return fail("expected a finite numeric value")
		}
		if d.Options != nil {
			if d.Options.Min != nil && n < *d.Options.Min {
				return fail("default is below min")
			}
			if d.Options.Max != nil && n > *d.Options.Max {
				return fail("default is above max")
			}
		}
	case TypeJSON:
		if _, err := json.Marshal(d.Default); err != nil {
			return fail("value must be JSON serializable")
		}
	}
	return nil
}

func isInteger(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, json.Number:
		if n, ok := numericValue(v); ok {
			return n == float64(int64(n))
		}
	}
	return false
}

func numericValue(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case json.Number:
		f, err := strconv.ParseFloat(string(n), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func cloneOptions(o *Options) *Options {
	if o == nil {
		return nil
	}
	c := *o
	if o.MinLength != nil {
		v := *o.MinLength
		c.MinLength = &v
	}
	if o.MaxLength != nil {
		v := *o.MaxLength
		c.MaxLength = &v
	}
	if o.Min != nil {
		v := *o.Min
		c.Min = &v
	}
	if o.Max != nil {
		v := *o.Max
		c.Max = &v
	}
	if o.Precision != nil {
		v := *o.Precision
		c.Precision = &v
	}
	if o.Scale != nil {
		v := *o.Scale
		c.Scale = &v
	}
	c.Choices = append([]string(nil), o.Choices...)
	return &c
}

// MarshalCanonical serializes a validated field set into deterministic JSON.
// Validation and name sorting are repeated so callers cannot accidentally
// serialize an invalid or nondeterministically ordered set.
func MarshalCanonical(input []Descriptor, source Source) ([]byte, error) {
	normalized, err := Validate(input, source)
	if err != nil {
		return nil, err
	}
	return json.Marshal(normalized)
}
