## Verification evidence (main)
- Docker Compose deployment execution implemented in deployment pipeline (worker + executor + progress endpoints).
- Deploy trigger/progress endpoints: 11a21c3; dashboard deploy/progress UI: 541d2f3.
- Verified: `go build ./...` passes; `go test ./arsy/deployments/... ./arsy/httpapi/... ./arsy/servers/...` pass.
Issue closed per MVP1 acceptance reconciliation.