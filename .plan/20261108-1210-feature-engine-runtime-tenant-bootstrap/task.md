# Task: Engine runtime wiring and tenant bootstrap

- [x] Add engine runtime bootstrap: default model registry (starter `base.company`) and tenant engine schema migrator for registered table-backed models (`arsy/engine/bootstrap`).
- [x] Wire `examples/base/main.go` to `EnsureBaaSDirectories` and `RegisterWithEngine` with runtime registry + tenant root (absolute tenant path).
- [x] Add superuser tenant management endpoints: list tenants, create tenant, provision/apply engine schema for active tenant (idempotent) (`arsy/httpapi/tenants.go`).
- [x] Add/extend HTTP tests for engine runtime wiring behavior and tenant management/provisioning; keep secrets/credentials out of responses.
- [x] Reverted accidental dashboard UI corruption; dashboard unchanged (engine UI deferred to avoid breaking existing Projects/Deployments views). Console dist unchanged.
- [x] Run focused tests, `go build ./...`, `go test ./arsy/...`; update report with evidence and residual risks.
