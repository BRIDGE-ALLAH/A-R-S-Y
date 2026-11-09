// Package aeng provides the Application Engine framework.
// AENG-036: Validate Add-On View and Action Metadata Against Runtime Registry.
package aeng

import (
	"fmt"
	"sort"
	"strings"
)

// MetadataValidator validates module manifest metadata against the runtime registry.
type MetadataValidator struct {
	registry *ModelRegistry
}

// NewMetadataValidator creates a validator backed by a registry.
func NewMetadataValidator(registry *ModelRegistry) *MetadataValidator {
	return &MetadataValidator{registry: registry}
}

// ValidationError describes a single validation failure.
type ValidationError struct {
	ModuleID string // Module that has the error
	ErrorID  string // Unique error identifier
	Message  string // Human-readable message
	Field    string // Optional: specific field name
}

// ValidationWarning describes a non-fatal validation concern.
type ValidationWarning struct {
	ModuleID     string
	Message      string
	SuggestedFix string
}

// Result holds validation errors and warnings.
type Result struct {
	Errors    []ValidationError
	Warnings  []ValidationWarning
	Validated bool
}

// ValidateAll runs every validation check against all loaded modules.
func (v *MetadataValidator) ValidateAll() *Result {
	v.registry.mu.RLock()
	defer v.registry.mu.RUnlock()

	res := &Result{Validated: true}

	for _, mod := range v.registry.modules {
		if mod.Manifest == nil {
			continue
		}
		res.merge(v.checkViewFieldReferences(mod.ID, mod.Manifest))
		res.merge(v.checkActionInputFields(mod.ID, mod.Manifest))
		res.merge(v.checkSecurityFieldPermissions(mod.ID, mod.Manifest))
		res.merge(v.checkRelationTargets(mod.ID, mod.Manifest))
	}

	return res
}

func (r *Result) merge(other *Result) {
	r.Errors = append(r.Errors, other.Errors...)
	r.Warnings = append(r.Warnings, other.Warnings...)
}

// Err returns a combined error if validation found failures.
func (r *Result) Err() error {
	if len(r.Errors) == 0 {
		return nil
	}
	msgs := make([]string, 0, len(r.Errors))
	for _, e := range r.Errors {
		msgs = append(msgs, fmt.Sprintf("[%s] %s: %s", e.ModuleID, e.ErrorID, e.Message))
	}
	return fmt.Errorf("metadata validation failed: %s", strings.Join(msgs, "; "))
}

// Report returns a human-readable validation report.
func (r *Result) Report() string {
	var b strings.Builder
	if len(r.Errors) > 0 {
		b.WriteString(fmt.Sprintf("Validation Errors (%d):\n", len(r.Errors)))
		for _, err := range r.Errors {
			b.WriteString(fmt.Sprintf("  [%s] %s: %s\n", err.ModuleID, err.ErrorID, err.Message))
		}
	}
	if len(r.Warnings) > 0 {
		b.WriteString(fmt.Sprintf("Validation Warnings (%d):\n", len(r.Warnings)))
		for _, w := range r.Warnings {
			b.WriteString(fmt.Sprintf("  [%s] %s\n", w.ModuleID, w.Message))
		}
	}
	if len(r.Errors) == 0 && len(r.Warnings) == 0 {
		b.WriteString("Validation passed: all metadata is consistent")
	}
	return b.String()
}

// checkViewFieldReferences validates that view field sources exist in models.
func (v *MetadataValidator) checkViewFieldReferences(moduleID string, manifest *ModuleManifest) *Result {
	res := &Result{Validated: true}
	for _, view := range manifest.Views {
		if view == nil {
			continue
		}
		for _, field := range view.Fields {
			if !fieldExistsInModel(moduleID, field.Source, v.registry) {
				res.Errors = append(res.Errors, ValidationError{
					ModuleID: moduleID,
					ErrorID:  "view_field_not_found",
					Message:  fmt.Sprintf("view %s: field %s not found in any model", view.Name, field.Source),
					Field:    field.Source,
				})
			}
		}
	}
	return res
}

// checkActionInputFields validates action input fields exist in models.
func (v *MetadataValidator) checkActionInputFields(moduleID string, manifest *ModuleManifest) *Result {
	res := &Result{Validated: true}
	for _, action := range manifest.Actions {
		if action == nil || action.Input == nil {
			continue
		}
		for fieldName := range action.Input.Fields {
			if !fieldExistsInModel(moduleID, fieldName, v.registry) {
				res.Errors = append(res.Errors, ValidationError{
					ModuleID: moduleID,
					ErrorID:  "action_input_field_not_found",
					Message:  fmt.Sprintf("action %s: input field %s not found in any model", action.Name, fieldName),
					Field:    fieldName,
				})
			}
		}
	}
	return res
}

// checkSecurityFieldPermissions validates security field permissions reference model fields.
func (v *MetadataValidator) checkSecurityFieldPermissions(moduleID string, manifest *ModuleManifest) *Result {
	res := &Result{Validated: true}
	if manifest.Security == nil {
		return res
	}
	for _, resource := range manifest.Security.Resources {
		for _, fp := range resource.FieldPermissions {
			if !fieldExistsInModel(moduleID, fp.Field, v.registry) {
				res.Errors = append(res.Errors, ValidationError{
					ModuleID: moduleID,
					ErrorID:  "security_field_not_in_model",
					Message:  fmt.Sprintf("security: field %s not found in model %s", fp.Field, resource.Name),
					Field:    fp.Field,
				})
			}
		}
	}
	return res
}

// checkRelationTargets validates relation targets reference existing models.
func (v *MetadataValidator) checkRelationTargets(moduleID string, manifest *ModuleManifest) *Result {
	res := &Result{Validated: true}
	for _, model := range manifest.Models {
		for _, rel := range model.Relations {
			if _, ok := v.registry.modelByName[rel.Target]; !ok {
				res.Errors = append(res.Errors, ValidationError{
					ModuleID: moduleID,
					ErrorID:  "relation_target_not_found",
					Message:  fmt.Sprintf("model %s: relation %s target %s not found", model.Name, rel.Name, rel.Target),
				})
			}
		}
	}
	return res
}

// fieldExistsInModel checks whether a field name exists in any model of the module.
func fieldExistsInModel(moduleID, fieldName string, registry *ModelRegistry) bool {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	for _, mod := range registry.modules {
		if mod.ID != moduleID {
			continue
		}
		if mod.Manifest == nil {
			continue
		}
		for _, model := range mod.Manifest.Models {
			for _, f := range model.Fields {
				if f.Name == fieldName {
					return true
				}
			}
		}
	}
	return false
}

// ValidateModuleOnLoad runs validation for a single module manifest.
func ValidateModuleOnLoad(registry *ModelRegistry, moduleID string, manifest *ModuleManifest) error {
	if registry == nil {
		return fmt.Errorf("aeng: nil registry")
	}
	if manifest == nil {
		return fmt.Errorf("aeng: nil manifest for module %s", moduleID)
	}

	v := NewMetadataValidator(registry)
	res := &Result{Validated: true}

	res.merge(v.checkViewFieldReferences(moduleID, manifest))
	res.merge(v.checkActionInputFields(moduleID, manifest))
	res.merge(v.checkSecurityFieldPermissions(moduleID, manifest))
	res.merge(v.checkRelationTargets(moduleID, manifest))

	if len(res.Errors) > 0 {
		sort.Slice(res.Errors, func(i, j int) bool {
			return res.Errors[i].ModuleID+res.Errors[i].ErrorID < res.Errors[j].ModuleID+res.Errors[j].ErrorID
		})
		return res.Err()
	}
	return nil
}
