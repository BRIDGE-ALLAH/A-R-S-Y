# Report

Status: COMPLETE (scoped slice)

## Tasks
- Engine runtime bootstrap added: `arsy/engine/bootstrap` registers starter `base.company` (`base_company`) and applies idempotent tenant SQLite schema (creates table + starter columns; preserves existing data). Identifier validation and tenant path validation enforced.
- Entrypoint wired: `examples/base/main.go` resolves absolute data dir, calls `httpapi.RegisterWithEngine` with `bootstrap.DefaultRegistry()` and `TenantsRoot`, and uses `config.EnsureBaaSDirectories` for runtime/doctor setup.
- Superuser tenant API added: `GET/POST /api/arsy/v1/tenants`, `POST /api/arsy/v1/tenants/{slug}/provision` (idempotent schema apply). Safe summaries only; invalid slugs rejected; duplicate slug returns conflict.
- Integration tests added: `arsy/httpapi/tenants_test.go` covers auth, create/list/provision, and live engine CRUD for provisioned tenant (`base.company`). Existing engine tests remain green.
- Dashboard: attempted minimal engine page caused a malformed generated block and was fully reverted to the indexed `dashboard/src/main.js`; no dist changes committed/generated beyond prior state.

## Evidence
- `go test ./arsy/httpapi/ -run 'TestTenant|TestEngine' -count=1 -v` → PASS (9 tests).
- `go build ./...` → `BUILD_EXIT=0`.
- `go test ./arsy/... -count=1` → PASS (all arsy packages; `arsy/engine/bootstrap` has no tests yet).

## Residual risks
- Tenant endpoints are superuser-gated only; no per-tenant user membership/authz or scoped tokens.
- Schema bootstrap covers starter `base.company` columns only; arbitrary dynamic model schema lifecycle/migrations still missing.
- Tenant create sets status `active` and provisions schema; provisioning failure marks tenant `failed` (safe) but may leave partial DB state; deeper transactional/provisioning state machine not implemented.
- Dashboard does not yet expose tenant/engine management safely; UI wiring deferred after revert.
- Real managed-host E2E and release/deploy remain blocked by infra/credentials/explicit authorization.
