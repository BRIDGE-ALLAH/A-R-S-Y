package tenancy

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrMigrationInvalid          = errors.New("tenancy: invalid migration")
	ErrMigrationDuplicateVersion = errors.New("tenancy: duplicate migration version")
	ErrMigrationChecksumDrift    = errors.New("tenancy: migration checksum drift")
)

type MigrationStatus string

const (
	MigrationApplied MigrationStatus = "applied"
	MigrationFailed  MigrationStatus = "failed"
)

const schemaVersionsTable = "arsy_schema_versions"

type Migration struct {
	Version     string
	Description string
	Checksum    string
	Up          func(db *sql.DB) error
	UpTx        func(tx *sql.Tx) error
	Down        func(db *sql.DB) error
}

func MigrationChecksum(parts ...string) string {
	digest := sha256.New()
	for idx, part := range parts {
		if idx > 0 {
			digest.Write([]byte{0})
		}
		digest.Write([]byte(part))
	}
	return hex.EncodeToString(digest.Sum(nil))
}

type MigrationRecord struct {
	Version     string
	Description string
	Checksum    string
	Status      MigrationStatus
	Error       string
	AppliedAt   time.Time
}

type MigrationRunner struct {
	migrations []Migration
}

func NewMigrationRunner(migrations []Migration) *MigrationRunner {
	sorted := make([]Migration, len(migrations))
	copy(sorted, migrations)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Version < sorted[j].Version
	})
	return &MigrationRunner{migrations: sorted}
}

type migrationLockEntry struct {
	mutex sync.Mutex
	refs  int
}

// migrationLocks serialises migration work per underlying *sql.DB handle.
// The key is the handle, not the tenant ID, because a Datastore owns exactly one
// pool; after an eviction the provider opens a new handle and SQLite's own writer
// serialisation (plus the checksum/version recheck) covers the cross-process case.
// Entries are reference-counted so closed handles do not leak.
var migrationLocks = struct {
	mutex   sync.Mutex
	entries map[*sql.DB]*migrationLockEntry
}{entries: make(map[*sql.DB]*migrationLockEntry)}

func acquireMigrationLock(db *sql.DB) *migrationLockEntry {
	migrationLocks.mutex.Lock()
	entry := migrationLocks.entries[db]
	if entry == nil {
		entry = &migrationLockEntry{}
		migrationLocks.entries[db] = entry
	}
	entry.refs++
	migrationLocks.mutex.Unlock()
	entry.mutex.Lock()
	return entry
}

func releaseMigrationLock(db *sql.DB, entry *migrationLockEntry) {
	entry.mutex.Unlock()
	migrationLocks.mutex.Lock()
	entry.refs--
	if entry.refs == 0 {
		delete(migrationLocks.entries, db)
	}
	migrationLocks.mutex.Unlock()
}

func (r *MigrationRunner) validate() error {
	seen := make(map[string]struct{}, len(r.migrations))
	for index, migration := range r.migrations {
		version := strings.TrimSpace(migration.Version)
		if version == "" {
			return fmt.Errorf("%w: index %d has an empty version", ErrMigrationInvalid, index)
		}
		if migration.Up == nil && migration.UpTx == nil {
			return fmt.Errorf("%w: version %q has no apply function", ErrMigrationInvalid, migration.Version)
		}
		if _, duplicate := seen[migration.Version]; duplicate {
			return fmt.Errorf("%w: %q", ErrMigrationDuplicateVersion, migration.Version)
		}
		seen[migration.Version] = struct{}{}
	}
	return nil
}

func (r *MigrationRunner) Migrate(ctx context.Context, ds Datastore) error {
	if ds == nil {
		return fmt.Errorf("%w: nil datastore", ErrMigrationInvalid)
	}
	if err := r.validate(); err != nil {
		return err
	}
	db := ds.DB()
	if db == nil {
		return fmt.Errorf("%w: datastore has no database handle", ErrMigrationInvalid)
	}
	lock := acquireMigrationLock(db)
	defer releaseMigrationLock(db, lock)

	if err := ensureSchemaVersionsTable(ctx, db); err != nil {
		return err
	}
	for _, migration := range r.migrations {
		if err := applyMigration(ctx, db, migration); err != nil {
			return err
		}
	}
	return nil
}

func (r *MigrationRunner) AppliedVersions(ctx context.Context, ds Datastore) ([]string, error) {
	records, err := r.AppliedMigrations(ctx, ds)
	if err != nil {
		return nil, err
	}
	versions := make([]string, 0, len(records))
	for _, record := range records {
		if record.Status == MigrationApplied {
			versions = append(versions, record.Version)
		}
	}
	return versions, nil
}

