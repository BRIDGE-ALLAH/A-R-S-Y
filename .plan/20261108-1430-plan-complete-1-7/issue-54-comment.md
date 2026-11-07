## Verification evidence (main)
- Background deployment worker lifecycle implemented in `arsy/deployments` (worker processing, timeout/cancellation/deterministic behavior).
- Deterministic worker tests added (0344972) and overlap-prevention fix verified (b867c2c, 71dbc7c).
- Verified: `go test ./arsy/deployments/... -count=1` passes.
Issue closed per MVP1 acceptance reconciliation.