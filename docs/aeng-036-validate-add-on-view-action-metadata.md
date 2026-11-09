# AENG-036: Validate Add-On View and Action Metadata Against Runtime Registry

**Status:** Draft
**Owner:** A-R-S-Y Team
**Date:** 2026-11-09
**Depends On:** AENG-020, AENG-006, AENG-019

## Purpose
Implement validation of add-on module view and action metadata against the runtime ModelRegistry, ensuring that all declared views and actions reference existing models, fields, and actions, and that security policies are consistent.

---

## 1. Validation Purpose

### 1.1 Why Validate at Runtime?
- Catch mismatches between manifest declarations and actual PocketBase schema
- Prevent runtime panics from missing fields/models/actions
- Ensure security metadata consistency across modules
- Provide clear error messages during module load, not at request time

### 1.2 Validation Scopes
| Scope | What's Validated | When |
|-------|-----------------|------|
| **Model fields** | View fields reference existing model fields | On module load |
| **Relation targets** | Relation names reference existing models | On module load |
| **Action inputs** | Action input fields exist in models | On module load |
| **Action compute functions** | Compute function names exist | On module load |
| **Security metadata** | Field permissions reference existing fields | On module load |
| **Cross-module refs** | Inter-module references are valid | On module load |

---

## 2. Validation Infrastructure

### 2.1 Validator Structure
```go
type MetadataValidator struct {
    registry   *ModelRegistry
    errors     []ValidationError
    warnings   []ValidationWarning
    validated  map[string]bool // modules already validated
}
```

### 2.2 Validation Error
```go
type ValidationError struct {
    ModuleID   string // Module that has the error
    ErrorID   string // Unique error identifier
    Message   string // Human-readable message
    Field     string // Optional: specific field name
    Line      int    // Optional: manifest line number
}
```

### 2.3 Validation Warning
```go
type ValidationWarning struct {
    ModuleID   string
    Message   string
    SuggestedFix string // Optional suggestion
}
```

---

## 3. Validation Checks

### 3.1 Check View Field References
```go
func (v *MetadataValidator) checkViewFieldReferences() {
    for _, mod := range v.registry.getModules() {
        manifest := mod.Manifest
        if manifest == nil {
            continue
        }
        
        for _, view := range manifest.Views {
            for _, field := range view.Fields {
                // Check if field source exists in model
                model := v.registry.getModel(field.Source)
                if model == nil {
                    v.errors = append(v.errors, ValidationError{
                        ModuleID: mod.ID,
                        ErrorID:  "view_field_not_found",
                        Message:  fmt.Sprintf("view %s: field %s not found in any model", view.Name, field.Source),
                    })
                    continue
                }
                
                // Check if field exists in model's fields
                fieldExists := false
                for _, mf := range model.Fields {
                    if mf.Name == field.Source || mf.GoName == field.Source {
                        fieldExists = true
                        break
                    }
                }
                
                if !fieldExists {
                    v.errors = append(v.errors, ValidationError{
                        ModuleID: mod.ID,
                        ErrorID:  "view_field_not_in_model",
                        Message:  fmt.Sprintf("view %s: field %s not in model %s", view.Name, field.Source, model.Name),
                    })
                }
            }
        }
    }
}
```

### 3.2 Check Action Input Fields
```go
func (v *MetadataValidator) checkActionInputFields() {
    for _, mod := range v.registry.getModules() {
        manifest := mod.Manifest
        if manifest == nil {
            continue
        }
        
        for _, action := range manifest.Actions {
            if action.Input == nil {
                continue
            }
            
            for fieldName := range action.Input.Fields {
                // Check if field exists in any model
                found := false
                for _, m := range manifest.Models {
                    for _, f := range m.Fields {
                        if f.Name == fieldName {
                            found = true
                            break
                        }
                    }
                    if found {
                        break
                    }
                }
                
                if !found {
                    v.errors = append(v.errors, ValidationError{
                        ModuleID: mod.ID,
                        ErrorID:  "action_input_field_not_found",
                        Message:  fmt.Sprintf("action %s: input field %s not found in any model", action.Name, fieldName),
                    })
                }
            }
        }
    }
}
```

### 3.3 Check Action Compute Functions
```go
func (v *MetadataValidator) checkActionComputeFunctions() {
    for _, mod := range v.registry.getModules() {
        manifest := mod.Manifest
        if manifest == nil {
            continue
        }
        
        for _, action := range manifest.Actions {
            if action.ComputeFn == "" {
                continue // no compute function, skip
            }
            
            // Check if function exists in the module's registered functions
            fnExists := v.registry.functionExists(action.ComputeFn, mod.ID)
            if !fnExists {
                v.errors = append(v.errors, ValidationError{
                    ModuleID: mod.ID,
                    ErrorID:  "action_compute_fn_not_found",
                    Message:  fmt.Sprintf("action %s: compute function %s not found", action.Name, action.ComputeFn),
                })
            }
        }
    }
}
```