func (r *MigrationRunner) AppliedMigrations(ctx context.Context, ds Datastore) ([]MigrationRecord, error) {
	if ds == nil {
		return nil, fmt.Errorf("%w: nil datastore", ErrMigrationInvalid)
	}
	db := ds.DB()
	if db == nil {
		return nil, fmt.Errorf("%w: datastore has no database handle", ErrMigrationInvalid)
	}
	lock := acquireMigrationLock(db)
	defer releaseMigrationLock(db, lock)

	if err := ensureSchemaVersionsTable(ctx, db); err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
		SELECT version, description, checksum, status, error, applied_at
		FROM `+schemaVersionsTable+` ORDER BY version`)
	if err != nil {
		return nil, fmt.Errorf("tenancy: list applied migrations: %w", err)
	}
	defer rows.Close()

	var records []MigrationRecord
	for rows.Next() {
		var (
			record    MigrationRecord
			status    string
			appliedAt any
		)
		if err := rows.Scan(&record.Version, &record.Description, &record.Checksum, &status, &record.Error, &appliedAt); err != nil {
			return nil, err
		}
		record.Status = MigrationStatus(status)
		if record.Status == "" {
			record.Status = MigrationApplied // legacy rows predate the status column
		}
		record.AppliedAt = parseMigrationTime(appliedAt)
		records = append(records, record)
	}
	return records, rows.Err()
}

func applyMigration(ctx context.Context, db *sql.DB, migration Migration) error {
	recorded, found, err := lookupMigration(ctx, db, migration.Version)
	if err != nil {
		return err
	}
	if found {
		status := recorded.Status
		if status == "" {
			status = MigrationApplied
		}
		switch status {
		case MigrationApplied:
			if recorded.Checksum != "" && migration.Checksum != "" && recorded.Checksum != migration.Checksum {
				return fmt.Errorf("%w: version %q recorded %s but runner supplies %s",
					ErrMigrationChecksumDrift, migration.Version, recorded.Checksum, migration.Checksum)
			}
			if recorded.Checksum == "" && migration.Checksum != "" {
				if _, err := db.ExecContext(ctx, `
					UPDATE `+schemaVersionsTable+` SET checksum=? WHERE version=?`,
					migration.Checksum, migration.Version); err != nil {
					return fmt.Errorf("tenancy migrate: record checksum for %s: %w", migration.Version, err)
				}
			}
			return nil
		case MigrationFailed:
			// Retry below; a failed version is never terminal.
		default:
			return fmt.Errorf("%w: version %q has unknown status %q", ErrMigrationInvalid, migration.Version, status)
		}
	}

	if migration.UpTx != nil {
		return appliedWithinTransaction(ctx, db, migration)
	}
	if err := migration.Up(db); err != nil {
		recordErr := recordMigrationFailure(ctx, db, migration, err)
		applyErr := fmt.Errorf("tenancy migrate: apply %s: %w", migration.Version, err)
		if recordErr != nil {
			return errors.Join(applyErr, recordErr)
		}
		return applyErr
	}
	if err := recordMigrationSuccess(ctx, db, migration); err != nil {
		return fmt.Errorf("tenancy migrate: record %s: %w", migration.Version, err)
	}
	return nil
}

func appliedWithinTransaction(ctx context.Context, db *sql.DB, migration Migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("tenancy migrate: begin %s: %w", migration.Version, err)
	}
	if err := migration.UpTx(tx); err != nil {
		_ = tx.Rollback()
		if recordErr := recordMigrationFailure(ctx, db, migration, err); recordErr != nil {
			return errors.Join(
				fmt.Errorf("tenancy migrate: apply %s: %w", migration.Version, err),
				recordErr,
			)
		}
		return fmt.Errorf("tenancy migrate: apply %s: %w", migration.Version, err)
	}
	if _, err := tx.ExecContext(ctx, upsertMigrationApplied, migration.Version,
		migration.Description, migration.Checksum); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("tenancy migrate: record %s: %w", migration.Version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("tenancy migrate: commit %s: %w", migration.Version, err)
	}
	return nil
}

func recordMigrationFailure(ctx context.Context, db *sql.DB, migration Migration, cause error) error {
	if _, err := db.ExecContext(ctx, upsertMigrationFailed,
		migration.Version, migration.Description, migration.Checksum, cause.Error()); err != nil {
		return fmt.Errorf("tenancy migrate: record failure for %s: %w", migration.Version, err)
	}
	return nil
}

func recordMigrationSuccess(ctx context.Context, db *sql.DB, migration Migration) error {
	if _, err := db.ExecContext(ctx, upsertMigrationApplied,
		migration.Version, migration.Description, migration.Checksum); err != nil {
		return fmt.Errorf("tenancy migrate: record %s: %w", migration.Version, err)
	}
	return nil
}

func lookupMigration(ctx context.Context, db *sql.DB, version string) (MigrationRecord, bool, error) {
	var (
		record    MigrationRecord
		status    string
		appliedAt any
	)
	err := db.QueryRowContext(ctx, `
		SELECT version, description, checksum, status, error, applied_at
		FROM `+schemaVersionsTable+` WHERE version=?`, version).
		Scan(&record.Version, &record.Description, &record.Checksum, &status, &record.Error, &appliedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return MigrationRecord{}, false, nil
	}
	if err != nil {
		return MigrationRecord{}, false, fmt.Errorf("tenancy migrate: read %s: %w", version, err)
	}
	record.Status = MigrationStatus(status)
	record.AppliedAt = parseMigrationTime(appliedAt)
	return record, true, nil
}

const upsertMigrationApplied = `
	INSERT INTO ` + schemaVersionsTable + ` (version, description, checksum, status, error, applied_at)
	VALUES (?, ?, ?, 'applied', '', strftime('%Y-%m-%dT%H:%M:%fZ','now'))
	ON CONFLICT(version) DO UPDATE SET
		description=excluded.description,
		checksum=excluded.checksum,
		status='applied',
		error='',
		applied_at=excluded.applied_at`

const upsertMigrationFailed = `
	INSERT INTO ` + schemaVersionsTable + ` (version, description, checksum, status, error, applied_at)
	VALUES (?, ?, ?, 'failed', ?, strftime('%Y-%m-%dT%H:%M:%fZ','now'))
	ON CONFLICT(version) DO UPDATE SET
		description=excluded.description,
		checksum=excluded.checksum,
		status='failed',
		error=excluded.error,
		applied_at=excluded.applied_at`

func ensureSchemaVersionsTable(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS `+schemaVersionsTable+` (
			version TEXT PRIMARY KEY,
			description TEXT NOT NULL DEFAULT '',
			checksum TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'applied',
			error TEXT NOT NULL DEFAULT '',
			applied_at DATETIME DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		)`); err != nil {
		return fmt.Errorf("tenancy migrate: ensure schema registry: %w", err)
	}

	rows, err := db.QueryContext(ctx, `PRAGMA table_info(`+schemaVersionsTable+`)`)
	if err != nil {
		return fmt.Errorf("tenancy migrate: inspect schema registry: %w", err)
	}
	present := make(map[string]struct{})
	for rows.Next() {
		var (
			columnID  int
			name      string
			columnTyp string
			notNull   int
			defaultV  sql.NullString
			primaryK  int
		)
		if err := rows.Scan(&columnID, &name, &columnTyp, &notNull, &defaultV, &primaryK); err != nil {
			_ = rows.Close()
			return fmt.Errorf("tenancy migrate: inspect schema registry: %w", err)
		}
		present[name] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}

	columns := []struct {
		name string
		ddl  string
	}{
		{"description", `ALTER TABLE ` + schemaVersionsTable + ` ADD COLUMN description TEXT NOT NULL DEFAULT ''`},
		{"checksum", `ALTER TABLE ` + schemaVersionsTable + ` ADD COLUMN checksum TEXT NOT NULL DEFAULT ''`},
		{"status", `ALTER TABLE ` + schemaVersionsTable + ` ADD COLUMN status TEXT NOT NULL DEFAULT 'applied'`},
		{"error", `ALTER TABLE ` + schemaVersionsTable + ` ADD COLUMN error TEXT NOT NULL DEFAULT ''`},
	}
	for _, column := range columns {
		if _, ok := present[column.name]; ok {
			continue
		}
		if _, err := db.ExecContext(ctx, column.ddl); err != nil {
			return fmt.Errorf("tenancy migrate: add %s column: %w", column.name, err)
		}
	}
	return nil
}

func parseMigrationTime(value any) time.Time {
	var text string
	switch typed := value.(type) {
	case time.Time:
		return typed
	case string:
		text = typed
	case []byte:
		text = string(typed)
	default:
		return time.Time{}
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05",
	} {
		if parsed, err := time.Parse(layout, text); err == nil {
			return parsed.UTC()
		}
	}
	return time.Time{}
}
