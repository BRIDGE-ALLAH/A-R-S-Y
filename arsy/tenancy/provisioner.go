package tenancy

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ErrProvisionKeyConflict is returned when an idempotency key was already used
// for a different tenant.
var ErrProvisionKeyConflict = errors.New("tenancy: idempotency key already used for another tenant")

// ErrProvisionFailed is returned when a previously attempted provisioning ended
// in a permanent failure.
var ErrProvisionFailed = errors.New("tenancy: provisioning failed permanently")

// ProvisionState is the recorded state of a provisioning operation.
type ProvisionState string

const (
	ProvisionPending         ProvisionState = "pending"
	ProvisionInProgress      ProvisionState = "in_progress"
	ProvisionSucceeded       ProvisionState = "success"
	ProvisionFailedRetryable ProvisionState = "failed_retryable"
)

// ProvisionRecord is one provisioning operation in a ProvisionRegistry.
type ProvisionRecord struct {
	Key           string
	TenantID      string
	State         ProvisionState
	ResultSummary string
}

// ProvisionRegistry durably tracks provisioning operations by idempotency key.
// Implementations must be safe for concurrent use.
type ProvisionRegistry interface {
	// LookupProvision returns the record for key and whether it exists.
	LookupProvision(ctx context.Context, key string) (ProvisionRecord, bool, error)
	// SaveProvision creates or updates the record for its key.
	SaveProvision(ctx context.Context, record ProvisionRecord) error
}

// MemoryProvisionRegistry is an in-process ProvisionRegistry. It survives
// repeated calls within one process but not a restart; use a durable registry
// (e.g. NewOutboxProvisionRegistry) when provisioning must resume after restart.
type MemoryProvisionRegistry struct {
	mu      sync.Mutex
	records map[string]ProvisionRecord
}

// NewMemoryProvisionRegistry creates an empty in-process registry.
func NewMemoryProvisionRegistry() *MemoryProvisionRegistry {
	return &MemoryProvisionRegistry{records: make(map[string]ProvisionRecord)}
}

// LookupProvision implements ProvisionRegistry.
func (r *MemoryProvisionRegistry) LookupProvision(_ context.Context, key string) (ProvisionRecord, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	record, ok := r.records[key]
	return record, ok, nil
}

// SaveProvision implements ProvisionRegistry.
func (r *MemoryProvisionRegistry) SaveProvision(_ context.Context, record ProvisionRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.records == nil {
		r.records = make(map[string]ProvisionRecord)
	}
	r.records[record.Key] = record
	return nil
}

// ProvisionOption configures a Provision call.
type ProvisionOption struct {
	// IdempotencyKey uniquely identifies a provisioning intent. Repeated Provision
	// calls with the same key are idempotent: no duplicate tenant/database is created.
	// If empty, tenantID is used as the key.
	IdempotencyKey string
	// Migrations overrides the Provisioner's default migration set.
	Migrations []Migration
	// Registry overrides the Provisioner's registry for this call.
	Registry ProvisionRegistry
}

// Provisioner creates a tenant database and applies its initial migrations.
// Provisioning is idempotent per idempotency key and only reports success after
// migrations and a datastore health check both succeed.
type Provisioner struct {
	provider Provider
	runner   *MigrationRunner
	registry ProvisionRegistry
}

// NewProvisioner creates a Provisioner using the given provider and migrations.
// It uses an in-process idempotency registry by default.
func NewProvisioner(provider Provider, migrations []Migration) *Provisioner {
	return &Provisioner{
		provider: provider,
		runner:   NewMigrationRunner(migrations),
		registry: NewMemoryProvisionRegistry(),
	}
}

// WithRegistry replaces the provisioning registry, enabling durable idempotency
// across process restarts. It returns the Provisioner for chaining.
func (p *Provisioner) WithRegistry(registry ProvisionRegistry) *Provisioner {
	if registry != nil {
		p.registry = registry
	}
	return p
}

// Provision opens (or creates) the tenant database, applies pending migrations,
// and verifies datastore health. Repeated calls with the same IdempotencyKey are
// idempotent. Returns ErrInvalidTenantID for a blank tenantID,
// ErrProvisionKeyConflict if the key belongs to another tenant, and
// ErrProvisionFailed if a previous attempt failed permanently.
func (p *Provisioner) Provision(ctx context.Context, tenantID string, opts ...ProvisionOption) error {
	if tenantID == "" {
		return ErrInvalidTenantID
	}

	key := tenantID
	migrations := p.runner.migrations
	registry := p.registry
	if len(opts) > 0 {
		if opts[0].IdempotencyKey != "" {
			key = opts[0].IdempotencyKey
		}
		if opts[0].Migrations != nil {
			migrations = opts[0].Migrations
		}
		if opts[0].Registry != nil {
			registry = opts[0].Registry
		}
	}

	if registry != nil {
		record, found, err := registry.LookupProvision(ctx, key)
		if err != nil {
			return fmt.Errorf("tenancy provision: lookup key %q: %w", key, err)
		}
		if found {
			if record.TenantID != tenantID {
				return fmt.Errorf("%w: key %q belongs to tenant %q", ErrProvisionKeyConflict, key, record.TenantID)
			}
			if record.State == ProvisionSucceeded {
				return nil
			}
		}
		if err := registry.SaveProvision(ctx, ProvisionRecord{
			Key: key, TenantID: tenantID, State: ProvisionInProgress,
		}); err != nil {
			return fmt.Errorf("tenancy provision: mark in progress for %q: %w", tenantID, err)
		}
	}

	ds, err := p.provider.Open(ctx, tenantID)
	if err != nil {
		p.recordFailure(ctx, registry, key, tenantID, fmt.Sprintf("open datastore: %v", err))
		return fmt.Errorf("tenancy provision: open datastore for %q: %w", tenantID, err)
	}

	if err := NewMigrationRunner(migrations).Migrate(ctx, ds); err != nil {
		p.recordFailure(ctx, registry, key, tenantID, fmt.Sprintf("migrate: %v", err))
		return fmt.Errorf("tenancy provision: migrate %q: %w", tenantID, err)
	}
	if err := ds.Ping(ctx); err != nil {
		p.recordFailure(ctx, registry, key, tenantID, fmt.Sprintf("health check: %v", err))
		return fmt.Errorf("tenancy provision: health check %q: %w", tenantID, err)
	}

	if registry != nil {
		if err := registry.SaveProvision(ctx, ProvisionRecord{
			Key: key, TenantID: tenantID, State: ProvisionSucceeded,
			ResultSummary: "migrations applied; datastore healthy",
		}); err != nil {
			return fmt.Errorf("tenancy provision: record success for %q: %w", tenantID, err)
		}
	}
	return nil
}

func (p *Provisioner) recordFailure(ctx context.Context, registry ProvisionRegistry, key, tenantID, summary string) {
	if registry == nil {
		return
	}
	_ = registry.SaveProvision(ctx, ProvisionRecord{
		Key: key, TenantID: tenantID, State: ProvisionFailedRetryable, ResultSummary: summary,
	})
}
