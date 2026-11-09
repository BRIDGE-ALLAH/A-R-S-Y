// Package aeng contains the Application Engine framework.
// AENG-032: Versioned Application Schema Migration Executor.
package aeng

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Execer is implemented by both *sql.DB and *sql.Tx so migrations run
// transactionally when a transaction is supplied.
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// MigrationFn applies or reverts a schema migration on a tenant database.
type MigrationFn func(ctx context.Context, db Execer) error

// Migration defines a single schema migration.
type Migration struct {
	ID          string      // Unique migration ID (e.g., "20261109-001")
	Version     string      // Migration version within module
	Description string      // Human-readable description
	Up          MigrationFn // Apply migration
	Down        MigrationFn // Revert migration
	DependsOn   []string    // Migration IDs this depends on
	Tags        []string    // Tags: "add", "modify", "remove", "data", "schema"
	Safe        bool        // Can run concurrently with other migrations
}

// MigrationPlan defines the full migration plan for a module.
type MigrationPlan struct {
	Module      string      // Module identifier
	Version     string      // Module version (e.g., "v1.2.3")
	Description string      // Human-readable description
	Migrations  []Migration // Ordered list of migrations
	Seeds       []*SeedDef  // Associated seed data
	DependsOn   []string    // Other module plan IDs this depends on
}

// MigrationExecutor applies migration plans to tenant databases.
type MigrationExecutor struct {
	mu      sync.Mutex
	db      *sql.DB
	applied map[string]bool // module:migration_id -> applied
}

// NewMigrationExecutor creates a migration executor for a database.
func NewMigrationExecutor(db *sql.DB) *MigrationExecutor {
	return &MigrationExecutor{
		db:      db,
		applied: make(map[string]bool),
	}
}

// ExecutePlan applies all migrations in a plan to the database.
func (e *MigrationExecutor) ExecutePlan(ctx context.Context, plan *MigrationPlan) error {
	if e.db == nil {
		return errors.New("aeng: nil database")
	}
	if plan == nil {
		return errors.New("aeng: nil migration plan")
	}

	// Ensure the tracking table exists
	if err := e.ensureMigrationTable(ctx); err != nil {
		return fmt.Errorf("aeng: ensure migration table: %w", err)
	}

	// Sort migrations by dependencies
	ordered, err := e.topologicalSort(plan.Migrations)
	if err != nil {
		return fmt.Errorf("aeng: topological sort: %w", err)
	}

	// Execute each migration in order
	for _, m := range ordered {
		key := plan.Module + ":" + m.ID

		// Skip already-applied migrations (idempotent)
		if e.applied[key] {
			continue
		}

		// Check if already applied in DB
		exists, err := e.isMigrationApplied(ctx, key)
		if err != nil {
			return fmt.Errorf("aeng: check migration %s: %w", key, err)
		}
		if exists {
			e.applied[key] = true
			continue
		}

		// Execute migration in a transaction
		tx, err := e.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("aeng: begin tx for migration %s: %w", key, err)
		}

		if err := m.Up(ctx, tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("aeng: migration %s up failed: %w", key, err)
		}

		if err := e.recordMigrationApplied(ctx, tx, key); err != nil {
			tx.Rollback()
			return fmt.Errorf("aeng: record migration %s: %w", key, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("aeng: commit migration %s: %w", key, err)
		}

		e.applied[key] = true
	}

	// Record module version
	if err := e.recordModuleVersion(ctx, plan.Module, plan.Version); err != nil {
		return fmt.Errorf("aeng: record module version: %w", err)
	}

	return nil
}

