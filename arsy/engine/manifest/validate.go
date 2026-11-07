package manifest

import (
	"fmt"
	"path/filepath"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func (m *Manifest) Validate() error {
	var errs []string

	if m.SchemaVersion == "" {
		errs = append(errs, "schema_version is required")
	} else if m.SchemaVersion != CurrentSchemaVersion {
		errs = append(errs, fmt.Sprintf("schema_version must be %q", CurrentSchemaVersion))
	}

	if m.Name == "" {
		errs = append(errs, "name is required")
	} else if !moduleNameRegex.MatchString(m.Name) {
		errs = append(errs, "name must be a valid module identifier (lowercase, dot-separated)")
	}

	if m.Version == "" {
		errs = append(errs, "version is required")
	} else if !semverRegex.MatchString(m.Version) {
		errs = append(errs, "version must be a valid semver string")
	}

	for i, dep := range m.Depends {
		if dep.Name == "" {
			errs = append(errs, fmt.Sprintf("depends[%d].name is required", i))
		} else if !moduleNameRegex.MatchString(dep.Name) {
			errs = append(errs, fmt.Sprintf("depends[%d].name %q is invalid", i, dep.Name))
		}
	}

	if err := validatePaths("models", m.Models); err != nil {
		errs = append(errs, err.Error())
	}
	if err := validatePaths("data", m.Data); err != nil {
		errs = append(errs, err.Error())
	}
	if err := validatePaths("security", m.Security); err != nil {
		errs = append(errs, err.Error())
	}
	if err := validatePaths("views", m.Views); err != nil {
		errs = append(errs, err.Error())
	}
	if err := validatePaths("migrations", m.Migrations); err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %s", ErrInvalidManifest, strings.Join(errs, "; "))
	}
	return nil
}

func validatePaths(field string, paths []string) error {
	for i, p := range paths {
		if p == "" {
			return fmt.Errorf("%s[%d]: path cannot be empty", field, i)
		}
		if filepath.IsAbs(p) {
			return fmt.Errorf("%s[%d]: path %q must be relative", field, i, p)
		}
		clean := filepath.ToSlash(filepath.Clean(p))
		if strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") || clean == ".." {
			return fmt.Errorf("%s[%d]: path %q escapes module directory", field, i, p)
		}
	}
	return nil
}
