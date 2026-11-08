package tenancy_test

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
)

func openTenantDatastore(t *testing.T, tenantID string) tenancy.Datastore {
	t.Helper()
	dirs, err := config.EnsureBaaSDirectories(t.TempDir(), "")
	if err != nil {
		t.Fatalf("EnsureBaaSDirectories: %v", err)
	}
	provider := tenancy.NewSQLiteProvider(dirs)
	t.Cleanup(func() { _ = provider.Close() })
	ds, err := provider.Open(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("Open(%q): %v", tenantID, err)
	}
	return ds
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var found string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, name).Scan(&found)
	return err == nil && found == name
}

func TestMigrationRunner_DuplicateAndInvalidVersions(t *testing.T) {
	ds := openTenantDatastore(t, "invalid-versions")
	ctx := context.Background()
	apply := func(_ *sql.DB) error { return nil }

	cases := []struct {
		name       string
		migrations []tenancy.Migration
		want       error
	}{
		{
			name: "duplicate",
			migrations: []tenancy.Migration{
				{Version: "1", Up: apply},
				{Version: "1", Up: apply},
			},
			want: tenancy.ErrMigrationDuplicateVersion,
		},
		{
			name:       "empty version",
			migrations: []tenancy.Migration{{Version: " ", Up: apply}},
			want:       tenancy.ErrMigrationInvalid,
		},
		{
			name:       "missing apply function",
			migrations: []tenancy.Migration{{Version: "2"}},
			want:       tenancy.ErrMigrationInvalid,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tenancy.NewMigrationRunner(tc.migrations).Migrate(ctx, ds)
			if !errors.Is(err, tc.want) {
				t.Fatalf("Migrate error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestMigrationRunner_FailedMigrationIsRecoverable(t *testing.T) {
	ds := openTenantDatastore(t, "failed-migration")
	ctx := context.Background()
	upstream := errors.New("boom")

	failing := tenancy.NewMigrationRunner([]tenancy.Migration{{
		Version:     "0001",
		Description: "create items",
		Checksum:    "checksum-0001",
		Up: func(db *sql.DB) error {
			if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS recoverable_items (id TEXT PRIMARY KEY)`); err != nil {
				return err
			}
			return upstream
		},
	}})
	err := failing.Migrate(ctx, ds)
	if !errors.Is(err, upstream) {
		t.Fatalf("Migrate error = %v, want %v", err, upstream)
	}

	versions, err := failing.AppliedVersions(ctx, ds)
	if err != nil {
		t.Fatalf("AppliedVersions: %v", err)
	}
	if len(versions) != 0 {
		t.Fatalf("failed migration reported as applied: %v", versions)
	}

	records, err := failing.AppliedMigrations(ctx, ds)
	if err != nil {
		t.Fatalf("AppliedMigrations: %v", err)
	}
	if len(records) != 1 || records[0].Status != tenancy.MigrationFailed || records[0].Error == "" {
		t.Fatalf("failure not recorded: %+v", records)
	}

	recovered := tenancy.NewMigrationRunner([]tenancy.Migration{{
		Version:     "0001",
		Description: "create items",
		Checksum:    "checksum-0001",
		Up: func(db *sql.DB) error {
			_, err := db.Exec(`CREATE TABLE IF NOT EXISTS recoverable_items (id TEXT PRIMARY KEY)`)
			return err
		},
	}})
	if err := recovered.Migrate(ctx, ds); err != nil {
		t.Fatalf("retry Migrate: %v", err)
	}
	versions, err = recovered.AppliedVersions(ctx, ds)
	if err != nil {
		t.Fatalf("AppliedVersions after retry: %v", err)
	}
	if len(versions) != 1 || versions[0] != "0001" {
		t.Fatalf("versions after retry = %v, want [0001]", versions)
	}
}

func TestMigrationRunner_RejectsChecksumDrift(t *testing.T) {
	ds := openTenantDatastore(t, "checksum-drift")
	ctx := context.Background()
	apply := func(db *sql.DB) error {
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS drift_items (id TEXT PRIMARY KEY)`)
		return err
	}

	original := tenancy.NewMigrationRunner([]tenancy.Migration{{
		Version: "0001", Checksum: tenancy.MigrationChecksum("CREATE TABLE drift_items"), Up: apply,
	}})
	if err := original.Migrate(ctx, ds); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}

	drifted := tenancy.NewMigrationRunner([]tenancy.Migration{{
		Version: "0001", Checksum: tenancy.MigrationChecksum("DROP TABLE drift_items"), Up: apply,
	}})
	err := drifted.Migrate(ctx, ds)
	if !errors.Is(err, tenancy.ErrMigrationChecksumDrift) {
		t.Fatalf("Migrate error = %v, want ErrMigrationChecksumDrift", err)
	}
}

func TestMigrationRunner_ConcurrentAttemptsApplyOnce(t *testing.T) {
	ds := openTenantDatastore(t, "concurrent-migrations")
	ctx := context.Background()

	var applied atomic.Int32
	migrations := []tenancy.Migration{
		{
			Version: "0001", Checksum: "concurrent-0001",
			Up: func(db *sql.DB) error {
				applied.Add(1)
				time.Sleep(20 * time.Millisecond)
				_, err := db.Exec(`CREATE TABLE IF NOT EXISTS concurrent_items (id TEXT PRIMARY KEY)`)
				return err
			},
		},
	}

	const goroutines = 16
	start := make(chan struct{})
	errs := make(chan error, goroutines)
	var wait sync.WaitGroup
	for index := 0; index < goroutines; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			errs <- tenancy.NewMigrationRunner(migrations).Migrate(ctx, ds)
		}()
	}
	close(start)
	wait.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent Migrate: %v", err)
		}
	}
	if got := applied.Load(); got != 1 {
		t.Fatalf("Up executions = %d, want 1", got)
	}
}

func TestMigrationRunner_TransactionalRollback(t *testing.T) {
	ds := openTenantDatastore(t, "transactional-migration")
	ctx := context.Background()
	upstream := errors.New("ddl failed")

	failing := tenancy.NewMigrationRunner([]tenancy.Migration{{
		Version: "0001", Checksum: "tx-0001",
		UpTx: func(tx *sql.Tx) error {
			if _, err := tx.Exec(`CREATE TABLE tx_items (id TEXT PRIMARY KEY)`); err != nil {
				return err
			}
			return upstream
		},
	}})
	if err := failing.Migrate(ctx, ds); !errors.Is(err, upstream) {
		t.Fatalf("Migrate error = %v, want %v", err, upstream)
	}
	if tableExists(t, ds.DB(), "tx_items") {
		t.Fatal("transactional migration left DDL behind after failure")
	}

	recovered := tenancy.NewMigrationRunner([]tenancy.Migration{{
		Version: "0001", Checksum: "tx-0001",
		UpTx: func(tx *sql.Tx) error {
			_, err := tx.Exec(`CREATE TABLE tx_items (id TEXT PRIMARY KEY)`)
			return err
		},
	}})
	if err := recovered.Migrate(ctx, ds); err != nil {
		t.Fatalf("retry Migrate: %v", err)
	}
	if !tableExists(t, ds.DB(), "tx_items") {
		t.Fatal("transactional migration did not apply on retry")
	}
}

func TestMigrationRunner_LegacyRegistryUpgrade(t *testing.T) {
	ds := openTenantDatastore(t, "legacy-registry")
	ctx := context.Background()
	db := ds.DB()

	// Simulate the registry written by the first BAAS-011 implementation.
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE arsy_schema_versions (
			version TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		)`); err != nil {
		t.Fatalf("create legacy registry: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO arsy_schema_versions (version) VALUES ('0001')`); err != nil {
		t.Fatalf("seed legacy registry: %v", err)
	}

	var legacyRerun, newApply atomic.Int32
	runner := tenancy.NewMigrationRunner([]tenancy.Migration{
		{
			Version: "0001", Checksum: "legacy-checksum",
			Up: func(*sql.DB) error { legacyRerun.Add(1); return nil },
		},
		{
			Version: "0002", Checksum: "new-checksum",
			Up: func(*sql.DB) error { newApply.Add(1); return nil },
		},
	})
	if err := runner.Migrate(ctx, ds); err != nil {
		t.Fatalf("Migrate legacy registry: %v", err)
	}
	if legacyRerun.Load() != 0 {
		t.Fatalf("legacy applied migration ran again")
	}
	if newApply.Load() != 1 {
		t.Fatalf("new migration runs = %d, want 1", newApply.Load())
	}

	records, err := runner.AppliedMigrations(ctx, ds)
	if err != nil {
		t.Fatalf("AppliedMigrations: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("records = %+v, want 2", records)
	}
	if records[0].Version != "0001" || records[0].Checksum != "legacy-checksum" || records[0].Status != tenancy.MigrationApplied {
		t.Fatalf("legacy record not upgraded: %+v", records[0])
	}
	if records[1].Version != "0002" || records[1].Checksum != "new-checksum" || records[1].Status != tenancy.MigrationApplied {
		t.Fatalf("new record = %+v", records[1])
	}
}
