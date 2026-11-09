package aeng

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestNewMigrationExecutor(t *testing.T) {
	db := openTestDB(t)
	e := NewMigrationExecutor(db)
	if e == nil {
		t.Fatal("NewMigrationExecutor() returned nil")
	}
}

func TestExecutePlanCreatesTrackingTable(t *testing.T) {
	db := openTestDB(t)
	e := NewMigrationExecutor(db)

	plan := &MigrationPlan{
		Module:  "base",
		Version: "v1.0.0",
		Migrations: []Migration{
			{
				ID:          "20261109-001",
				Version:     "001",
				Description: "Create companies table",
				Up: func(ctx context.Context, db Execer) error {
					_, err := db.ExecContext(ctx, "CREATE TABLE companies (id TEXT PRIMARY KEY, name TEXT)")
					return err
				},
				Down: func(ctx context.Context, db Execer) error {
					_, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS companies")
					return err
				},
			},
		},
	}

	ctx := context.Background()
	if err := e.ExecutePlan(ctx, plan); err != nil {
		t.Fatalf("ExecutePlan() error = %v", err)
	}

	// Verify migration applied
	exists, err := e.isMigrationApplied(ctx, "base:20261109-001")
	if err != nil {
		t.Fatalf("isMigrationApplied() error = %v", err)
	}
	if !exists {
		t.Fatal("migration not recorded as applied")
	}

	// Verify table created
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM companies").Scan(&count); err != nil {
		t.Fatalf("companies table not created: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected empty companies table, got %d", count)
	}
}

func TestExecutePlanIdempotent(t *testing.T) {
	db := openTestDB(t)
	e := NewMigrationExecutor(db)

	applied := 0
	plan := &MigrationPlan{
		Module:  "base",
		Version: "v1.0.0",
		Migrations: []Migration{
			{
				ID: "20261109-001",
				Up: func(ctx context.Context, db Execer) error {
					applied++
					_, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS t (id TEXT)")
					return err
				},
				Down: func(ctx context.Context, db Execer) error {
					_, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS t")
					return err
				},
			},
		},
	}

	ctx := context.Background()
	if err := e.ExecutePlan(ctx, plan); err != nil {
		t.Fatalf("first ExecutePlan() error = %v", err)
	}
	if applied != 1 {
		t.Fatalf("expected 1 apply, got %d", applied)
	}

	// Re-run: should skip
	if err := e.ExecutePlan(ctx, plan); err != nil {
		t.Fatalf("second ExecutePlan() error = %v", err)
	}
	if applied != 1 {
		t.Fatalf("expected still 1 apply after re-run, got %d", applied)
	}
}

func TestExecutePlanRollbackOnFailure(t *testing.T) {
	db := openTestDB(t)
	e := NewMigrationExecutor(db)

	plan := &MigrationPlan{
		Module:  "base",
		Version: "v1.0.0",
		Migrations: []Migration{
			{
				ID: "20261109-001",
				Up: func(ctx context.Context, db Execer) error {
					_, err := db.ExecContext(ctx, "CREATE TABLE t1 (id TEXT)")
					return err
				},
				Down: func(ctx context.Context, db Execer) error {
					_, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS t1")
					return err
				},
			},
			{
				ID: "20261109-002",
				Up: func(ctx context.Context, db Execer) error {
					return errors.New("boom")
				},
				Down: func(ctx context.Context, db Execer) error {
					return nil
				},
			},
		},
	}

	ctx := context.Background()
	if err := e.ExecutePlan(ctx, plan); err == nil {
		t.Fatal("expected error on failing migration")
	}

	// Each migration runs in its own transaction: 001 committed before 002 failed,
	// so t1 exists and 001 is recorded. Only 002 is not applied.
	var tableCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='t1'").Scan(&tableCount); err != nil {
		t.Fatalf("t1 table should exist after first migration committed: %v", err)
	}
	if tableCount != 1 {
		t.Fatalf("expected t1 table to exist, got %d", tableCount)
	}

	exists, err := e.isMigrationApplied(ctx, "base:20261109-001")
	if err != nil {
		t.Fatalf("isMigrationApplied() error = %v", err)
	}
	if !exists {
		t.Fatal("first migration should be recorded as applied")
	}

	exists, err = e.isMigrationApplied(ctx, "base:20261109-002")
	if err != nil {
		t.Fatalf("isMigrationApplied() error = %v", err)
	}
	if exists {
		t.Fatal("failing migration should not be recorded as applied")
	}
}

func TestExecutePlanTopologicalOrder(t *testing.T) {
	db := openTestDB(t)
	e := NewMigrationExecutor(db)

	var order []string
	plan := &MigrationPlan{
		Module:  "base",
		Version: "v1.0.0",
		Migrations: []Migration{
			{
				ID:        "003",
				DependsOn: []string{"001"},
				Up: func(ctx context.Context, db Execer) error {
					order = append(order, "003")
					return nil
				},
				Down: func(ctx context.Context, db Execer) error { return nil },
			},
			{
				ID: "001",
				Up: func(ctx context.Context, db Execer) error {
					order = append(order, "001")
					return nil
				},
				Down: func(ctx context.Context, db Execer) error { return nil },
			},
			{
				ID:        "002",
				DependsOn: []string{"001"},
				Up: func(ctx context.Context, db Execer) error {
					order = append(order, "002")
					return nil
				},
				Down: func(ctx context.Context, db Execer) error { return nil },
			},
		},
	}

	ctx := context.Background()
	if err := e.ExecutePlan(ctx, plan); err != nil {
		t.Fatalf("ExecutePlan() error = %v", err)
	}

	if len(order) != 3 {
		t.Fatalf("expected 3 migrations, got %d: %v", len(order), order)
	}
	if order[0] != "001" {
		t.Fatalf("001 should run first, got %v", order)
	}
}

