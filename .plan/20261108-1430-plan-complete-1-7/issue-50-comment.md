## Verification evidence (main)
- Secure SFTP manifest upload path implemented in MVP1 deployment pipeline (SFTP/manifest handling in arsy deployments/servers flow).
- Verified on current main workspace: `go build ./...` passes; `go test ./arsy/deployments/... ./arsy/servers/... ./arsy/httpapi/...` pass.
- Related deployment/worker commits present in history (e.g. deployment execution pipeline and deterministic tests).
Issue closed per MVP1 acceptance reconciliation.