# Spec (P0-P1 completion)

## REQ-001 Hotfix absolute tenant root
Default/relative `ARSY_DATA_DIR` and tenant roots resolve to absolute paths before BaaS directory validation/use. API: internal config/startup; `examples/base` uses validated `TenantsRoot`. AC: `serve/doctor` no longer fail with "tenant data root must be an absolute path" for default config.

## REQ-002 Console copy
A-R-S-Y console login uses A-R-S-Y branding text (not "PocketBase superuser account"). Embedded dist rebuilt. AC: `/arsy` login subtitle reflects A-R-S-Y.

## REQ-003 #84 Tenant membership registry
Control-plane collection `baas_tenant_memberships` (or equivalent idempotent migration): tenant, principal/user, role `owner|admin|developer|viewer`, status `active|inactive|invited?` (minimal implemented set documented), unique tenant+principal. Membership changes auditable where feasible. AC: migration/validation tests pass.

## REQ-004 #138/#91 Tenant admin/membership API + authz
Authenticated control-plane endpoints for tenant list/view/rename/suspend/resume-safe operations and membership add/list/change role/deactivate/remove. Pagination and validation. Server-side membership/role checks on every protected tenant route; client tenant selector checked against principal membership. Allow-listed DTOs only. AC: 401/403/404/409 contracts; inactive tenant states fail closed.

## REQ-005 #92 Adversarial isolation
Tests cover cross-tenant access, missing/wrong selectors, deactivated membership, suspended/deleting tenant, path/secret leakage. AC: tests pass; no leakage.

## REQ-006 #89 Idempotent provisioning
Provisioning has idempotency/safe retry behavior and explicit state handling; tenant active only after datastore/schema health and migrations succeed; no partial active exposure. AC: duplicate/retry tests pass.

## REQ-007 #131 Static persistence mapper
Typed create/get/update/delete for registered static model through explicit adapter; field mapping/conversion errors typed and field-addressable; missing/constraint errors normalized; isolated tests; PocketBase APIs unchanged. AC: criteria verified with tests.

## NFR
- `go test ./arsy/... -count=1` PASS for touched scope at minimum; full arsy PASS before claiming completion.
- `dashboard` build PASS when console source changes.
- No production deploy. Release actions only with explicit authorization.
