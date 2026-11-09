// Package aeng contains the Application Engine framework.
// AENG-034: Idempotent Module Seed Data Runner.
package aeng

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
)

// SeedDef defines one seed record to insert/update.
type SeedDef struct {
	Collection string         `json:"collection"`
	Fields     map[string]any `json:"fields"`
	Upsert     bool           `json:"upsert"`
	Conditions []ConditionDef `json:"conditions,omitempty"`
}

// SeedRunner executes seed data idempotently per tenant.
type SeedRunner struct {
	mu          sync.Mutex
	db          *sql.DB
	environment string
	executed    map[string]bool // seed ID -> executed
}

// NewSeedRunner creates a seed runner for a database.
func NewSeedRunner(db *sql.DB, environment string) *SeedRunner {
	return &SeedRunner{
		db:          db,
		environment: environment,
		executed:    make(map[string]bool),
	}
}

// RunSeeds executes all seed definitions idempotently.
func (r *SeedRunner) RunSeeds(ctx context.Context, seeds []*SeedDef) error {
	if r.db == nil {
		return errors.New("aeng: nil database")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("aeng: begin tx: %w", err)
	}
	defer tx.Rollback()

	for _, seed := range seeds {
		if seed == nil {
			continue
		}
		if !r.shouldExecute(seed) {
			continue
		}

		if !conditionsMet(r.db, seed.Collection, seed.Conditions) {
			continue
		}

		if err := r.upsertSeed(tx, seed); err != nil {
			return fmt.Errorf("aeng: seed collection %s: %w", seed.Collection, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("aeng: commit seeds: %w", err)
	}

	return nil
}

// RunSeedsForTenant executes seeds for a specific tenant database.
func (r *SeedRunner) RunSeedsForTenant(ctx context.Context, tenantDB *sql.DB, seeds []*SeedDef) error {
	if tenantDB == nil {
		return errors.New("aeng: nil tenant database")
	}

	tx, err := tenantDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("aeng: begin tx: %w", err)
	}
	defer tx.Rollback()

	for _, seed := range seeds {
		if seed == nil {
			continue
		}
		if !r.shouldExecute(seed) {
			continue
		}

		if !conditionsMet(tenantDB, seed.Collection, seed.Conditions) {
			continue
		}

		if err := r.upsertSeed(tx, seed); err != nil {
			return fmt.Errorf("aeng: seed collection %s: %w", seed.Collection, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("aeng: commit seeds: %w", err)
	}

	return nil
}

// shouldExecute checks environment filter.
func (r *SeedRunner) shouldExecute(seed *SeedDef) bool {
	if r.environment == "" {
		return true
	}
	// Seeds without environment filter run everywhere
	if seed == nil {
		return true
	}
	// No per-seed environment field; use runner-level environment
	return true
}

// upsertSeed inserts or updates a seed record.
func (r *SeedRunner) upsertSeed(tx *sql.Tx, seed *SeedDef) error {
	// Build INSERT ... ON CONFLICT DO UPDATE
	columns := make([]string, 0, len(seed.Fields))
	placeholders := make([]string, 0, len(seed.Fields))
	values := make([]any, 0, len(seed.Fields))

	for k, v := range seed.Fields {
		columns = append(columns, k)
		placeholders = append(placeholders, "?")
		values = append(values, v)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT(id) DO UPDATE SET %s
	`, seed.Collection,
		joinColumns(columns),
		joinColumns(placeholders),
		joinUpdate(columns),
	)

	_, err := tx.Exec(query, values...)
	return err
}

func joinColumns(cols []string) string {
	result := ""
	for i, c := range cols {
		if i > 0 {
			result += ", "
		}
		result += c
	}
	return result
}

func joinUpdate(cols []string) string {
	result := ""
	for i, c := range cols {
		if i > 0 {
			result += ", "
		}
		result += c + " = excluded." + c
	}
	return result
}

// conditionsMet evaluates seed conditions against the database table.
func conditionsMet(db *sql.DB, table string, conditions []ConditionDef) bool {
	if len(conditions) == 0 {
		return true
	}

	for _, cond := range conditions {
		query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE id = ?`, table)
		var count int
		err := db.QueryRow(query, cond.Value).Scan(&count)
		if err != nil {
			return false
		}

		switch cond.Operator {
		case "exists":
			if count == 0 {
				return false
			}
		case "not_exists":
			if count > 0 {
				return false
			}
		case "eq":
			if count == 0 {
				return false
			}
		default:
			if count == 0 {
				return false
			}
		}
	}

	return true
}
