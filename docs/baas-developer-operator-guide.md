# BaaS Developer & Operator Guide

This guide documents the general-purpose BaaS foundation shipped in `arsy/tenancy`.
It is complementary to [installation.md](installation.md) (how to run A-R-S-Y) and
[server-requirements.md](server-requirements.md) (target host hardening).

---

## 1. Tenant model

A tenant is an isolated SQLite database under the private tenant data root
(`<ARSY_DATA_DIR>/tenants` by default). The control plane keeps a registry in
`baas_tenants`; each tenant owns exactly one datastore handle.

- **Dedicated database, not dedicated compute.** Every tenant gets its own
  SQLite file and its own `*sql.DB` handle. A tenant does not get its own
  process, port, or control-plane instance.
- **No shared-table tenancy.** The datastore contract is per-tenant: a handle
  opened for tenant A is never reused for tenant B.

## 2. Directory layout

| Path | Owner | Purpose |
| --- | --- | --- |
| `<ARSY_DATA_DIR>/tenants` | service account, mode `0700` | Per-tenant SQLite files (`<tenantID>.db`) |
| `<ARSY_DATA_DIR>/backups` | service account | Tenant and control-plane backups |
| `<ARSY_DATA_DIR>/outbox` | service account | Durable outbox queue |

`config.EnsureBaaSDirectories` creates the layout at startup and refuses a
relative tenant root that resolves outside the data directory.

## 3. Provisioning a tenant

Provisioning is idempotent: calling it twice for the same tenant applies only
pending migrations.

```go
provider := tenancy.NewSQLiteProvider(dirs)
runner := tenancy.NewMigrationRunner(tenantMigrations)
p := tenancy.NewProvisioner(provider, runner)
if err := p.Provision(ctx, tenantID); err != nil {
    // handle
}
```

The runner records every applied version in `arsy_schema_versions` with a
description, checksum, status and applied timestamp. A version that failed is
recorded as `failed` and retried on the next run; it is never treated as applied.

## 4. Datastore contract

```go
type Datastore interface {
    DB() *sql.DB
    Ping(ctx context.Context) error
    Close() error
    TenantID() string
}
```

- `Open` is idempotent: repeated calls for one tenant return the same handle.
- Handles are cached, bounded by `maxHandles` (default 64) and idle-evicted
  after `idleTimeout` (default 30 minutes).
- `CloseTenant` closes and evicts a handle; reopening a tenant creates a new
  handle. `Close` closes every open handle and rejects further opens.

## 5. Routing requests

`tenancy.Router` re-reads tenant state from the catalog on every call, so a
stale request context cannot keep access to a suspended or deleted tenant:

| Status | Result |
| --- | --- |
| `active` | Opens the tenant datastore |
| `suspended` | `ErrTenantSuspended` |
| `deleting` | `ErrTenantDeleting` |
| otherwise | `ErrTenantNotActive` |

## 6. Lifecycle

`provisioning → active ↔ suspended → deleting → removed`. Transitions are
validated by `tenancy.ValidateTransition`. `BeginDeletion` evicts the datastore
handle; the control-plane record is removed by a separate cleanup step.

## 7. Engine policy

- **Default engine is SQLite** (`modernc.org/sqlite`). No PostgreSQL service is
  required to build, test or run the default deployment.
- The datastore interface is engine-neutral. An optional PostgreSQL adapter
  is a candidate consumer of this contract, not a dependency of the BaaS core.
- Per-tenant handles use `db.SetMaxOpenConns(1)`; SQLite serialises writers.

## 8. Backup and restore

- `tenancy.Backup` produces a checksummed backup of a tenant database.
- `tenancy.Restore` restores a backup and reassigns the tenant database path.
- Backups are integrity-verified; a checksum mismatch returns
  `tenancy.ErrInvalidChecksum`.

## 9. Outbox and reconciliation

- `tenancy.InitOutbox` creates the `arsy_outbox` table in the control-plane
  database. Records are created durably before an operation is applied.
- A reconciliation worker drains the outbox and retries failed, retryable
  operations. Permanent failures are recorded as `failed_permanent`.

## 10. Health and observability

- `tenancy.HealthChecker.Health` pings a tenant datastore and reports
  `HealthStatus{Healthy, Err}`.
- `tenancy.SQLiteProvider.Stats` reports open handles and the cache bound
  without exposing paths or secrets.

## 11. Local development

```sh
# SQLite-only default
go test ./arsy/tenancy/...
go build -o .builds/arsy ./examples/base
./.builds/arsy serve
```

Tests use isolated temp directories and never touch production data. To
provision two tenants locally, run the provisioner twice with distinct IDs:

```go
_ = p.Provision(ctx, "tenant-alpha")
_ = p.Provision(ctx, "tenant-beta")
```

## 12. Security model

- Tenant data root is `0700` and never world-readable.
- Tenant IDs are validated against path traversal; a tenant ID may not contain
  path components.
- Secrets are referenced by `tenancy.SecretRef`; raw values never enter the
  tenancy package.
- Backups are stored under the private data root and are not published.