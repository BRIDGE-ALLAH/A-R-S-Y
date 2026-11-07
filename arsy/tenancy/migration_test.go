 package tenancy_test

 import (
 	"context"
 	"database/sql"
 	"testing"

 	"github.com/bridge-allah/a-r-s-y/arsy/config"
 	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
 )

 func sampleMigrations() []tenancy.Migration {
 	return []tenancy.Migration{
 		{
 			Version: "00000001",
 			Up: func(db *sql.DB) error {
 				_, err := db.Exec(`CREATE TABLE IF NOT EXISTS test_items (id TEXT PRIMARY KEY, name TEXT)`)
 				return err
 			},
 			Down: func(db *sql.DB) error {
 				_, err := db.Exec(`DROP TABLE IF EXISTS test_items`)
 				return err
 			},
 		},
 	}
 }

 func TestMigrationRunner_Apply(t *testing.T) {
 	root := t.TempDir()
 	dirs, _ := config.EnsureBaaSDirectories(root, "")
 	provider := tenancy.NewSQLiteProvider(dirs)
 	defer provider.Close()

 	ctx := context.Background()
 	ds, _ := provider.Open(ctx, "migtest")

 	runner := tenancy.NewMigrationRunner(sampleMigrations())
 	if err := runner.Migrate(ctx, ds); err != nil {
 		t.Fatalf("Migrate: %v", err)
 	}

 	// Idempotent: second run must not error.
 	if err := runner.Migrate(ctx, ds); err != nil {
 		t.Fatalf("second Migrate: %v", err)
 	}

 	versions, err := runner.AppliedVersions(ctx, ds)
 	if err != nil {
 		t.Fatalf("AppliedVersions: %v", err)
 	}
 	if len(versions) != 1 || versions[0] != "00000001" {
 		t.Errorf("expected [00000001], got %v", versions)
 	}
 }

 func TestProvisioner_Idempotent(t *testing.T) {
 	root := t.TempDir()
 	dirs, _ := config.EnsureBaaSDirectories(root, "")
 	provider := tenancy.NewSQLiteProvider(dirs)
 	defer provider.Close()

 	p := tenancy.NewProvisioner(provider, sampleMigrations())
 	ctx := context.Background()

 	if err := p.Provision(ctx, "prov1"); err != nil {
 		t.Fatalf("first Provision: %v", err)
 	}
 	if err := p.Provision(ctx, "prov1"); err != nil {
 		t.Fatalf("second Provision (idempotent): %v", err)
 	}
 }

 func TestProvisioner_EmptyIDRejected(t *testing.T) {
 	root := t.TempDir()
 	dirs, _ := config.EnsureBaaSDirectories(root, "")
 	provider := tenancy.NewSQLiteProvider(dirs)
 	defer provider.Close()

 	p := tenancy.NewProvisioner(provider, nil)
 	if err := p.Provision(context.Background(), ""); err == nil {
 		t.Fatal("expected error for empty tenantID")
 	}
 }
