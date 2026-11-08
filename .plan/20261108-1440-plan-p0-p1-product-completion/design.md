# Design (P0-P1 completion)

## Architecture
- **Control plane (PocketBase/SQLite):** `baas_tenants`, new `baas_tenant_memberships` (and optional provisioning operation record if required by #89) managed via `arsy/migrations/*` with idempotent `ensureCollection` upgrades.
- **Authorization:** server-side helper resolves tenant record + principal membership/role; superuser retains control-plane admin capability. Engine tenant-scoped routes require active tenant + authorized membership/ownership policy (explicit, least privilege). Responses never include internal paths/secret refs.
- **DTO boundary:** allow-listed structs/maps only.
- **Provisioning:** reuse `config.EnsureBaaSDirectories`, `tenancy` provider/bootstrap; make API path idempotent and state-safe (active only on success).
- **Persistence mapper (#131):** explicit adapter over registered model/fields + dbx/PocketBase boundaries; typed errors.

## Decisions
- Keep changes in `arsy/*`, `dashboard/src`, embedded console dist, `examples/base`, migrations; avoid `core/`.
- Prefer additive migrations/collections over destructive schema rewrites.
- Membership roles: `owner, admin, developer, viewer` per backlog; permission matrix implemented minimally and tested.

## Risks
- Existing engine tests assume superuser; update fixtures carefully while adding membership cases.
- BaaS backlog dependencies (#80/#81/#85 etc.) may be partially satisfied; implement minimal safe subset and document residual gaps.
- v1.0.0 already tagged/released; post-release fixes require new patch version/release with explicit approval.
