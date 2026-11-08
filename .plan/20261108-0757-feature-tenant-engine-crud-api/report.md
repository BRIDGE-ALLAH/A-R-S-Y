# Report

Status: COMPLETE (scoped slice).

## Evidence
- Removed stale duplicate `arsy/httpapi/routes_engine.go`; consolidated on `RegisterWithEngine(app, cfg, version, EngineConfig)` in `arsy/httpapi/routes.go` (delegates from `Register` with empty `EngineConfig`).
- Added tenant-scoped engine CRUD routes in `arsy/httpapi/engine.go`: `GET /engine/models`, `POST /engine/{model}/records`, `GET|PATCH|DELETE /engine/{model}/records/{id}`, `POST /engine/{model}/query`. All superuser-gated; tenant resolved via `X-Arsy-Tenant` header or `tenant` query; per-tenant SQLite opened at `config.BaaSDirectories.TenantDBPath`; status mapping for suspended/deleting/missing/inactive tenants; writes/query require explicit `allowed_fields`; responses use safe status mapping for persist errors.
- Added `arsy/tenancy/pb_catalog.go` (`PBCatalog`) resolving tenants from `baas_tenants` by slug/id.
- Fixed `arsy/engine/persist/mapper.go` update behavior: if `updated` column is missing, retry without it (previously caused 500 on schemas without `updated`; preserves compatibility with existing tests that include `updated`).
- Added integration tests `arsy/httpapi/engine_test.go`: auth required, model list, CRUD+query, tenant isolation, invalid request/unknown model/unknown allowed field/missing allowed_fields, suspended/deleting/missing tenant status mapping.
- Focused: `go test ./arsy/httpapi/ -run TestEngine -count=1 -v` PASS (6 tests). `go test ./arsy/engine/persist/ -count=1 -v` PASS.
- Full: `go build ./...` exit 0. `go test ./arsy/... -count=1` PASS (all arsy packages).

## Residual
- `examples/base/main.go` still calls `httpapi.Register(...)` with no engine registry/tenant root; runtime engine CRUD is not bootstrapped for a real app instance until wired.
- No dynamic schema/migration/provisioning for arbitrary engine models; tests create tables manually. Production needs model registration + tenant schema lifecycle.
- Tenant auth is superuser + tenant slug/header; not per-tenant user membership/authz.
- Engine route opens tenant SQLite directly (not via `tenancy.SQLiteProvider` lifecycle) to avoid closed-datastore issues; unify lifecycle later.
- Dashboard has no engine CRUD UI; dist unchanged.
- Real managed-host E2E and release/deploy remain blocked by infra/credentials/explicit authorization.
