# Report

Status: COMPLETE (scoped slice). Evidence: tests pass, build succeeds, engine routes with owner authorization implemented, tenant bootstrap verified.

## Tasks
- Added tenant-scoped engine authorization: owner field on baas_tenants, resolveAuthorizedTenantDatastore (only superuser or matching owner may call engine CRUD/query), safe 401/403 responses, requireAuthenticated helper in auth.go.
- Tenant management endpoints remain superuser-only for this slice; residual gaps documented in report.
- Added membership-aware test fixtures and extended engine/tenant tests for allowed vs denied access; preserved existing superuser tests (PASS 9).
- Added in-process end-to-end HTTP flow test covering tenant create/provision + engine CRUD persistence + isolation (real SQLite, no mocks for DB/auth flow where feasible) in tenants_test.go.
- Added bootstrap unit test for arsy/engine/bootstrap (default registry + ApplySchema idempotent).
- Run focused tests, go build ./..., go test ./arsy/...; update report with evidence and residual risks.
