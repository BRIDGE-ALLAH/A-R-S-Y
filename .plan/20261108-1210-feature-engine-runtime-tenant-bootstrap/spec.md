# Spec

## Requirements
- REQ-001: Runtime can register starter engine model(s) and expose them via `/api/arsy/v1/engine/models` when engine is wired.
- REQ-002: `examples/base/main.go` initializes BaaS tenant directories and registers engine routes with a real registry and tenant root.
- REQ-003: Superuser can create/list `baas_tenants` records and trigger idempotent tenant DB provisioning/schema application for engine models.
- REQ-004: Engine CRUD works against provisioned tenant DB for starter model when superuser provides tenant + `allowed_fields`.
- REQ-005: No secrets leaked; errors are safe; unknown/invalid inputs return explicit 4xx.

## Contracts
- `POST /api/arsy/v1/tenants` body `{slug, display_name, plan?}` -> creates tenant record (status provisioning/active as implemented) and provisions schema when active.
- `GET /api/arsy/v1/tenants` -> list safe tenant summaries.
- `POST /api/arsy/v1/tenants/{slug}/provision` -> idempotent schema/provisioning for tenant DB.

## Acceptance
- New integration tests cover tenant create/list/provision and engine CRUD after provisioning.
- Existing `go test ./arsy/...` remains green.
