package fields

import "encoding/json"

// Type is a supported, portable application field type.
type Type string

const (
	TypeString   Type = "string"
	TypeText     Type = "text"
	TypeEmail    Type = "email"
	TypeURL      Type = "url"
	TypeBoolean  Type = "boolean"
	TypeInteger  Type = "integer"
	TypeNumber   Type = "number"
	TypeDecimal  Type = "decimal"
	TypeDate     Type = "date"
	TypeDateTime Type = "datetime"
	TypeJSON     Type = "json"
	TypeSelect   Type = "select"
)

// Source identifies whether field metadata came from trusted Go code or an
// untrusted, installable module manifest.
type Source string

const (
	SourceStaticGo Source = "static_go"
	SourceMetadata Source = "metadata"
)

// Options contains constraints that are meaningful only for selected types.
// A nil pointer means that the constraint was not supplied.
type Options struct {
	MinLength *int     `json:"min_length,omitempty"`
	MaxLength *int     `json:"max_length,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	Precision *int     `json:"precision,omitempty"`
	Scale     *int     `json:"scale,omitempty"`
	Choices   []string `json:"choices,omitempty"`
}

// Descriptor is the declarative contract for one application model field.
// Default must be JSON-compatible and must match the declared field type.
type Descriptor struct {
	Name     string   `json:"name"`
	Type     Type     `json:"type"`
	Label    string   `json:"label"`
	Required bool     `json:"required,omitempty"`
	ReadOnly bool     `json:"readonly,omitempty"`
	Default  any      `json:"default,omitempty"`
	Index    bool     `json:"index,omitempty"`
	Unique   bool     `json:"unique,omitempty"`
	HelpText string   `json:"help_text,omitempty"`
	Options  *Options `json:"options,omitempty"`
	// SystemManaged is derived by validation for reserved engine fields. It is
	// never trusted as an input permission from a module manifest.
	SystemManaged bool `json:"system_managed,omitempty"`
}

// IsZeroDefault reports whether the descriptor has no explicit default.
// The JSON representation of a nil interface is omitted from metadata.
func (d Descriptor) IsZeroDefault() bool { return d.Default == nil }

// MarshalDefault is a small helper for callers that need the canonical JSON
// representation of a single field default.
func (d Descriptor) MarshalDefault() ([]byte, error) { return json.Marshal(d.Default) }
