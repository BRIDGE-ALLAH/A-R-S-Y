 package tenancy_test

 import (
 	"context"
 	"testing"

 	"github.com/bridge-allah/a-r-s-y/arsy/config"
 	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
 )

 func newTestProvider(t *testing.T) *tenancy.SQLiteProvider {
 	t.Helper()
 	root := t.TempDir()
 	dirs, err := config.EnsureBaaSDirectories(root, "")
 	if err != nil {
 		t.Fatalf("setup dirs: %v", err)
 	}
 	return tenancy.NewSQLiteProvider(dirs)
 }

 func TestSQLiteProvider_OpenAndPing(t *testing.T) {
 	p := newTestProvider(t)
 	defer p.Close()

 	ctx := context.Background()
 	ds, err := p.Open(ctx, "tenant01")
 	if err != nil {
 		t.Fatalf("Open: %v", err)
 	}
 	if ds.TenantID() != "tenant01" {
 		t.Errorf("expected tenant01, got %q", ds.TenantID())
 	}
 	if err := ds.Ping(ctx); err != nil {
 		t.Errorf("Ping: %v", err)
 	}
 }

 func TestSQLiteProvider_Idempotent(t *testing.T) {
 	p := newTestProvider(t)
 	defer p.Close()

 	ctx := context.Background()
 	ds1, _ := p.Open(ctx, "tenant02")
 	ds2, _ := p.Open(ctx, "tenant02")
 	if ds1 != ds2 {
 		t.Error("Open must return the same Datastore for the same tenant")
 	}
 }

 func TestSQLiteProvider_MultipleTenants(t *testing.T) {
 	p := newTestProvider(t)
 	defer p.Close()

 	ctx := context.Background()
 	ds1, _ := p.Open(ctx, "tenantA")
 	ds2, _ := p.Open(ctx, "tenantB")
 	if ds1.DB() == ds2.DB() {
 		t.Error("different tenants must have different DB handles")
 	}
 }

 func TestSQLiteProvider_ClosedRejectsOpen(t *testing.T) {
 	p := newTestProvider(t)
 	p.Close()

 	_, err := p.Open(context.Background(), "tenant03")
 	if err == nil {
 		t.Fatal("expected error after provider Close, got nil")
 	}
 }

 func TestSQLiteProvider_InvalidTenantID(t *testing.T) {
 	p := newTestProvider(t)
 	defer p.Close()

 	ctx := context.Background()
 	for _, bad := range []string{"", "../escape", "sub/path"} {
 		if _, err := p.Open(ctx, bad); err == nil {
 			t.Errorf("expected error for tenant ID %q, got nil", bad)
 		}
 	}
 }

 func TestSQLiteProvider_Health(t *testing.T) {
 	p := newTestProvider(t)
 	defer p.Close()

 	status := p.Health(context.Background(), "healthcheck")
 	if !status.Healthy {
 		t.Errorf("expected healthy, got err: %v", status.Err)
 	}
 }
