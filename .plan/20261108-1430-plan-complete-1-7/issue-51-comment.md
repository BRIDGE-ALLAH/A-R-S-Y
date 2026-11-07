## Verification evidence (main)
- Compose validation/execution path implemented in deployment flow (validation + worker execution).
- Verified on current main workspace: `go build ./...` passes; `go test ./arsy/deployments/... ./arsy/servers/... ./arsy/httpapi/...` pass.
- SSH test server/readiness support added for deterministic validation paths (commits 3742bd1, 0344972).
Issue closed per MVP1 acceptance reconciliation.