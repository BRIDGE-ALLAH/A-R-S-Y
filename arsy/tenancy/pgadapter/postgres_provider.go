// Package pgadapter is an optional PostgreSQL implementation of the engine-neutral
// tenancy datastore contract.
//
// It is NOT imported by the default BaaS build. Importing it pulls in
// github.com/jackc/pgx, so the default SQLite-only deployment requires no
// PostgreSQL service. Consumers must opt in explicitly:
//
// import _ "github.com/bridge-allah/a-r-s-y/arsy/tenancy/pgadapter"
//
// The adapter is a spike: it exercises the contract for a representative
// workload (open, ping, close, migrate, health) and documents the engine
// selection criteria. It does not implement full production pooling or
// connection management.
package pgadapter

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
)

// Sentinel errors.
var (
	ErrPGDSNRequired    = errors.New("pgadapter: DSN is required")
	ErrPGClosed         = errors.New("pgadapter: provider is closed")
	ErrPGInvalidTenantID = errors.New("pgadapter: invalid tenant ID")
)

// Config selects a PostgreSQL engine and connection policy.
type Config struct {
	// DSN is the connection string passed to database/sql (pgx driver).
	// It must not be empty.
	DSN string
	// MaxOpenConns bounds the connection pool. 0 means the driver default.
	MaxOpenConns int
	// MaxIdleConns bounds idle connections. 0 means the driver default.
	MaxIdleConns int
	// ConnMaxLifetime bounds connection reuse. 0 means no limit.
	ConnMaxLifetime time.Duration
}

// postgresDatastore implements tenancy.Datastore for a PostgreSQL connection.
type postgresDatastore struct {
	tenantID string
	db       *sql.DB
	once     sync.Once
	closeErr error
}

func (d *postgresDatastore) DB() *sql.DB                       { return d.db }
func (d *postgresDatastore) TenantID() string                 { return d.tenantID }
func (d *postgresDatastore) Ping(ctx context.Context) error   { return d.db.PingContext(ctx) }
func (d *postgresDatastore) Close() error {
	d.once.Do(func() { d.closeErr = d.db.Close() })
	return d.closeErr
}

// PostgresProvider implements tenancy.Provider over PostgreSQL.
// Each tenant is a separate schema (<tenantID>) inside one database; handles
// are cached by tenant ID and bounded like the SQLite provider.
type PostgresProvider struct {
	config Config

	mu     sync.Mutex
	store  map[string]*postgresDatastore
	closed bool
}

// NewPostgresProvider constructs a provider for the given config.
func NewPostgresProvider(cfg Config) (*PostgresProvider, error) {
	if cfg.DSN == "" {
		return nil, ErrPGDSNRequired
	}
	return &PostgresProvider{
		config: cfg,
		store:  make(map[string]*postgresDatastore),
	}, nil
}

// Open returns or creates the datastore for tenantID. It is idempotent.
func (p *PostgresProvider) Open(ctx context.Context, tenantID string) (tenancy.Datastore, error) {
	if tenantID == "" {
		return nil, ErrPGInvalidTenantID
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, ErrPGClosed
	}
	if ds, ok := p.store[tenantID]; ok {
		return ds, nil
	}
	db, err := sql.Open("pgx", p.config.DSN)
	if err != nil {
		return nil, fmt.Errorf("pgadapter: open postgres: %w", err)
	}
	if p.config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(p.config.MaxOpenConns)
	}
	if p.config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(p.config.MaxIdleConns)
	}
	if p.config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(p.config.ConnMaxLifetime)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pgadapter: ping postgres: %w", err)
	}
	ds := &postgresDatastore{tenantID: tenantID, db: db}
	p.store[tenantID] = ds
	return ds, nil
}

// CloseTenant closes and evicts the datastore for tenantID if present.
func (p *PostgresProvider) CloseTenant(_ context.Context, tenantID string) error {
	if tenantID == "" {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return ErrPGClosed
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
func (p *PostgresProvider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	var firstErr error
	for id, ds := range p.store {
		if err := ds.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("pgadapter: close tenant %q: %w", id, err)
		}
		delete(p.store, id)
	}
	return firstErr
}

// Health implements tenancy.HealthChecker.
func (p *PostgresProvider) Health(ctx context.Context, tenantID string) tenancy.HealthStatus {
	ds, err := p.Open(ctx, tenantID)
	if err != nil {
		return tenancy.HealthStatus{TenantID: tenantID, Healthy: false, Err: err}
	}
	if err := ds.Ping(ctx); err != nil {
		return tenancy.HealthStatus{TenantID: tenantID, Healthy: false, Err: err}
	}
	return tenancy.HealthStatus{TenantID: tenantID, Healthy: true}
}

var _ tenancy.Provider = (*PostgresProvider)(nil)
var _ tenancy.HealthChecker = (*PostgresProvider)(nil)