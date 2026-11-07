# A-R-S-Y BaaS Architecture & Implementation Plan

Status: **Approved baseline — 2026-11-07**

## 1. Architecture decisions

| Area | Decision | Implementation rule |
|---|---|---|
| Database | Hybrid | PocketBase + SQLite remains the control-plane and default tenant datastore initially. PostgreSQL is an optional backend for workloads that need it; first planned workload is AAA/RADIUS accounting. |
| Tenant isolation | Dedicated | Each tenant receives its own database boundary and lifecycle. The control plane stores tenant registry and placement metadata. |
| Product scope | General-purpose BaaS | Build reusable database, auth, API, event, storage, backup and operations primitives. ISP is the first vertical application, not a constraint on the core model. |

### Meaning of dedicated tenancy

Dedicated means **one tenant → one independently managed database**. It does not mean one tenant ID column in shared business tables. The first implementation may place multiple tenant SQLite database files on the same host/process, but they must have independent connections, migration state, backup/restore and access boundaries. This is database isolation, not yet dedicated compute or host-level isolation. Enterprise deployments may later place a tenant database on a dedicated host/cluster.

The control-plane database remains separate and contains tenant identity, lifecycle state, database engine, placement, non-secret connection reference, schema version and operational metadata. It must not contain tenant business records.

### Hybrid database boundary

- Control plane and default BaaS tenant engine: existing PocketBase/SQLite.
- PostgreSQL: optional engine adapter for high-write/relational workloads (initial candidate: RADIUS/AAA); do not add PostgreSQL to every installation by default.
- No cross-database transaction assumptions. Use explicit APIs, idempotent jobs/outbox and reconciliation when data must move between engines.
- Secrets/DSNs are stored through secret references or encrypted configuration, never in logs or API responses.

## 2. Repository baseline and constraints

The repository currently is a Go 1.27 application derived from PocketBase. README.md describes PocketBase as the embedded foundation for SQLite, auth, migrations and HTTP runtime. go.mod contains PocketBase-derived dbx packages and modernc.org/sqlite, but no PostgreSQL driver. A-R-S-Y-owned code is organized under arsy/ with config, HTTP API, migrations, security, servers and deployments packages. A-R-S-Y migrations are registered separately under arsy/migrations; existing PocketBase migrations under migrations/ must remain intact.

The current README explicitly excludes multi-tenant RBAC from MVP-1. Do not retrofit tenancy into the existing server/project/deployment collections in a way that changes MVP-1 semantics. Introduce a separately tested BaaS control-plane and tenant data-plane boundary. Preserve PocketBase upstream behavior and avoid forking its core for this feature.

## 3. Target logical architecture

```
Clients / SDKs
     |
     v
A-R-S-Y HTTP API + AuthN/AuthZ
     |
     +---- Control-plane services ---- PocketBase / SQLite (control.db)
     |       tenant registry, placement, lifecycle, policy,
     |       database catalog, migration status, audit metadata
     |
     +---- Tenant resolver / policy boundary
              |
              +---- Tenant Store interface
                       |
                       +---- Dedicated SQLite/PocketBase tenant DB (default)
                       |
                       +---- PostgreSQL adapter (opt-in workload)
```

- Control plane owns tenant provisioning and routing; tenant data plane owns tenant records.
- A request must resolve an authenticated principal, authorized tenant membership and active tenant placement before obtaining a tenant store.
- Never accept a raw database path, DSN or engine override from a client request.
- Do not share a tenant database handle across tenant identities. Cache handles by immutable tenant/database identity and close them on lifecycle transitions.
- Tenant status must be checked on every request or through a cache with bounded invalidation; suspended/deleting tenants must fail closed.
- Tenant-local migrations run independently and record version/checksum.
- Provisioning, migration, backup and restore are asynchronous/idempotent operations with explicit states.

## 4. Proposed package boundaries

These are proposed new packages; adapt naming to repository conventions during BaaS-001 audit.

- arsy/tenancy: tenant IDs, registry service, lifecycle state machine, tenant context and membership policy.
- arsy/datastore: engine-neutral tenant store/provider interfaces and typed errors.
- arsy/datastore/sqlite: dedicated SQLite tenant database factory, pragmas, connection lifecycle and health.
- arsy/datastore/postgres: optional PostgreSQL adapter; introduced only with a concrete workload and integration tests.
- arsy/tenantapi or equivalent: tenant-scoped API routes; reuse central route registration and auth middleware.
- arsy/operations: provisioning/migration/backup job orchestration, idempotency and audit.
- arsy/migrations: continue to own A-R-S-Y control-plane migrations; tenant schema migrations must have an explicit separate runner/registry.

