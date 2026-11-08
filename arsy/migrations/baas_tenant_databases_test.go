package migrations_test

import (
	"testing"

	_ "github.com/bridge-allah/a-r-s-y/arsy/migrations"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestBaaSTenantDatabasesCollection(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	col, err := app.FindCollectionByNameOrId("baas_tenant_databases")
	if err != nil {
		t.Fatalf("baas_tenant_databases missing: %v", err)
	}
	for _, name := range []string{"tenant", "engine", "database_id", "state", "secret_ref", "is_primary", "placement_key", "schema_version"} {
		if col.Fields.GetByName(name) == nil {
			t.Fatalf("missing field %q", name)
		}
	}
	if !containsMigrationIndex(col.Indexes, "idx_baas_tenant_databases_database_id") {
		t.Fatal("missing database_id unique index")
	}
}