### 3.4 Check Security Field Permissions
```go
func (v *MetadataValidator) checkSecurityFieldPermissions() {
    for _, mod := range v.registry.getModules() {
        security := v.registry.getSecurity(mod.ID)
        if security == nil {
            continue
        }
        
        for _, res := range security.Resources {
            for _, fp := range res.FieldPermissions {
                // Check if field exists in the resource's model
                model := v.registry.getModel(res.Name)
                if model == nil {
                    continue
                }
                
                fieldExists := false
                for _, mf := range model.Fields {
                    if mf.Name == fp.Field {
                        fieldExists = true
                        break
                    }
                }
                
                if !fieldExists {
                    v.errors = append(v.errors, ValidationError{
                        ModuleID: mod.ID,
                        ErrorID:  "security_field_not_in_model",
                        Message:  fmt.Sprintf("security: field %s not found in model %s", fp.Field, res.Name),
                    })
                }
            }
        }
    }
}
```

---

## 4. Validation Execution

### 4.1 Run All Validations
```go
func (v *MetadataValidator) ValidateAll() error {
    v.checkViewFieldReferences()
    v.checkActionInputFields()
    v.checkActionComputeFunctions()
    v.checkSecurityFieldPermissions()
    // ... other checks
    
    if len(v.errors) > 0 {
        return fmt.Errorf("metadata validation failed %d errors: %v", len(v.errors), v.errors)
    }
    return nil
}
```

### 3.2 Validate on Module Load
```go
func ValidateModuleOnLoad(registry *ModelRegistry, moduleID string) error {
    validator := newMetadataValidator(registry)
    validator.validateModule(moduleID)
    return validator.GetErrors()
}
```

### 3.3 Validation Order
1. Model field references (views, security)
2. Action references (inputs, compute functions)
3. Security permission consistency
4. Cross-module dependency validity

---

## 5. Validation Report

### 5.1 Human-Readable Report
```go
func (v *MetadataValidator) Report() string {
    var b strings.Builder
    
    if len(v.errors) > 0 {
        b.WriteString(fmt.Sprintf("Validation Errors (%d):\n", len(v.errors)))
        for _, err := range v.errors {
            b.WriteString(fmt.Sprintf("  [%s] %s: %s\n", err.ModuleID, err.ErrorID, err.Message))
        }
    }
    
    if len(v.warnings) > 0 {
        b.WriteString(fmt.Sprintf("Validation Warnings (%d):\n", len(v.warnings)))
        for _, w := range v.warnings {
            b.WriteString(fmt.Sprintf("  [%s] %s\n", w.ModuleID, w.Message))
        }
    }
    
    if len(v.errors) == 0 && len(v.warnings) == 0 {
        b.WriteString("Validation passed: all metadata is consistent")
    }
    
    return b.String()
}
```

### 5.2 Example Report
```
Validation Errors (2):
  [base] view_field_not_found: view company_summary: field employee_count not found in any model
  [crm] action_input_field_not_found: action archive: input field confirm not found in any model
```

---

## 6. Integration with Module Loader

### 6.1 Module Load Sequence with Validation
```
1. Read module manifest from disk
2. Parse manifest into Go struct
3. Run MetadataValidator.ValidateAll()
4. If errors: return error to caller, don't load module
5. If valid: index models/actions/security in ModelRegistry
6. Module is now available for use
```

### 6.2 Example Module Load Flow
```go
func loadModule(ctx context.Context, registry *ModelRegistry, modulePath string) error {
    // 1. Read manifest
    data, err := os.ReadFile(filepath.Join(modulePath, "manifest.json"))
    if err != nil {
        return err
    }
    
    // 2. Parse
    var manifest ModuleManifest
    if err := json.Unmarshal(data, &manifest); err != nil {
        return err
    }
    
    // 3. Validate
    if err := ValidateModuleOnLoad(registry, manifest.Module); err != nil {
        return fmt.Errorf("module %s validation failed: %w", manifest.Module, err)
    }
    
    // 4. Index in registry
    if err := registry.IndexModule(manifest); err != nil {
        return err
    }
    
    return nil
}
```

---

## 6. Open Issues

1. **Performance:** Validating on every module load vs. one-time validation at deployment
2. **Automatic fixing:** Should the validator suggest fixes or just report errors?
3. **Version-specific validation:** Should validation be version-aware (v1 vs v2 schemas)?
4. **Partial validation:** Can we validate incrementally (one check at a time)?
5. **Integration with codegen:** Should validation inform code generation?

---

## 7. References

- [AENG-006 ORM Architecture](aeng-006-orm-architecture.md)
- [AENG-019 Module Security Metadata](aeng-019-module-security-metadata.md)
- [AENG-020 View/Action Manifest](aeng-020-view-action-manifest.md)
- [AENG-001 PocketBase Capabilities](aeng-001-pocketbase-capabilities.md)