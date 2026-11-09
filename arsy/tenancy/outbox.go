package tenancy

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// OutboxRecord is a pending operation in the durable outbox.
// It tracks intent, idempotency key, state, attempt count, and result summary.
type OutboxRecord struct {
	ID             string    `json:"id"`
	Operation      string    `json:"operation"` // "provision", "migrate", "backup", "restore", "delete"
	TenantID       string    `json:"tenant_id"`
	IdempotencyKey string    `json:"idempotency_key"`
	State          string    `json:"state"` // "pending", "in_progress", "success", "failed_retryable", "failed_permanent"
	AttemptCount   int       `json:"attempt_count"`
	ResultSummary  string    `json:"result_summary,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// InitOutbox ensures the outbox table exists in the control-plane database.
func InitOutbox(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS arsy_outbox (
		id TEXT PRIMARY KEY,
		operation TEXT NOT NULL,
		tenant_id TEXT NOT NULL,
		idempotency_key TEXT,
		state TEXT NOT NULL DEFAULT 'pending',
		attempt_count INTEGER DEFAULT 0,
		result_summary TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)
	`)
	if err != nil {
		return fmt.Errorf("init outbox: %w", err)
	}
	return nil
}

// ClaimOutboxRecord atomically claims a pending outbox record for processing.
// Returns nil if no records are pending.
func ClaimOutboxRecord(ctx context.Context, db *sql.DB) (*OutboxRecord, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("claim: begin tx: %w", err)
	}
	defer tx.Rollback()
	var rec OutboxRecord
	row := tx.QueryRowContext(ctx, `
	SELECT id, operation, tenant_id, idempotency_key, state, attempt_count, result_summary, created_at, updated_at
	FROM arsy_outbox
	WHERE state IN ('pending', 'in_progress')
	ORDER BY created_at ASC
	LIMIT 1
	`)
	if err := row.Scan(&rec.ID, &rec.Operation, &rec.TenantID, &rec.IdempotencyKey, &rec.State, &rec.AttemptCount, &rec.ResultSummary, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("claim: scan: %w", err)
	}
	now := time.Now().UTC()
	_, err = tx.ExecContext(ctx, `
	UPDATE arsy_outbox SET state = ?, attempt_count = attempt_count + 1, updated_at = ?
	WHERE id = ?
	`, "in_progress", now.Format(time.RFC3339Nano), rec.ID)
	if err != nil {
		return nil, fmt.Errorf("claim: update: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("claim: commit: %w", err)
	}
	return &rec, nil
}

// UpdateOutboxRecord marks an operation as succeeded or failed with a result summary.
func UpdateOutboxRecord(ctx context.Context, db *sql.DB, id, newState, summary string) error {
	now := time.Now().UTC()
	_, err := db.ExecContext(ctx, `
	UPDATE arsy_outbox SET state = ?, result_summary = ?, updated_at = ?
	WHERE id = ?
	`, newState, summary, now.Format(time.RFC3339Nano), id)
	if err != nil {
		return fmt.Errorf("update outbox: %w", err)
	}
	return nil
}

// ReconcileStalledRecords moves stalled in-progress records back to pending
// if they exceeded a retry window (e.g., 1 hour).
func ReconcileStalledRecords(ctx context.Context, db *sql.DB, stalledAfter time.Duration) error {
	cutoff := time.Now().UTC().Add(-stalledAfter)
	_, err := db.ExecContext(ctx, `
	UPDATE arsy_outbox SET state = 'pending'
	WHERE state = 'in_progress' AND updated_at < ?
	`, cutoff.Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("reconcile stalled: %w", err)
	}
	return nil
}

// OutboxProvisionRegistry is a durable ProvisionRegistry backed by the
// control-plane outbox table. Records survive process restarts, so a repeated
// provisioning request with the same idempotency key is recognised after a crash.
type OutboxProvisionRegistry struct {
	db *sql.DB
}

// NewOutboxProvisionRegistry creates a registry on db and ensures the outbox
// table exists.
func NewOutboxProvisionRegistry(ctx context.Context, db *sql.DB) (*OutboxProvisionRegistry, error) {
	if db == nil {
		return nil, fmt.Errorf("tenancy: provision registry requires a database")
	}
	if err := InitOutbox(ctx, db); err != nil {
		return nil, err
	}
	return &OutboxProvisionRegistry{db: db}, nil
}

// LookupProvision implements ProvisionRegistry.
func (r *OutboxProvisionRegistry) LookupProvision(ctx context.Context, key string) (ProvisionRecord, bool, error) {
	var (
		record ProvisionRecord
		state  string
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT tenant_id, state, COALESCE(result_summary, '')
		FROM arsy_outbox
		WHERE operation = 'provision' AND idempotency_key = ?
		ORDER BY updated_at DESC
		LIMIT 1`, key).
		Scan(&record.TenantID, &state, &record.ResultSummary)
	if errors.Is(err, sql.ErrNoRows) {
		return ProvisionRecord{}, false, nil
	}
	if err != nil {
		return ProvisionRecord{}, false, fmt.Errorf("tenancy provision: read registry key %q: %w", key, err)
	}
	record.Key = key
	record.State = ProvisionState(state)
	return record, true, nil
}

// SaveProvision implements ProvisionRegistry.
func (r *OutboxProvisionRegistry) SaveProvision(ctx context.Context, record ProvisionRecord) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	id := "provision:" + record.Key
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO arsy_outbox
			(id, operation, tenant_id, idempotency_key, state, attempt_count, result_summary, created_at, updated_at)
		VALUES (?, 'provision', ?, ?, ?, 0, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			tenant_id = excluded.tenant_id,
			state = excluded.state,
			result_summary = excluded.result_summary,
			attempt_count = arsy_outbox.attempt_count + 1,
			updated_at = excluded.updated_at`,
		id, record.TenantID, record.Key, string(record.State), record.ResultSummary, now, now)
	if err != nil {
		return fmt.Errorf("tenancy provision: save registry key %q: %w", record.Key, err)
	}
	return nil
}
