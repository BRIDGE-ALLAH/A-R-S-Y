package tenancy

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite" // register sqlite driver used by PocketBase

	"github.com/bridge-allah/a-r-s-y/arsy/config"
)

const (
	// defaultMaxHandles bounds how many tenant database handles a provider keeps
	// cached. Eviction is least-recently-used.
	defaultMaxHandles = 64
	// defaultIdleTimeout is how long a cached handle may sit unused before
	// it is evicted on the next Open/CloseIdle call.
	defaultIdleTimeout = 30 * time.Minute
)

// sqliteDatastore implements Datastore for a SQLite file.
type sqliteDatastore struct {
	tenantID string
	db       *sql.DB
	once     sync.Once
	closeErr error

	// lastUsed is guarded by the owning SQLiteProvider mutex.
	lastUsed time.Time
}

// TenantID returns the immutable tenant identifier.
func (s *sqliteDatastore) TenantID() string { return s.tenantID }

// Ping verifies connectivity to the tenant database.
func (s *sqliteDatastore) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

// DB returns the underlying connection pool. Callers must not close it.
func (s *sqliteDatastore) DB() *sql.DB { return s.db }

// Close releases the database resources.
func (s *sqliteDatastore) Close() error {
	s.once.Do(func() { s.closeErr = s.db.Close() })
	return s.closeErr
}

// SQLiteProvider is a Provider that opens one SQLite file per tenant under a
// shared directory managed by config.BaaSDirectories. Handles are cached by
// immutable tenant ID, bounded by maxHandles, and idle-evicted.
type SQLiteProvider struct {
	dirs config.BaaSDirectories

	mu          sync.Mutex
	store       map[string]*sqliteDatastore
	closed      bool
	maxHandles  int
	idleTimeout time.Duration
}

// NewSQLiteProvider constructs a provider rooted at dirs.TenantsRoot.
func NewSQLiteProvider(dirs config.BaaSDirectories) *SQLiteProvider {
	return &SQLiteProvider{
		dirs:        dirs,
		store:       make(map[string]*sqliteDatastore),
		maxHandles:  defaultMaxHandles,
		idleTimeout: defaultIdleTimeout,
	}
}

// SetMaxHandles bounds the number of cached tenant handles. n <= 0 means
// unbounded (not recommended for production).
func (p *SQLiteProvider) SetMaxHandles(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.maxHandles = n
}

// SetIdleTimeout sets how long an unused cached handle is retained.
func (p *SQLiteProvider) SetIdleTimeout(d time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.idleTimeout = d
}

// Open returns or creates the SQLite datastore for tenantID.
func (p *SQLiteProvider) Open(ctx context.Context, tenantID string) (Datastore, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}
	if filepath.Base(tenantID) != tenantID {
		return nil, fmt.Errorf("%w: path components not allowed in tenant ID %q", ErrInvalidTenantID, tenantID)
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, ErrDatastoreClosed
	}
	now := time.Now()
	if ds, ok := p.store[tenantID]; ok {
		ds.lastUsed = now
		return ds, nil
	}
	// Bounded cache: evict idle handles, then the least-recently-used handle.
	p.evictIdleLocked(now.Add(-p.idleTimeout))
	if p.maxHandles > 0 && len(p.store) >= p.maxHandles {
		p.evictOldestLocked()
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

	ds := &sqliteDatastore{tenantID: tenantID, db: db, lastUsed: now}
	p.store[tenantID] = ds
	return ds, nil
}

// CloseTenant closes and evicts the Datastore for tenantID if present.
func (p *SQLiteProvider) CloseTenant(_ context.Context, tenantID string) error {
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

// CloseIdle evicts tenant handles unused for longer than idleAfter.
func (p *SQLiteProvider) CloseIdle(_ context.Context, idleAfter time.Duration) error {
	if idleAfter <= 0 {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return ErrDatastoreClosed
	}
	return p.evictIdleLocked(time.Now().Add(-idleAfter))
}

func (p *SQLiteProvider) evictIdleLocked(cutoff time.Time) error {
	var firstErr error
	for id, ds := range p.store {
		if ds.lastUsed.Before(cutoff) {
			if err := ds.Close(); err != nil && firstErr == nil {
				firstErr = fmt.Errorf("tenancy: close idle tenant %q: %w", id, err)
			}
			delete(p.store, id)
		}
	}
	return firstErr
}

func (p *SQLiteProvider) evictOldestLocked() {
	var oldestID string
	var oldest time.Time
	for id, ds := range p.store {
		if oldestID == "" || ds.lastUsed.Before(oldest) {
			oldestID, oldest = id, ds.lastUsed
		}
	}
	if oldestID == "" {
		return
	}
	_ = p.store[oldestID].Close()
	delete(p.store, oldestID)
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

var _ Provider = (*SQLiteProvider)(nil)
var _ HealthChecker = (*SQLiteProvider)(nil)
var _ IdleCloser = (*SQLiteProvider)(nil)
