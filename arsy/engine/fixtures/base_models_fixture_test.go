package fixtures

import (
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/engine/fields"
	"github.com/bridge-allah/a-r-s-y/arsy/engine/manifest"
	"github.com/bridge-allah/a-r-s-y/arsy/engine/registry"
)

func TestBaseModelsFixture(t *testing.T) {
	// Validate manifest
	m, err := manifest.LoadModuleManifest("base_models")
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if m.Name != "base" {
		t.Fatalf("expected module name base, got %s", m.Name)
	}

	// Test deterministic registration
	r := registry.New()
	err = r.Register(registry.Descriptor{
		ID:          "base.company",
		Label:       "Company",
		TableName:   "base_company",
		Version:     "1.0.0",
		OwnerModule: "base",
		Source:      registry.SourceMetadata,
	})
	if err != nil {
		t.Fatalf("register model: %v", err)
	}

	// Deterministic lookup
	d1, err := r.Lookup("base.company")
	if err != nil {
		t.Fatalf("lookup model: %v", err)
	}
	d2, err := r.Lookup("base.company")
	if err != nil {
		t.Fatalf("lookup model again: %v", err)
	}
	if d1.ID != d2.ID || d1.Label != d2.Label {
		t.Fatalf("non-deterministic registration")
	}

	// Field validation
	fieldsList, err := fields.Validate([]fields.Descriptor{
		{Name: "name", Label: "Name", Type: fields.TypeText},
		{Name: "code", Label: "Code", Type: fields.TypeText},
		{Name: "active", Label: "Active", Type: fields.TypeBoolean},
	}, fields.SourceMetadata)
	if err != nil {
		t.Fatalf("validate fields: %v", err)
	}
	if len(fieldsList) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(fieldsList))
	}

	// No relations in this fixture
	// No real DB seeding - just in-memory
	_ = m
}
