package bootstrap

import (
	"context"
	"database/sql"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	_ "modernc.org/sqlite"
)

func TestDefaultRegistryAndApplySchema(t *testing.T) {
	reg, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.Snapshot()) == 0 {
		t.Fatal("expected starter models")
	}
	dirs, err := config.EnsureBaaSDirectories(t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	// Create an empty DB file path target by applying schema (sqlite creates file).
	if err := ApplySchema(ctx, dirs, "tenant1", reg); err != nil {
		t.Fatal(err)
	}
	dbPath, _ := dirs.TenantDBPath("tenant1")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`INSERT INTO base_company (id,name,code,active) VALUES ('c1','Acme','ACME','true')`); err != nil {
		t.Fatal(err)
	}
	// Idempotent reapply.
	if err := ApplySchema(ctx, dirs, "tenant1", reg); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM base_company WHERE id='c1'`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("row lost after idempotent schema apply: n=%d err=%v", n, err)
	}
}
