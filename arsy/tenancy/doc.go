// Package tenancy implements the BaaS multi-tenant datastore layer for A-R-S-Y.
//
// The tenancy package owns:
//   - Engine-neutral datastore interfaces (Datastore, Provider, HealthChecker).
//   - SQLite-per-tenant factory.
//   - Tenant context resolution and routing middleware.
//   - Tenant lifecycle: provisioning, suspension, deletion.
//
// It depends on arsy/config for directory paths and arsy/migrations for the
// control-plane schema. It does NOT import upstream PocketBase internals.
package tenancy
