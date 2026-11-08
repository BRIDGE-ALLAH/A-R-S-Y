package tenancy_test

import (
	"context"
	
	
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
)

func TestAdversarial_TwoRealTenantDatabasesIsolation(t *testing.T) {
	root := t.TempDir()
	dirs, err := config.EnsureBaaSDirectories(root, "")
	if err != nil {
		t.Fatal(err)
	}
	provider := tenancy.NewSQLiteProvider(dirs)
	defer provider.Close()
	ctx := context.Background()

	dsA, _ := provider.Open(ctx, "tenant-a-iso")
	dsB, _ := provider.Open(ctx, "tenant-b-iso")
	if dsA.DB() == dsB.DB() {
		t.Fatal("shared handle")
	}
	// Debug: print DB paths
	pathA, _ := dirs.TenantDBPath("tenant-a-iso")
	pathB, _ := dirs.TenantDBPath("tenant-b-iso")
	t.Logf("DB paths: A=%s B=%s", pathA, pathB)

	for _, ds := range []tenancy.Datastore{dsA, dsB} {
		_, _ = ds.DB().ExecContext(ctx, `CREATE TABLE items (id TEXT PRIMARY KEY, owner TEXT, secret TEXT)`)
	}
	_, _ = dsA.DB().ExecContext(ctx, `INSERT INTO items VALUES ('r1','A','only-a')`)
	_, _ = dsB.DB().ExecContext(ctx, `INSERT INTO items VALUES ('r1','B','only-b')`)

	var owner, secret string
	_ = dsA.DB().QueryRowContext(ctx, `SELECT owner,secret FROM items WHERE id='r1'`).Scan(&owner, &secret)
	t.Logf("A read: owner=%q secret=%q", owner, secret)
	if owner != "A" || secret != "only-a" {
		t.Fatalf("A wrong row: %s %s", owner, secret)
	}
	_ = dsB.DB().QueryRowContext(ctx, `SELECT owner,secret FROM items WHERE id='r1'`).Scan(&owner, &secret)
	t.Logf("B read: owner=%q secret=%q", owner, secret)
	if owner != "B" || secret != "only-b" {
		t.Fatalf("B wrong row: %s %s", owner, secret)
	}

	// Write isolation
	_, _ = dsA.DB().ExecContext(ctx, `UPDATE items SET secret='mutated-a' WHERE id='r1'`)
	_ = dsB.DB().QueryRowContext(ctx, `SELECT secret FROM items WHERE id='r1'`).Scan(&secret)
	if secret != "only-b" {
		t.Fatalf("B corrupted: %s", secret)
	}
}
