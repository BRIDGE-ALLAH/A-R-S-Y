## Verification evidence (main)
- Persistent deployment job claiming implemented in `arsy/deployments` (queue/claim semantics).
- Overlap-prevention regression fix preserved guarantees (#149): commits b867c2c, 71dbc7c.
- Verified: `go test ./arsy/deployments/... -count=1` passes (ok).
Issue closed per MVP1 acceptance reconciliation.