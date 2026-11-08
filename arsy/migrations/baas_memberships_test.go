package migrations_test

import (
	"testing"

	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestBaaSTenantMembershipsCollectionCreated(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	col, err := app.FindCollectionByNameOrId("baas_tenant_memberships")
	if err != nil {
		t.Fatalf("baas_tenant_memberships collection missing: %v", err)
	}
	if col.ListRule != nil || col.ViewRule != nil || col.CreateRule != nil || col.UpdateRule != nil || col.DeleteRule != nil {
		t.Fatal("baas_tenant_memberships must be control-plane only")
	}
	fields := col.Fields.AsMap()
	for _, name := range []string{"tenant", "principal", "role", "status", "granted_by"} {
		if _, ok := fields[name]; !ok {
			t.Fatalf("missing field %q", name)
		}
	}
	if _, ok := fields["principal"].(*core.TextField); !ok {
		t.Fatal("principal must be a text field")
	}
	if _, ok := fields["role"].(*core.TextField); !ok {
		t.Fatal("role must be a text field")
	}
	if _, ok := fields["status"].(*core.TextField); !ok {
		t.Fatal("status must be a text field")
	}
	if !containsMigrationIndex(col.Indexes, "idx_baas_tenant_memberships_tenant_principal") {
		t.Fatal("missing unique tenant+principal index")
	}
}

func TestBaaSTenantMembershipValidation(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	tenantCol, _ := app.FindCollectionByNameOrId("baas_tenants")
	tenant := core.NewRecord(tenantCol)
	tenant.Set("slug", "memtest")
	tenant.Set("display_name", "Mem Test")
	tenant.Set("status", "active")
	if err := app.Save(tenant); err != nil {
		t.Fatal(err)
	}

	mCol, _ := app.FindCollectionByNameOrId("baas_tenant_memberships")
	ok := core.NewRecord(mCol)
	ok.Set("tenant", tenant.Id)
	ok.Set("principal", "user1abc")
	ok.Set("role", "developer")
	ok.Set("status", "active")
	if err := app.Save(ok); err != nil {
		t.Fatalf("valid membership rejected: %v", err)
	}

	badRole := core.NewRecord(mCol)
	badRole.Set("tenant", tenant.Id)
	badRole.Set("principal", "user2abc")
	badRole.Set("role", "hacker")
	badRole.Set("status", "active")
	if err := app.Save(badRole); err == nil {
		t.Fatal("expected invalid role to be rejected")
	}
}

