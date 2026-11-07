package manifest

import "regexp"

const (
	CurrentSchemaVersion = "1.0"
)

var (
	moduleNameRegex = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$`)
	semverRegex     = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)
)

// Manifest defines an A-R-S-Y add-on module manifest.
type Manifest struct {
	SchemaVersion string       `json:"schema_version" yaml:"schema_version"`
	Name          string       `json:"name" yaml:"name"`
	Version       string       `json:"version" yaml:"version"`
	Title         string       `json:"title,omitempty" yaml:"title,omitempty"`
	Description   string       `json:"description,omitempty" yaml:"description,omitempty"`
	License       string       `json:"license,omitempty" yaml:"license,omitempty"`
	Authors       []string     `json:"authors,omitempty" yaml:"authors,omitempty"`
	Depends       []Dependency `json:"depends,omitempty" yaml:"depends,omitempty"`
	Models        []string     `json:"models,omitempty" yaml:"models,omitempty"`
	Data          []string     `json:"data,omitempty" yaml:"data,omitempty"`
	Security      []string     `json:"security,omitempty" yaml:"security,omitempty"`
	Views         []string     `json:"views,omitempty" yaml:"views,omitempty"`
	Migrations    []string     `json:"migrations,omitempty" yaml:"migrations,omitempty"`
}

// Dependency defines a module dependency with an optional version constraint.
type Dependency struct {
	Name       string `json:"name" yaml:"name"`
	Constraint string `json:"constraint,omitempty" yaml:"constraint,omitempty"`
}
