 package tenancy

 import (
 	"context"
 	"database/sql"
 )

 // Datastore is the engine-neutral handle for a single tenant's database.
 // Each tenant gets exactly one Datastore; implementations must be safe for
 // concurrent use by multiple goroutines.
 type Datastore interface {
 	// DB returns the underlying *sql.DB for the tenant. Callers must not close it.
 	DB() *sql.DB

 	// Ping verifies connectivity to the tenant database.
 	Ping(ctx context.Context) error

 	// Close releases all resources. After Close returns, DB() must not be used.
 	Close() error

 	// TenantID returns the immutable identifier of the owning tenant.
 	TenantID() string
 }

 // Provider manages the lifecycle of Datastore instances.
 // Implementations are responsible for creation, caching and cleanup.
 type Provider interface {
 	// Open returns or creates the Datastore for the given tenantID.
 	// It is idempotent: multiple calls for the same ID return the same handle.
 	Open(ctx context.Context, tenantID string) (Datastore, error)

 	// Close closes all open datastores and releases internal resources.
 	// After Close, Open must return an error.
 	Close() error
 }

 // HealthStatus summarises the result of a health probe.
 type HealthStatus struct {
 	TenantID string
 	Healthy  bool
 	Err      error
 }

 // HealthChecker can report tenant database health.
 type HealthChecker interface {
 	Health(ctx context.Context, tenantID string) HealthStatus
 }
