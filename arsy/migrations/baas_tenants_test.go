 package migrations_test

 import (
 	"testing"

 	_ "github.com/bridge-allah/a-r-s-y/arsy/migrations"
 	"github.com/bridge-allah/a-r-s-y/core"
 	"github.com/bridge-allah/a-r-s-y/tests"
 )

 func TestBaaSTenantsCollectionCreated(t *testing.T) {
 	app, err := tests.NewTestApp()
 	if err != nil {
 		t.Fatal(err)
 	}
 	defer app.Cleanup()

 	col, err := app.FindCollectionByNameOrId("baas_tenants")
 	if err != nil {
 		t.Fatalf("baas_tenants collection was not created: %v", err)
 	}

 	// Must be superuser-only (no public rules).
 	if col.ListRule != nil || col.ViewRule != nil || col.CreateRule != nil || col.UpdateRule != nil || col.DeleteRule != nil {
 		t.Fatal("baas_tenants must be superuser-only (all rules nil)")
 	}

 	fields := col.Fields.AsMap()

 	if _, ok := fields["slug"].(*core.TextField); !ok {
 		t.Error("baas_tenants.slug must be a text field")
 	}
 	if _, ok := fields["status"].(*core.TextField); !ok {
 		t.Error("baas_tenants.status must be a text field")
 	}
 	if _, ok := fields["display_name"].(*core.TextField); !ok {
 		t.Error("baas_tenants.display_name must be a text field")
 	}

 	// Unique slug index must exist.
 	if !containsMigrationIndex(col.Indexes, "idx_baas_tenants_slug") {
 		t.Error("baas_tenants must have a unique slug index")
 	}
 }

 // TestMVP1CollectionsUnchanged ensures BAAS migration does not break MVP-1.
 func TestMVP1CollectionsUnchanged(t *testing.T) {
 	app, err := tests.NewTestApp()
 	if err != nil {
 		t.Fatal(err)
 	}
 	defer app.Cleanup()

 	for _, name := range []string{
 		"arsy_projects", "arsy_servers", "arsy_applications",
 		"arsy_deployments", "arsy_deployment_logs",
 	} {
 		if _, err := app.FindCollectionByNameOrId(name); err != nil {
 			t.Fatalf("MVP-1 collection %q missing after BAAS migration: %v", name, err)
 		}
 	}
 }
