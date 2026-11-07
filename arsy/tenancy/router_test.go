 package tenancy_test

 import (
 	"context"
 	"testing"

 	"github.com/bridge-allah/a-r-s-y/arsy/config"
 	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
 )

 // fakeCatalog is a stub Catalog for tests.
 type fakeCatalog struct{ tenant tenancy.Tenant }

 func (f *fakeCatalog) Get(_ context.Context, id string) (tenancy.Tenant, error) {
 	if id == f.tenant.ID {
 		return f.tenant, nil
 	}
 	return tenancy.Tenant{}, tenancy.ErrTenantNotFound
 }
 func (f *fakeCatalog) GetBySlug(_ context.Context, slug string) (tenancy.Tenant, error) {
 	if slug == f.tenant.Slug {
 		return f.tenant, nil
 	}
 	return tenancy.Tenant{}, tenancy.ErrTenantNotFound
 }

 func newTestRouter(t *testing.T, tenant tenancy.Tenant) (*tenancy.Router, *tenancy.SQLiteProvider) {
 	t.Helper()
 	root := t.TempDir()
 	dirs, _ := config.EnsureBaaSDirectories(root, "")
 	provider := tenancy.NewSQLiteProvider(dirs)
 	catalog := &fakeCatalog{tenant: tenant}
 	return tenancy.NewRouter(catalog, provider), provider
 }

 func TestRouter_ActiveTenantReturnsDatastore(t *testing.T) {
 	tenant := tenancy.Tenant{ID: "t1", Slug: "acme", Status: tenancy.StatusActive}
 	router, provider := newTestRouter(t, tenant)
 	defer provider.Close()

 	ctx := tenancy.NewContext(context.Background(), tenant)
 	ds, err := router.Datastore(ctx)
 	if err != nil {
 		t.Fatalf("expected datastore, got %v", err)
 	}
 	if ds.TenantID() != "t1" {
 		t.Errorf("wrong tenant ID: %q", ds.TenantID())
 	}
 }

 func TestRouter_SuspendedRejectsAccess(t *testing.T) {
 	tenant := tenancy.Tenant{ID: "t2", Slug: "suspended-co", Status: tenancy.StatusSuspended}
 	router, provider := newTestRouter(t, tenant)
 	defer provider.Close()

 	ctx := tenancy.NewContext(context.Background(), tenant)
 	_, err := router.Datastore(ctx)
 	if err != tenancy.ErrTenantSuspended {
 		t.Errorf("expected ErrTenantSuspended, got %v", err)
 	}
 }

 func TestRouter_DeletingRejectsAccess(t *testing.T) {
 	tenant := tenancy.Tenant{ID: "t3", Slug: "gone", Status: tenancy.StatusDeleting}
 	router, provider := newTestRouter(t, tenant)
 	defer provider.Close()

 	ctx := tenancy.NewContext(context.Background(), tenant)
 	_, err := router.Datastore(ctx)
 	if err != tenancy.ErrTenantDeleting {
 		t.Errorf("expected ErrTenantDeleting, got %v", err)
 	}
 }

 func TestRouter_NoTenantInContext(t *testing.T) {
 	tenant := tenancy.Tenant{ID: "t4", Status: tenancy.StatusActive}
 	router, provider := newTestRouter(t, tenant)
 	defer provider.Close()

 	_, err := router.Datastore(context.Background())
 	if err == nil {
 		t.Fatal("expected error when no tenant in context")
 	}
 }
