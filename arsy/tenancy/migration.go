 package tenancy

 import (
 	"context"
 	"database/sql"
 	"fmt"
 	"sort"
 )

 // Migration is a versioned schema change for a tenant database.
 // Version must be a sortable string (e.g. "20261107000001").
 type Migration struct {
 	Version string
 	Up      func(db *sql.DB) error
 	Down    func(db *sql.DB) error
 }

 // MigrationRunner applies pending migrations to a tenant Datastore.
 type MigrationRunner struct {
 	migrations []Migration
 }

 // NewMigrationRunner creates a runner for the given ordered migrations.
 func NewMigrationRunner(migrations []Migration) *MigrationRunner {
 	sorted := make([]Migration, len(migrations))
 	copy(sorted, migrations)
 	sort.Slice(sorted, func(i, j int) bool {
 		return sorted[i].Version < sorted[j].Version
 	})
 	return &MigrationRunner{migrations: sorted}
 }

 // Migrate ensures the schema_versions table exists and applies all pending migrations.
 func (r *MigrationRunner) Migrate(ctx context.Context, ds Datastore) error {
 	db := ds.DB()
 	if _, err := db.ExecContext(ctx, `
 		CREATE TABLE IF NOT EXISTS arsy_schema_versions (
 			version TEXT PRIMARY KEY,
 			applied_at DATETIME DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
 		)`); err != nil {
 		return fmt.Errorf("tenancy migrate: ensure schema_versions: %w", err)
 	}
 	for _, m := range r.migrations {
 		var exists int
 		row := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM arsy_schema_versions WHERE version=?", m.Version)
 		if err := row.Scan(&exists); err != nil {
 			return fmt.Errorf("tenancy migrate: check %s: %w", m.Version, err)
 		}
 		if exists > 0 {
 			continue
 		}
 		if err := m.Up(db); err != nil {
 			return fmt.Errorf("tenancy migrate: apply %s: %w", m.Version, err)
 		}
 		if _, err := db.ExecContext(ctx,
 			"INSERT INTO arsy_schema_versions (version) VALUES (?)", m.Version); err != nil {
 			return fmt.Errorf("tenancy migrate: record %s: %w", m.Version, err)
 		}
 	}
 	return nil
 }

 // AppliedVersions returns all versions already applied in ds.
 func (r *MigrationRunner) AppliedVersions(ctx context.Context, ds Datastore) ([]string, error) {
 	rows, err := ds.DB().QueryContext(ctx, `
 		SELECT version FROM arsy_schema_versions ORDER BY version`)
 	if err != nil {
 		return nil, fmt.Errorf("tenancy: list applied versions: %w", err)
 	}
 	defer rows.Close()
 	var versions []string
 	for rows.Next() {
 		var v string
 		if err := rows.Scan(&v); err != nil {
 			return nil, err
 		}
 		versions = append(versions, v)
 	}
 	return versions, rows.Err()
 }
