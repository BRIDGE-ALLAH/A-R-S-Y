# Spec

- REQ-001: Engine routes require authentication. Superuser: full access as before. Authenticated member of tenant (or equivalent safe membership check implemented in this slice): may perform engine CRUD/query for that tenant only.
- REQ-002: Unauthorized/non-member access to tenant-scoped engine data is denied with safe status (403/404 as chosen consistently).
- REQ-003: Tenant create/list/provision remain superuser-only.
- REQ-004: E2E test exercises real HTTP handlers + control-plane DB + tenant SQLite: create tenant, provision, create/read/update/query/delete records, verify isolation/persistence.
- REQ-005: Existing `go test ./arsy/...` remains green.
