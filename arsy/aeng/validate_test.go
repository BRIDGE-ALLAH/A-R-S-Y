package aeng

import (
	"context"
	"testing"
)

func TestNewMetadataValidator(t *testing.T) {
	r := NewModelRegistry()
	v := NewMetadataValidator(r)
	if v == nil {
		t.Fatal("NewMetadataValidator() returned nil")
	}
	if v.registry != r {
		t.Fatal("validator registry not set")
	}
}

func TestValidateAllEmptyRegistry(t *testing.T) {
	r := NewModelRegistry()
	v := NewMetadataValidator(r)
	res := v.ValidateAll()
	if res == nil {
		t.Fatal("ValidateAll() returned nil")
	}
	if !res.Validated {
		t.Fatal("result should be marked validated")
	}
	if len(res.Errors) != 0 {
		t.Fatalf("expected no errors on empty registry, got %d", len(res.Errors))
	}
	if res.Err() != nil {
		t.Fatalf("expected nil error, got %v", res.Err())
	}
}

func TestValidateViewFieldReferences(t *testing.T) {
	r := NewModelRegistry()
	ctx := context.Background()

	manifest := &ModuleManifest{
		Module: "base",
		Models: []ModelDef{
			{Name: "Company", Fields: []FieldDef{{Name: "name", Type: "Text"}}},
		},
		Views: []*ViewDef{
			{Name: "company_summary", Fields: []ViewFieldDef{
				{Name: "company_name", Source: "name"},             // exists
				{Name: "nonexistent", Source: "nonexistent_field"}, // missing
			}},
		},
	}

	if err := r.LoadModule(ctx, "base", manifest); err != nil {
		t.Fatal(err)
	}

	v := NewMetadataValidator(r)
	res := v.checkViewFieldReferences("base", manifest)
	if len(res.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(res.Errors), res.Errors)
	}
	if res.Errors[0].ErrorID != "view_field_not_found" {
		t.Fatalf("unexpected error id: %s", res.Errors[0].ErrorID)
	}
	if res.Err() == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateActionInputFields(t *testing.T) {
	r := NewModelRegistry()
	ctx := context.Background()

	manifest := &ModuleManifest{
		Module: "base",
		Models: []ModelDef{
			{Name: "Company", Fields: []FieldDef{{Name: "name", Type: "Text"}}},
		},
		Actions: []*ActionDef{
			{Name: "archive", Input: &ActionInputDef{
				Fields: map[string]any{
					"confirm":      true, // not in model -> error
					"company_name": "x",  // not in model either -> error
				},
			}},
		},
	}

	if err := r.LoadModule(ctx, "base", manifest); err != nil {
		t.Fatal(err)
	}

	v := NewMetadataValidator(r)
	res := v.checkActionInputFields("base", manifest)
	if len(res.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(res.Errors), res.Errors)
	}
}

func TestValidateSecurityFieldPermissions(t *testing.T) {
	r := NewModelRegistry()
	ctx := context.Background()

	manifest := &ModuleManifest{
		Module: "base",
		Models: []ModelDef{
			{Name: "Company", Fields: []FieldDef{{Name: "name", Type: "Text"}}},
		},
		Security: &ModuleSecurity{
			Module: "base",
			Resources: []ResourceMeta{
				{
					Name: "Company",
					FieldPermissions: []FieldPermission{
						{Field: "name", Effect: EffectAllow},  // valid
						{Field: "salary", Effect: EffectDeny}, // invalid - not in model
					},
				},
			},
		},
	}

	if err := r.LoadModule(ctx, "base", manifest); err != nil {
		t.Fatal(err)
	}

	v := NewMetadataValidator(r)
	res := v.checkSecurityFieldPermissions("base", manifest)
	if len(res.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(res.Errors), res.Errors)
	}
	if res.Errors[0].ErrorID != "security_field_not_in_model" {
		t.Fatalf("unexpected error id: %s", res.Errors[0].ErrorID)
	}
}

func TestValidateRelationTargets(t *testing.T) {
	r := NewModelRegistry()
	ctx := context.Background()

	manifest := &ModuleManifest{
		Module: "base",
		Models: []ModelDef{
			{Name: "Company", Fields: []FieldDef{{Name: "name", Type: "Text"}},
				Relations: []RelationDef{
					{Name: "department", Type: "has_one", Target: "Department"},
					{Name: "missing", Type: "has_one", Target: "NonExistent"},
				}},
			{Name: "Department", Fields: []FieldDef{{Name: "name", Type: "Text"}}},
		},
	}

	if err := r.LoadModule(ctx, "base", manifest); err != nil {
		t.Fatal(err)
	}

	v := NewMetadataValidator(r)
	res := v.checkRelationTargets("base", manifest)
	if len(res.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(res.Errors), res.Errors)
	}
}

func TestValidateModuleOnLoad(t *testing.T) {
	r := NewModelRegistry()

	manifest := &ModuleManifest{
		Module: "base",
		Models: []ModelDef{
			{Name: "Company", Fields: []FieldDef{{Name: "name", Type: "Text"}}},
		},
	}

	if err := ValidateModuleOnLoad(r, "base", manifest); err != nil {
		t.Fatalf("ValidateModuleOnLoad() error = %v", err)
	}

	// Invalid manifest
	badManifest := &ModuleManifest{
		Module: "base",
		Security: &ModuleSecurity{
			Module: "base",
			Resources: []ResourceMeta{
				{Name: "Company", FieldPermissions: []FieldPermission{{Field: "nonexistent", Effect: EffectAllow}}},
			},
		},
	}
	if err := ValidateModuleOnLoad(r, "base", badManifest); err == nil {
		t.Fatal("expected validation error for invalid security field")
	}

	// Nil registry
	if err := ValidateModuleOnLoad(nil, "base", manifest); err == nil {
		t.Fatal("expected error for nil registry")
	}

	// Nil manifest
	if err := ValidateModuleOnLoad(r, "base", nil); err == nil {
		t.Fatal("expected error for nil manifest")
	}
}

func TestValidateAllReport(t *testing.T) {
	r := NewModelRegistry()
	v := NewMetadataValidator(r)
	res := v.ValidateAll()
	report := res.Report()
	if report != "Validation passed: all metadata is consistent" {
		t.Fatalf("unexpected report: %s", report)
	}
}