func TestExecutePlanCircularDependency(t *testing.T) {
	db := openTestDB(t)
	e := NewMigrationExecutor(db)

	plan := &MigrationPlan{
		Module:  "base",
		Version: "v1.0.0",
		Migrations: []Migration{
			{
				ID:        "001",
				DependsOn: []string{"002"},
				Up:        func(ctx context.Context, db Execer) error { return nil },
				Down:      func(ctx context.Context, db Execer) error { return nil },
			},
			{
				ID:        "002",
				DependsOn: []string{"001"},
				Up:        func(ctx context.Context, db Execer) error { return nil },
				Down:      func(ctx context.Context, db Execer) error { return nil },
			},
		},
	}

	ctx := context.Background()
	if err := e.ExecutePlan(ctx, plan); err == nil {
		t.Fatal("expected circular dependency error")
	}
}

func TestExecutePlanUnknownDependency(t *testing.T) {
	db := openTestDB(t)
	e := NewMigrationExecutor(db)

	plan := &MigrationPlan{
		Module:  "base",
		Version: "v1.0.0",
		Migrations: []Migration{
			{
				ID:        "001",
				DependsOn: []string{"missing"},
				Up:        func(ctx context.Context, db Execer) error { return nil },
				Down:      func(ctx context.Context, db Execer) error { return nil },
			},
		},
	}

	ctx := context.Background()
	if err := e.ExecutePlan(ctx, plan); err == nil {
		t.Fatal("expected unknown dependency error")
	}
}

func TestRollbackPlan(t *testing.T) {
	db := openTestDB(t)
	e := NewMigrationExecutor(db)

	plan := &MigrationPlan{
		Module:  "base",
		Version: "v1.0.0",
		Migrations: []Migration{
			{
				ID: "20261109-001",
				Up: func(ctx context.Context, db Execer) error {
					_, err := db.ExecContext(ctx, "CREATE TABLE t (id TEXT)")
					return err
				},
				Down: func(ctx context.Context, db Execer) error {
					_, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS t")
					return err
				},
			},
		},
	}

	ctx := context.Background()
	if err := e.ExecutePlan(ctx, plan); err != nil {
		t.Fatalf("ExecutePlan() error = %v", err)
	}

	if err := e.RollbackPlan(ctx, plan); err != nil {
		t.Fatalf("RollbackPlan() error = %v", err)
	}

	// Table should be gone
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='t'").Scan(&count); err != nil {
		t.Fatalf("query error: %v", err)
	}
	if count != 0 {
		t.Fatal("table t should be dropped after rollback")
	}
}

func TestRecordModuleVersion(t *testing.T) {
	db := openTestDB(t)
	e := NewMigrationExecutor(db)

	plan := &MigrationPlan{
		Module:  "base",
		Version: "v1.2.3",
		Migrations: []Migration{
			{
				ID:   "20261109-001",
				Up:   func(ctx context.Context, db Execer) error { return nil },
				Down: func(ctx context.Context, db Execer) error { return nil },
			},
		},
	}

	ctx := context.Background()
	if err := e.ExecutePlan(ctx, plan); err != nil {
		t.Fatal(err)
	}

	var version string
	if err := db.QueryRowContext(ctx, "SELECT version FROM aeng_module_versions WHERE module = ?", "base").Scan(&version); err != nil {
		t.Fatalf("module version not recorded: %v", err)
	}
	if version != "v1.2.3" {
		t.Fatalf("expected version v1.2.3, got %s", version)
	}
}

func TestParseKey(t *testing.T) {
	tests := []struct {
		key string
		mod string
		mig string
	}{
		{"base:20261109-001", "base", "20261109-001"},
		{"crm:001", "crm", "001"},
		{"no_colon", "no_colon", ""},
		{"a:b:c", "a:b", "c"},
	}
	for _, tt := range tests {
		mod, mig := parseKey(tt.key)
		if mod != tt.mod || mig != tt.mig {
			t.Errorf("parseKey(%q) = (%q, %q), want (%q, %q)", tt.key, mod, mig, tt.mod, tt.mig)
		}
	}
}

func TestMigrationExecutorNilDB(t *testing.T) {
	e := NewMigrationExecutor(nil)
	ctx := context.Background()
	if err := e.ExecutePlan(ctx, &MigrationPlan{Module: "base"}); err == nil {
		t.Fatal("expected error for nil db")
	}
}

func TestMigrationPlanNil(t *testing.T) {
	db := openTestDB(t)
	e := NewMigrationExecutor(db)
	ctx := context.Background()
	if err := e.ExecutePlan(ctx, nil); err == nil {
		t.Fatal("expected error for nil plan")
	}
}
