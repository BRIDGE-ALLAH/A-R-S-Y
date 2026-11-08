# Task: Per-tenant authz and E2E API flow tests

- [ ] Add tenant-scoped engine authorization: superusers keep full access; authenticated non-superusers may access engine CRUD only for tenants they belong to (membership/owner checks), with safe 401/403/404 behavior.
- [ ] Keep tenant management endpoints superuser-only for this slice; document residual gap.
- [ ] Add membership-aware test fixtures and extend engine/tenant tests for allowed vs denied access; preserve existing superuser tests.
- [ ] Add an in-process end-to-end HTTP flow test covering tenant create/provision + engine CRUD persistence + isolation (real SQLite, no mocks for DB/auth flow where feasible).
- [ ] Add bootstrap unit test for `arsy/engine/bootstrap` schema application.
- [ ] Run focused tests, `go build ./...`, `go test ./arsy/...`; update report with evidence and residual risks.