Avoid prematurely implementing a generic SQL ORM or rewriting PocketBase. Start with narrow interfaces that cover the actual CRUD/query needs.

## 5. Data model (initial proposal)

Control-plane collections/tables:
- baas_tenants: immutable ID, name, slug, status, plan, created/updated timestamps.
- baas_tenant_databases: tenant ID, engine, placement key, database identifier, schema version, status, secret reference, timestamps. No plaintext DSN/password.
- baas_tenant_memberships: tenant ID, principal/user ID, role, status, unique tenant+principal.
- baas_tenant_operations: operation ID, tenant ID, type, idempotency key, state, safe result summary, timestamps.
- baas_audit_events: actor, tenant, action, resource type/id, outcome, timestamp and redacted metadata.

Tenant-local data starts with a minimal metadata/schema-version table. Do not add ISP entities to the generic core schema.

## 6. Security invariants

1. No tenant store can be opened without a validated tenant context and authorization decision.
2. Tenant A cannot read, write, list, count, expand, export, backup, restore or subscribe to Tenant B data.
3. Tenant identifiers from URL/body/header are selectors only; membership is independently checked against authenticated identity.
4. Tenant database paths are generated from server-side IDs under a private data root; reject traversal and symlink escapes.
5. Database credentials and encryption keys never enter logs, errors, DTOs or audit payloads.
6. Provisioning and restore are idempotent, bounded and auditable.
7. Tenant deletion is staged: disable access → backup/retention window → explicit purge.
8. Cross-tenant tests must use at least two real isolated databases and adversarial identifiers.

## 7. Delivery milestones

### BaaS-0 — Architecture and baseline
Audit current PocketBase lifecycle, startup, migrations, auth, routing and data directory. Capture baseline tests and approve ADRs. No behavior changes.

### BaaS-1 — Control-plane tenancy
Add tenant registry, database catalog, membership and operation schemas through A-R-S-Y-owned migrations. Add tenant ID and lifecycle validation.

### BaaS-2 — Storage abstraction and dedicated SQLite
Define a narrow store/provider contract, create isolated SQLite databases under a private tenant data root, validate path safety, configure connection limits/pragmas, health-check and close handles safely.

### BaaS-3 — Tenant request boundary
Resolve principal + membership + tenant status; inject immutable tenant context; route to the correct store. Protect every tenant-scoped API and test cross-tenant denial.

### BaaS-4 — Lifecycle and migrations
Provision, suspend, resume and deprovision tenants through idempotent jobs. Add per-tenant schema migration version/checksum, locking, retries and failure recovery.

### BaaS-5 — Backup, restore and operations
Per-tenant consistent backup, checksum, restore into a new location, integrity check, atomic switch, audit trail, health metrics and operational docs.

### BaaS-6 — Hybrid PostgreSQL capability
Document engine selection and workload criteria. Implement PostgreSQL adapter only behind the same narrow interface, with integration tests and migration semantics. Keep it optional. The first concrete consumer can be AAA/RADIUS; do not couple ISP domain types into BaaS core.

### BaaS-7 — General-purpose BaaS API
Tenant-scoped CRUD/query APIs, pagination/filter limits, auth/RBAC, audit and API contract tests. Add events/jobs/storage only after tenancy boundary is proven.

## 8. Definition of Done

- Two or more tenants have separate database files and independently tracked schema versions.
- Tenant routing is based on authenticated membership, not client-provided tenant ID alone.
- Cross-tenant read/write/list/count/export/backup/restore tests are denied.
- Provisioning and migration are idempotent and recover from interruption.
- Backup can be restored and integrity-checked without overwriting another tenant.
- Control-plane APIs never expose filesystem paths, DSNs or credentials.
- Existing MVP-1 server/project/deployment flows and PocketBase admin behavior remain green.
- go test ./..., race tests for tenancy packages, and CI build checks pass.
- PostgreSQL is optional and no PostgreSQL service is required for a default SQLite deployment.

## 9. Explicit non-goals for the first release

- Do not replace PocketBase or write a database engine.
- Do not force PostgreSQL on all tenants.
- Do not implement shared-table multi-tenancy as the primary model.
- Do not implement dedicated compute/Kubernetes orchestration yet.
- Do not couple generic BaaS collections to ISP billing, PPPoE or RADIUS domain schemas.
- Do not claim HA, zero-downtime migration or point-in-time recovery until tested.

## 10. Issue execution order

See GitHub issues tagged by title prefix [BAAS-###]. Start with BAAS-001 audit. Architecture implementation must follow dependencies; do not start tenant CRUD endpoints before isolation and authorization tests exist.
