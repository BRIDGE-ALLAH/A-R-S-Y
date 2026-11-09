package aeng

import (
	"context"
	"testing"
)

func TestNewModelRegistry(t *testing.T) {
	r := NewModelRegistry()
	if r == nil {
		t.Fatal("NewModelRegistry() returned nil")
	}
	if r.modules == nil || r.modelByName == nil || r.actionByName == nil || r.securityByModule == nil {
		t.Fatal("registry maps not initialized")
	}
}

func TestLoadModuleAndGet(t *testing.T) {
	r := NewModelRegistry()
	manifest := &ModuleManifest{
		Module:  "base",
		Version: "v1.0.0",
		Models: []ModelDef{
			{
				Name: "Company",
				Fields: []FieldDef{
					{Name: "id", Type: "Text", Required: true, Unique: true},
					{Name: "name", Type: "Text", Required: true},
				},
			},
		},
		Actions: []*ActionDef{
			{
				Name: "archive",
				Access: &AccessControlDef{
					Roles: []string{"admin"},
				},
			},
		},
		Security: &ModuleSecurity{
			Module:  "base",
			Version: "v1.0.0",
			Roles:   []RoleMeta{{Name: "admin", Description: "admin"}},
		},
	}

	ctx := context.Background()
	if err := r.LoadModule(ctx, "base", manifest); err != nil {
		t.Fatalf("LoadModule() error = %v", err)
	}

	mod, ok := r.GetModule("base")
	if !ok {
		t.Fatal("GetModule(base) not found")
	}
	if mod.ID != "base" || mod.Version != "v1.0.0" {
		t.Fatalf("module info mismatch: %+v", mod)
	}
	if !mod.IsLoaded {
		t.Fatal("module not marked loaded")
	}

	model, ok := r.GetModel("Company")
	if !ok {
		t.Fatal("GetModel(Company) not found")
	}
	if len(model.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(model.Fields))
	}

	action, ok := r.GetAction("archive")
	if !ok {
		t.Fatal("GetAction(archive) not found")
	}
	if action.ActionDef.Name != "archive" {
		t.Fatalf("action name mismatch: %s", action.ActionDef.Name)
	}

	sec, ok := r.GetSecurity("base")
	if !ok {
		t.Fatal("GetSecurity(base) not found")
	}
	if sec.Module != "base" {
		t.Fatalf("security module mismatch: %s", sec.Module)
	}

	// Duplicate load should error
	if err := r.LoadModule(ctx, "base", manifest); err == nil {
		t.Fatal("expected error on duplicate module load")
	}
}

func TestLoadModuleNilManifest(t *testing.T) {
	r := NewModelRegistry()
	ctx := context.Background()
	if err := r.LoadModule(ctx, "base", nil); err == nil {
		t.Fatal("expected error for nil manifest")
	}
}

func TestResolveDependencies(t *testing.T) {
	r := NewModelRegistry()
	ctx := context.Background()

	// base has no deps
	if err := r.LoadModule(ctx, "base", &ModuleManifest{Module: "base", Version: "v1.0.0"}); err != nil {
		t.Fatal(err)
	}
	// crm depends on base
	if err := r.LoadModule(ctx, "crm", &ModuleManifest{Module: "crm", Version: "v1.0.0"}); err != nil {
		t.Fatal(err)
	}
	if err := r.LoadModule(ctx, "billing", &ModuleManifest{Module: "billing", Version: "v1.0.0"}); err != nil {
		t.Fatal(err)
	}
	// Set dependencies on ModuleInfo directly
	r.modules["crm"].Dependencies = []string{"base"}
	r.modules["billing"].Dependencies = []string{"base", "crm"}

	if err := r.ResolveDependencies(); err != nil {
		t.Fatalf("ResolveDependencies() error = %v", err)
	}

	order := r.GetLoadOrder()
	if len(order) != 3 {
		t.Fatalf("expected 3 modules in load order, got %d: %v", len(order), order)
	}
	// base must come first
	if order[0] != "base" {
		t.Fatalf("base should be first, got %s", order[0])
	}
	// crm before billing (crm is a dependency of billing)
	crmIdx := -1
	billingIdx := -1
	for i, id := range order {
		if id == "crm" {
			crmIdx = i
		}
		if id == "billing" {
			billingIdx = i
		}
	}
	if crmIdx >= billingIdx {
		t.Fatalf("crm should come before billing: %v", order)
	}
}

func TestResolveDependenciesCircular(t *testing.T) {
	r := NewModelRegistry()
	ctx := context.Background()

	if err := r.LoadModule(ctx, "a", &ModuleManifest{Module: "a", Version: "v1.0.0"}); err != nil {
		t.Fatal(err)
	}
	if err := r.LoadModule(ctx, "b", &ModuleManifest{Module: "b", Version: "v1.0.0"}); err != nil {
		t.Fatal(err)
	}
	r.modules["a"].Dependencies = []string{"b"}
	r.modules["b"].Dependencies = []string{"a"}

	if err := r.ResolveDependencies(); err == nil {
		t.Fatal("expected circular dependency error")
	}
}

