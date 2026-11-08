package tenancy_test

import (
	"context"
	"database/sql"
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
	for _, ds := range []tenancy.Datastore{dsA, dsB} {
		_, _ = ds.DB().ExecContext(ctx, `CREATE TABLE items (id TEXT PRIMARY KEY, owner TEXT, secret TEXT)`)
	}
	_, _ = dsA.DB().ExecContext(ctx, `INSERT INTO items VALUES ('r1','A','only-a')`)
	_, _ = dsB.DB().ExecContext(ctx, `INSERT INTO items VALUES ('r1','B','only-b')`)

	var owner, secret string
	_ = dsA.DB().QueryRowContext(ctx, `SELECT owner,secret FROM items WHERE id='r1'`).Scan(&owner, &secret)
	if owner != "A" || secret != "only-a" {
		t.Fatalf("A wrong row: %s %s", owner, secret)
	}
	_ = dsB.DB().QueryRowContext(ctx, `SELECT owner,secret FROM items WHERE id='r1'`).Scan(&owner, &secret)
	if owner != "B" || secret != "only-b" {
		t.Fatalf("B wrong row: %s %s", owner, secret)
	}
	_, _ = dsA.DB().ExecContext(ctx, `UPDATE items SET secret='mutated-a' WHERE id='r1'`)
	_ = dsB.DB().QueryRowContext(ctx, `SELECT secret FROM items WHERE id='r1'`).Scan(&secret)
	if secret != "only-b" {
		t.Fatalf("B corrupted: %s", secret)
	}
}

func TestAdversarial_ForgedTenantSelectors(t *testing.T) {
	root := t.TempDir()
	dirs, _ := config.EnsureBaaSDirectories(root, "")
	provider := tenancy.NewSQLiteProvider(dirs)
	defer provider.Close()
	ctx := context.Background()
	for _, id := range []string{"", "../escape", "sub/path"} {
		if _, err := provider.Open(ctx, id); err == nil {
			t.Fatalf("forged selector accepted: %q", id)
		}
	}
}

func TestAdversarial_StaleCacheEvictionAndReopen(t *testing.T) {
	root := t.TempDir()
	dirs, _ := config.EnsureBaaSDirectories(root, "")
	provider := tenancy.NewSQLiteProvider(dirs)
	defer provider.Close()
	ctx := context.Background()
	ds1, _ := provider.Open(ctx, "evict-me")
	_, _ = ds1.DB().ExecContext(ctx, `CREATE TABLE marker (v TEXT)`)
	_, _ = ds1.DB().ExecContext(ctx, `INSERT INTO marker (v) VALUES ('before')`)
	if err := provider.CloseTenant(ctx, "evict-me"); err != nil {
		t.Fatal(err)
	}
	ds2, _ := provider.Open(ctx, "evict-me")
	if ds1 == ds2 {
		t.Fatal("eviction did not return fresh handle")
	}
	var v string
	_ = ds2.DB().QueryRowContext(ctx, `SELECT v FROM marker`).Scan(&v)
	if v != "before" {
		t.Fatalf("data lost after reopen: %s", v)
	}
}

func TestAdversarial_SuspendedAndDeletingFailClosed(t *testing.T) {
	for _, st := range []tenancy.Status{tenancy.StatusSuspended, tenancy.StatusDeleting} {
		t.Run(string(st), func(t *testing.T) {
			root := t.TempDir()
			dirs, _ := config.EnsureBaaSDirectories(root, "")
			provider := tenancy.NewSQLiteProvider(dirs)
			defer provider.Close()
			tenant := tenancy.Tenant{ID: "t-" + string(st), Slug: "s", Status: st}
			r := tenancy.NewRouter(&fakeCatalog{tenant: tenant}, provider)
			ctx := tenancy.NewContext(context.Background(), tenant)
			_, err := r.Datastore(ctx)
			if err == nil {
				t.Fatal("expected error for non-active status")
			}
		})
	}
}

var _ = sql.Open
