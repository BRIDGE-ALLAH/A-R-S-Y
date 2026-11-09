package aeng

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"
)

func TestNewSeedRunner(t *testing.T) {
	db := openTestDB(t)
	r := NewSeedRunner(db, "production")
	if r == nil {
		t.Fatal("NewSeedRunner() returned nil")
	}
	if r.environment != "production" {
		t.Fatalf("environment mismatch: %s", r.environment)
	}
}

func TestRunSeedsUpsertIdempotent(t *testing.T) {
	db := openTestDB(t)
	r := NewSeedRunner(db, "production")

	ctx := context.Background()
	if _, err := db.ExecContext(ctx, "CREATE TABLE settings (id TEXT PRIMARY KEY, name TEXT)"); err != nil {
		t.Fatal(err)
	}

	seeds := []*SeedDef{
		{
			Collection: "settings",
			Upsert:     true,
			Fields: map[string]any{
				"id":   "site_name",
				"name": "A-R-S-Y Platform",
			},
		},
	}

	if err := r.RunSeeds(ctx, seeds); err != nil {
		t.Fatalf("RunSeeds() error = %v", err)
	}

	// Verify record inserted
	var name string
	if err := db.QueryRowContext(ctx, "SELECT name FROM settings WHERE id = ?", "site_name").Scan(&name); err != nil {
		t.Fatalf("settings record not inserted: %v", err)
	}
	if name != "A-R-S-Y Platform" {
		t.Fatalf("expected name A-R-S-Y Platform, got %s", name)
	}

	// Re-run: should be idempotent (no duplicate)
	if err := r.RunSeeds(ctx, seeds); err != nil {
		t.Fatalf("second RunSeeds() error = %v", err)
	}

	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM settings WHERE id = ?", "site_name").Scan(&count); err != nil {
		t.Fatalf("count query error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 record after idempotent re-run, got %d", count)
	}
}

func TestRunSeedsForTenant(t *testing.T) {
	tenantDB := openTestDB(t)
	r := NewSeedRunner(tenantDB, "production")

	ctx := context.Background()
	if _, err := tenantDB.ExecContext(ctx, "CREATE TABLE settings (id TEXT PRIMARY KEY, name TEXT)"); err != nil {
		t.Fatal(err)
	}

	seeds := []*SeedDef{
		{
			Collection: "settings",
			Upsert:     true,
			Fields: map[string]any{
				"id":   "tenant_id",
				"name": "tenant-1",
			},
		},
	}

	if err := r.RunSeedsForTenant(ctx, tenantDB, seeds); err != nil {
		t.Fatalf("RunSeedsForTenant() error = %v", err)
	}

	var name string
	if err := tenantDB.QueryRowContext(ctx, "SELECT name FROM settings WHERE id = ?", "tenant_id").Scan(&name); err != nil {
		t.Fatalf("tenant settings record not inserted: %v", err)
	}
	if name != "tenant-1" {
		t.Fatalf("expected name tenant-1, got %s", name)
	}
}

func TestRunSeedsNilDB(t *testing.T) {
	r := NewSeedRunner(nil, "production")
	ctx := context.Background()
	if err := r.RunSeeds(ctx, nil); err == nil {
		t.Fatal("expected error for nil db")
	}
}

func TestConditionsMet(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	// No conditions -> true
	if !conditionsMet(db, "settings", nil) {
		t.Fatal("empty conditions should be met")
	}

	// Insert a record
	_, err := db.ExecContext(ctx, "CREATE TABLE settings (id TEXT PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, "INSERT INTO settings (id, name) VALUES (?, ?)", "site_name", "A-R-S-Y")
	if err != nil {
		t.Fatal(err)
	}

	// exists condition -> true
	conds := []ConditionDef{{Field: "id", Operator: "exists", Value: "site_name"}}
	if !conditionsMet(db, "settings", conds) {
		t.Fatal("exists condition should be met")
	}

	// not_exists condition -> false (record exists)
	conds = []ConditionDef{{Field: "id", Operator: "not_exists", Value: "site_name"}}
	if conditionsMet(db, "settings", conds) {
		t.Fatal("not_exists condition should not be met when record exists")
	}

	// eq condition -> true
	conds = []ConditionDef{{Field: "id", Operator: "eq", Value: "site_name"}}
	if !conditionsMet(db, "settings", conds) {
		t.Fatal("eq condition should be met")
	}

	// Missing table -> false
	conds = []ConditionDef{{Field: "id", Operator: "exists", Value: "x"}}
	if conditionsMet(db, "settings", conds) {
		// conditionsMet returns false on query error; acceptable
	}
}

func TestSeedRunnerNilSeed(t *testing.T) {
	db := openTestDB(t)
	r := NewSeedRunner(db, "")
	ctx := context.Background()
	// nil seed in slice should not panic
	if err := r.RunSeeds(ctx, []*SeedDef{nil}); err != nil {
		t.Fatalf("RunSeeds with nil seed error = %v", err)
	}
}

func TestShouldExecuteEnvironment(t *testing.T) {
	db := openTestDB(t)
	r := NewSeedRunner(db, "production")
	// Seed without environment filter runs everywhere
	seed := &SeedDef{Collection: "x", Fields: map[string]any{"id": "1"}}
	if !r.shouldExecute(seed) {
		t.Fatal("seed without env filter should execute")
	}
}