func TestResolveDependenciesMissing(t *testing.T) {
	r := NewModelRegistry()
	ctx := context.Background()

	if err := r.LoadModule(ctx, "a", &ModuleManifest{Module: "a", Version: "v1.0.0"}); err != nil {
		t.Fatal(err)
	}
	r.modules["a"].Dependencies = []string{"missing"}

	if err := r.ResolveDependencies(); err == nil {
		t.Fatal("expected missing module error")
	}
}

func TestGoTypeFromPB(t *testing.T) {
	tests := []struct {
		pbType string
		want   string
	}{
		{"Text", "string"},
		{"Editor", "string"},
		{"URL", "string"},
		{"Email", "string"},
		{"Number", "float64"},
		{"Bool", "bool"},
		{"Select", "string"},
		{"Date", "time.Time"},
		{"DateTime", "time.Time"},
		{"Autodate", "time.Time"},
		{"File", "[]string"},
		{"JSON", "map[string]any"},
		{"Relation", "string"},
		{"Unknown", "any"},
	}
	for _, tt := range tests {
		if got := goTypeFromPB(tt.pbType); got != tt.want {
			t.Errorf("goTypeFromPB(%q) = %q, want %q", tt.pbType, got, tt.want)
		}
	}
}

func TestValidateSecurityMeta(t *testing.T) {
	r := NewModelRegistry()
	ctx := context.Background()

	manifest := &ModuleManifest{
		Module:  "base",
		Version: "v1.0.0",
		Models: []ModelDef{
			{Name: "Company", Fields: []FieldDef{{Name: "name", Type: "Text"}, {Name: "salary", Type: "Number"}}},
		},
		Security: &ModuleSecurity{
			Module:  "base",
			Version: "v1.0.0",
			Resources: []ResourceMeta{
				{
					Name: "Company",
					FieldPermissions: []FieldPermission{
						{Field: "name", Effect: EffectAllow, Roles: []string{"admin"}},
						{Field: "salary", Effect: EffectDeny, Roles: []string{"viewer"}},
					},
				},
			},
		},
	}

	if err := r.LoadModule(ctx, "base", manifest); err != nil {
		t.Fatal(err)
	}

	if err := ValidateSecurityMeta(r); err != nil {
		t.Fatalf("ValidateSecurityMeta() error = %v", err)
	}

	// Invalid: field permission references unknown model
	badSec := &ModuleSecurity{
		Module: "base",
		Resources: []ResourceMeta{
			{Name: "UnknownModel", FieldPermissions: []FieldPermission{{Field: "x", Effect: EffectAllow}}},
		},
	}
	r2 := NewModelRegistry()
	if err := r2.LoadModule(ctx, "base", &ModuleManifest{Module: "base"}); err != nil {
		t.Fatal(err)
	}
	r2.securityByModule["base"] = badSec
	if err := ValidateSecurityMeta(r2); err == nil {
		t.Fatal("expected error for security resource referencing unknown model")
	}

	// Invalid: field permission references unknown field
	badSec2 := &ModuleSecurity{
		Module: "base",
		Resources: []ResourceMeta{
			{Name: "Company", FieldPermissions: []FieldPermission{{Field: "nonexistent", Effect: EffectAllow}}},
		},
	}
	r3 := NewModelRegistry()
	if err := r3.LoadModule(ctx, "base", manifest); err != nil {
		t.Fatal(err)
	}
	r3.securityByModule["base"] = badSec2
	if err := ValidateSecurityMeta(r3); err == nil {
		t.Fatal("expected error for security field not found in model")
	}
}

func TestSecurityEnforcer(t *testing.T) {
	r := NewModelRegistry()
	enforcer := NewSecurityEnforcer(r)

	resource := &ResourceMeta{
		Name: "Company",
		FieldPermissions: []FieldPermission{
			{Field: "salary", Effect: EffectDeny, Roles: []string{"viewer"}},
			{Field: "email", Effect: EffectAllow, Roles: []string{"admin", "viewer"}},
		},
		Actions: []string{"read", "write"},
	}

	// Superuser bypass
	allowed, reason := enforcer.CheckPermission(context.Background(), &superuser{}, "read", resource)
	if !allowed {
		t.Errorf("superuser should be allowed, reason: %s", reason)
	}

	// Field deny
	allowed, reason = enforcer.CheckPermission(context.Background(), &viewer{}, "salary", resource)
	if allowed {
		t.Errorf("viewer should be denied salary field, reason: %s", reason)
	}

	// Field allow
	allowed, reason = enforcer.CheckPermission(context.Background(), &viewer{}, "email", resource)
	if !allowed {
		t.Errorf("viewer should be allowed email field, reason: %s", reason)
	}

	// Nil resource
	allowed, reason = enforcer.CheckPermission(context.Background(), &viewer{}, "read", nil)
	if allowed {
		t.Fatal("nil resource should be denied")
	}
}

type superuser struct{}

func (s *superuser) IsSuperuser() bool { return true }

type viewer struct{}

func (v *viewer) IsSuperuser() bool { return false }

func TestSecurityEnforcerNilRegistry(t *testing.T) {
	enforcer := NewSecurityEnforcer(nil)
	allowed, reason := enforcer.CheckActionPermission(context.Background(), &superuser{}, "archive")
	if allowed {
		t.Fatal("nil registry should deny action permission")
	}
	if reason != "registry not initialized" {
		t.Fatalf("unexpected reason: %s", reason)
	}
}