// RollbackPlan reverts migrations for a plan in reverse order.
func (e *MigrationExecutor) RollbackPlan(ctx context.Context, plan *MigrationPlan) error {
	if plan == nil {
		return errors.New("aeng: nil migration plan")
	}

	// Reverse order
	for i := len(plan.Migrations) - 1; i >= 0; i-- {
		m := plan.Migrations[i]
		key := plan.Module + ":" + m.ID

		exists, err := e.isMigrationApplied(ctx, key)
		if err != nil {
			return fmt.Errorf("aeng: check migration %s: %w", key, err)
		}
		if !exists {
			continue // not applied, skip
		}

		tx, err := e.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("aeng: begin tx for rollback %s: %w", key, err)
		}

		if err := m.Down(ctx, tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("aeng: migration %s down failed: %w", key, err)
		}

		if err := e.recordMigrationReverted(ctx, tx, key); err != nil {
			tx.Rollback()
			return fmt.Errorf("aeng: record rollback %s: %w", key, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("aeng: commit rollback %s: %w", key, err)
		}

		delete(e.applied, key)
	}

	return nil
}

// ensureMigrationTable creates the migration tracking table if it doesn't exist.
func (e *MigrationExecutor) ensureMigrationTable(ctx context.Context) error {
	_, err := e.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS aeng_migrations (
			module TEXT NOT NULL,
			migration_id TEXT NOT NULL,
			applied_at TEXT NOT NULL DEFAULT '',
			PRIMARY KEY (module, migration_id)
		)
	`)
	return err
}

// isMigrationApplied checks if a migration has been applied.
func (e *MigrationExecutor) isMigrationApplied(ctx context.Context, key string) (bool, error) {
	var count int
	err := e.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM aeng_migrations WHERE module || ':' || migration_id = ?
	`, key).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// recordMigrationApplied records a migration as applied.
func (e *MigrationExecutor) recordMigrationApplied(ctx context.Context, tx *sql.Tx, key string) error {
	// Parse module:migration_id
	module, migrationID := parseKey(key)
	_, err := tx.ExecContext(ctx, `
		INSERT OR REPLACE INTO aeng_migrations (module, migration_id, applied_at)
		VALUES (?, ?, '')
	`, module, migrationID)
	return err
}

// recordMigrationReverted removes a migration record on rollback.
func (e *MigrationExecutor) recordMigrationReverted(ctx context.Context, tx *sql.Tx, key string) error {
	module, migrationID := parseKey(key)
	_, err := tx.ExecContext(ctx, `
		DELETE FROM aeng_migrations WHERE module = ? AND migration_id = ?
	`, module, migrationID)
	return err
}

// recordModuleVersion records the module version in a separate table.
func (e *MigrationExecutor) recordModuleVersion(ctx context.Context, module, version string) error {
	_, err := e.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS aeng_module_versions (
			module TEXT PRIMARY KEY,
			version TEXT NOT NULL,
			updated_at TEXT NOT NULL DEFAULT ''
		)
	`)
	if err != nil {
		return err
	}

	_, err = e.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO aeng_module_versions (module, version, updated_at)
		VALUES (?, ?, '')
	`, module, version)
	return err
}

// topologicalSort sorts migrations by their dependencies.
func (e *MigrationExecutor) topologicalSort(migrations []Migration) ([]Migration, error) {
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	var sorted []Migration

	index := make(map[string]int)
	for i, m := range migrations {
		index[m.ID] = i
	}

	var visit func(m Migration) error
	visit = func(m Migration) error {
		if visited[m.ID] {
			return nil
		}
		if visiting[m.ID] {
			return fmt.Errorf("aeng: circular dependency at migration %s", m.ID)
		}

		visiting[m.ID] = true

		for _, dep := range m.DependsOn {
			depIdx, ok := index[dep]
			if !ok {
				return fmt.Errorf("aeng: migration %s depends on unknown migration %s", m.ID, dep)
			}
			if err := visit(migrations[depIdx]); err != nil {
				return err
			}
		}

		visiting[m.ID] = false
		visited[m.ID] = true
		sorted = append(sorted, m)

		return nil
	}

	// Sort migrations by ID for deterministic ordering
	sortedByID := make([]Migration, len(migrations))
	copy(sortedByID, migrations)
	sort.Slice(sortedByID, func(i, j int) bool {
		return sortedByID[i].ID < sortedByID[j].ID
	})

	for _, m := range sortedByID {
		if err := visit(m); err != nil {
			return nil, err
		}
	}

	return sorted, nil
}

func parseKey(key string) (string, string) {
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == ':' {
			return key[:i], key[i+1:]
		}
	}
	return key, ""
}
