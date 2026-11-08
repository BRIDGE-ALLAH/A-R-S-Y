# Spec

## Routes (superuser-only, tenant-scoped)
- `GET /api/arsy/v1/engine/models` — list registered model descriptors (safe metadata).
- `POST /api/arsy/v1/engine/{model}/records` — create; body `{values, allowed_fields}`.
- `GET /api/arsy/v1/engine/{model}/records/{id}` — get one.
- `PATCH /api/arsy/v1/engine/{model}/records/{id}` — update; body `{values, allowed_fields}`.
- `DELETE /api/arsy/v1/engine/{model}/records/{id}` — delete.
- `POST /api/arsy/v1/engine/{model}/query` — query; body `{query, allowed_fields}`.

Tenant resolved via existing tenancy mechanism (e.g. `X-Arsy-Tenant`/slug -> catalog/provider/router); missing/invalid tenant => 400/404/409 per tenancy state.

## Acceptance
- Real CRUD/query against tenant SQLite DB through `persist.Store`.
- 404 unknown model/record, 400/422 unknown/invalid field or body, 409 constraint/tenant state.
- Isolation: record created in tenant A not visible in tenant B.
- Tests prove auth + isolation + CRUD/query + error mapping.
