# Task: Tenant-scoped engine CRUD API

- [x] Audit engine/persist + tenancy contracts and define safe API surface.
- [x] Implement tenant-scoped engine CRUD HTTP handlers using existing `persist.Store` + `registry` + `tenancy` (explicit allowlist, auth, error mapping, no fake data).
- [x] Register routes and wire config/bootstrap dependency (RegisterWithEngine added; examples/base/main.go still uses zero EngineConfig and is not wired to a runtime registry/tenant root).
- [x] Add integration-style HTTP tests: auth, tenant isolation, CRUD/query, invalid model/fields, status mapping.
- [x] Run focused tests, `go build ./...`, `go test ./arsy/...`.
- [x] Update report with evidence and residual risks.
