package tenancy

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite" // register sqlite driver used by PocketBase

	"github.com/bridge-allah/a-r-s-y/arsy/config"
)

// sqliteDatastore implements Datastore for a SQLite file.
type sqliteDatastore struct {
	tenantID string
	db       *sql.DB
	once     sync.Once
	closeErr error
}

func (s *sqliteDatastore) DB() *sql.DB                    { return s.db }
func (s *sqliteDatastore) TenantID() string               { return s.tenantID }
func (s *sqliteDatastore) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
func (s *sqliteDatastore) Close() error {
	s.once.Do(func() { s.closeErr = s.db.Close() })
	return s.closeErr
}

// SQLiteProvider is a Provider that opens one SQLite file per tenant under a
// shared directory managed by config.BaaSDirectories.
type SQLiteProvider struct {
	dirs config.BaaSDirectories

	mu     sync.RWMutex
	store  map[string]*sqliteDatastore
	closed bool
}

// NewSQLiteProvider constructs a provider rooted at dirs.TenantsRoot.
func NewSQLiteProvider(dirs config.BaaSDirectories) *SQLiteProvider {
	return &SQLiteProvider{
		dirs:  dirs,
		store: make(map[string]*sqliteDatastore),
	}
}

// Open returns or creates the SQLite datastore for tenantID.
func (p *SQLiteProvider) Open(ctx context.Context, tenantID string) (Datastore, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}
	if filepath.Base(tenantID) != tenantID {
		return nil, fmt.Errorf("%w: path components not allowed in tenant ID %q", ErrInvalidTenantID, tenantID)
	}

	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return nil, ErrDatastoreClosed
	}
	if ds, ok := p.store[tenantID]; ok {
		p.mu.RUnlock()
		return ds, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, ErrDatastoreClosed
	}
	// Double-check after acquiring write lock.
	if ds, ok := p.store[tenantID]; ok {
		return ds, nil
	}

	dbPath, err := p.dirs.TenantDBPath(tenantID)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("tenancy: open sqlite for tenant %q: %w", tenantID, err)
	}
	db.SetMaxOpenConns(1) // SQLite WAL-mode; single writer is safest default.

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("tenancy: ping sqlite for tenant %q: %w", tenantID, err)
	}

	ds := &sqliteDatastore{tenantID: tenantID, db: db}
	p.store[tenantID] = ds
	return ds, nil
}

// CloseTenant closes and evicts the Datastore for tenantID if present.
func (p *SQLiteProvider) CloseTenant(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return ErrDatastoreClosed
	}
	ds, ok := p.store[tenantID]
	if !ok {
		return nil
	}
	err := ds.Close()
	delete(p.store, tenantID)
	return err
}

// Close closes all open datastores.
func (p *SQLiteProvider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	var firstErr error
	for id, ds := range p.store {
		if err := ds.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("tenancy: close tenant %q: %w", id, err)
		}
		delete(p.store, id)
	}
	return firstErr
}

// Health implements HealthChecker.
func (p *SQLiteProvider) Health(ctx context.Context, tenantID string) HealthStatus {
	ds, err := p.Open(ctx, tenantID)
	if err != nil {
		return HealthStatus{TenantID: tenantID, Healthy: false, Err: err}
	}
	if err := ds.Ping(ctx); err != nil {
		return HealthStatus{TenantID: tenantID, Healthy: false, Err: err}
	}
	return HealthStatus{TenantID: tenantID, Healthy: true}
}
