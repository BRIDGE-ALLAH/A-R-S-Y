package tenancy_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
)

func newProvisionTestProvider(t *testing.T) *tenancy.SQLiteProvider {
	t.Helper()
	dirs, err := config.EnsureBaaSDirectories(t.TempDir(), "")
	if err != nil {
		t.Fatalf("setup dirs: %v", err)
	}
	return tenancy.NewSQLiteProvider(dirs)
}

func openControlDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "control.db"))
	if err != nil {
		t.Fatalf("open control db: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func countUp(counter *atomic.Int32, createSQL string) func(*sql.DB) error {
	return func(db *sql.DB) error {
		counter.Add(1)
		_, err := db.Exec(createSQL)
		return err
	}
}

func TestProvisioner_IdempotentSameKey(t *testing.T) {
	provider := newProvisionTestProvider(t)
	defer provider.Close()
	ctx := context.Background()

	var runs atomic.Int32
	migrations := []tenancy.Migration{{
		Version: "0001", Checksum: "prov-0001",
		Up: countUp(&runs, `CREATE TABLE IF NOT EXISTS prov_items (id TEXT PRIMARY KEY)`),
	}}
	p := tenancy.NewProvisioner(provider, migrations)

	if err := p.Provision(ctx, "tenant-acme", tenancy.ProvisionOption{IdempotencyKey: "key-acme"}); err != nil {
		t.Fatalf("first Provision: %v", err)
	}
	if err := p.Provision(ctx, "tenant-acme", tenancy.ProvisionOption{IdempotencyKey: "key-acme"}); err != nil {
		t.Fatalf("second Provision same key: %v", err)
	}
	if got := runs.Load(); got != 1 {
		t.Fatalf("migration runs = %d, want 1 (idempotent)", got)
	}
}

func TestProvisioner_KeyConflictDifferentTenant(t *testing.T) {
	provider := newProvisionTestProvider(t)
	defer provider.Close()
	ctx := context.Background()
	p := tenancy.NewProvisioner(provider, nil)

	if err := p.Provision(ctx, "tenant-acme", tenancy.ProvisionOption{IdempotencyKey: "shared-key"}); err != nil {
		t.Fatalf("first Provision: %v", err)
	}
	err := p.Provision(ctx, "tenant-other", tenancy.ProvisionOption{IdempotencyKey: "shared-key"})
	if !errors.Is(err, tenancy.ErrProvisionKeyConflict) {
		t.Fatalf("Provision error = %v, want ErrProvisionKeyConflict", err)
	}
}

func TestProvisioner_FailedMigrationRecoverable(t *testing.T) {
	provider := newProvisionTestProvider(t)
	defer provider.Close()
	ctx := context.Background()

	upstream := errors.New("boom")
	failing := tenancy.NewProvisioner(provider, []tenancy.Migration{{
		Version: "0001", Checksum: "recover-0001",
		Up: func(*sql.DB) error { return upstream },
	}})
	err := failing.Provision(ctx, "tenant-recover", tenancy.ProvisionOption{IdempotencyKey: "recover-key"})
	if !errors.Is(err, upstream) {
		t.Fatalf("Provision error = %v, want %v", err, upstream)
	}

	var runs atomic.Int32
	recovered := tenancy.NewProvisioner(provider, []tenancy.Migration{{
		Version: "0001", Checksum: "recover-0001",
		Up: countUp(&runs, `CREATE TABLE IF NOT EXISTS recover_items (id TEXT PRIMARY KEY)`),
	}})
	if err := recovered.Provision(ctx, "tenant-recover", tenancy.ProvisionOption{IdempotencyKey: "recover-key"}); err != nil {
		t.Fatalf("retry Provision: %v", err)
	}
	if runs.Load() != 1 {
		t.Fatalf("migration runs after retry = %d, want 1", runs.Load())
	}

	ds, err := provider.Open(ctx, "tenant-recover")
	if err != nil {
		t.Fatalf("reopen tenant: %v", err)
	}
	if err := ds.Ping(ctx); err != nil {
		t.Fatalf("tenant unhealthy after recovery: %v", err)
	}
}

// fakeDatastore lets tests force a health-check failure.
type fakeDatastore struct {
	db      *sql.DB
	pingErr error
}

func (d *fakeDatastore) DB() *sql.DB                { return d.db }
func (d *fakeDatastore) TenantID() string           { return "fake" }
func (d *fakeDatastore) Ping(context.Context) error { return d.pingErr }
func (d *fakeDatastore) Close() error               { return nil }

type fakeProvider struct {
	ds tenancy.Datastore
}

func (p *fakeProvider) Open(context.Context, string) (tenancy.Datastore, error) { return p.ds, nil }
func (p *fakeProvider) CloseTenant(context.Context, string) error               { return nil }
func (p *fakeProvider) Close() error                                            { return nil }

func TestProvisioner_HealthGateBlocksSuccess(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "health.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	healthErr := errors.New("ping failed")
	provider := &fakeProvider{ds: &fakeDatastore{db: db, pingErr: healthErr}}
	registry := tenancy.NewMemoryProvisionRegistry()
	p := tenancy.NewProvisioner(provider, []tenancy.Migration{{
		Version: "0001", Checksum: "health-0001",
		Up: func(db *sql.DB) error {
			_, err := db.Exec(`CREATE TABLE IF NOT EXISTS health_items (id TEXT PRIMARY KEY)`)
			return err
		},
	}}).WithRegistry(registry)

	err = p.Provision(context.Background(), "tenant-health", tenancy.ProvisionOption{IdempotencyKey: "health-key"})
	if !errors.Is(err, healthErr) {
		t.Fatalf("Provision error = %v, want %v", err, healthErr)
	}
	record, found, err := registry.LookupProvision(context.Background(), "health-key")
	if err != nil || !found {
		t.Fatalf("registry lookup: found=%v err=%v", found, err)
	}
	if record.State != tenancy.ProvisionFailedRetryable {
		t.Fatalf("state = %q, want %q", record.State, tenancy.ProvisionFailedRetryable)
	}
}

func TestProvisioner_DurableRegistrySurvivesRestart(t *testing.T) {
	provider := newProvisionTestProvider(t)
	defer provider.Close()
	ctx := context.Background()
	control := openControlDB(t)

	upstream := errors.New("partial")
	firstRegistry, err := tenancy.NewOutboxProvisionRegistry(ctx, control)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	first := tenancy.NewProvisioner(provider, []tenancy.Migration{{
		Version: "0001", Checksum: "durable-0001",
		Up: func(*sql.DB) error { return upstream },
	}}).WithRegistry(firstRegistry)
	if err := first.Provision(ctx, "tenant-durable", tenancy.ProvisionOption{IdempotencyKey: "durable-key"}); !errors.Is(err, upstream) {
		t.Fatalf("first Provision error = %v, want %v", err, upstream)
	}

	// Simulate a process restart: new registry instance on the same control DB.
	secondRegistry, err := tenancy.NewOutboxProvisionRegistry(ctx, control)
	if err != nil {
		t.Fatalf("reopen registry: %v", err)
	}
	record, found, err := secondRegistry.LookupProvision(ctx, "durable-key")
	if err != nil || !found {
		t.Fatalf("lookup after restart: found=%v err=%v", found, err)
	}
	if record.State != tenancy.ProvisionFailedRetryable || record.TenantID != "tenant-durable" {
		t.Fatalf("record after restart = %+v", record)
	}

	var runs atomic.Int32
	second := tenancy.NewProvisioner(provider, []tenancy.Migration{{
		Version: "0001", Checksum: "durable-0001",
		Up: countUp(&runs, `CREATE TABLE IF NOT EXISTS durable_items (id TEXT PRIMARY KEY)`),
	}}).WithRegistry(secondRegistry)
	if err := second.Provision(ctx, "tenant-durable", tenancy.ProvisionOption{IdempotencyKey: "durable-key"}); err != nil {
		t.Fatalf("retry after restart: %v", err)
	}
	if runs.Load() != 1 {
		t.Fatalf("migration runs = %d, want 1", runs.Load())
	}
	record, found, err = secondRegistry.LookupProvision(ctx, "durable-key")
	if err != nil || !found {
		t.Fatalf("lookup success: found=%v err=%v", found, err)
	}
	if record.State != tenancy.ProvisionSucceeded {
		t.Fatalf("state = %q, want %q", record.State, tenancy.ProvisionSucceeded)
	}

	// A third call with the same key is a no-op.
	if err := second.Provision(ctx, "tenant-durable", tenancy.ProvisionOption{IdempotencyKey: "durable-key"}); err != nil {
		t.Fatalf("third Provision: %v", err)
	}
	if runs.Load() != 1 {
		t.Fatalf("migration re-ran after success: %d", runs.Load())
	}
}

func TestProvisioner_DifferentKeysIndependent(t *testing.T) {
	provider := newProvisionTestProvider(t)
	defer provider.Close()
	ctx := context.Background()
	p := tenancy.NewProvisioner(provider, nil)

	if err := p.Provision(ctx, "tenant-a", tenancy.ProvisionOption{IdempotencyKey: "key-a"}); err != nil {
		t.Fatalf("Provision tenant-a: %v", err)
	}
	if err := p.Provision(ctx, "tenant-b", tenancy.ProvisionOption{IdempotencyKey: "key-b"}); err != nil {
		t.Fatalf("Provision tenant-b: %v", err)
	}
	dsA, err := provider.Open(ctx, "tenant-a")
	if err != nil {
		t.Fatalf("Open tenant-a: %v", err)
	}
	dsB, err := provider.Open(ctx, "tenant-b")
	if err != nil {
		t.Fatalf("Open tenant-b: %v", err)
	}
	if dsA.TenantID() == dsB.TenantID() {
		t.Fatal("different tenants share an identity")
	}
}
