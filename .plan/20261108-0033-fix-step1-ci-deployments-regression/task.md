# Task — Step 1: Restore CI / Fix deployments regressions (#149-safe)

Status: IN_PROGRESS
Scope: fix (goal step 1)
Goal: `Selesaikan semua 3 backlog MVP1 issue...` (long-run)

## Checklist

- [x] [M] Diagnose 5 regressions in `arsy/deployments` (jobs/worker) caused by overlap-prevention semantics (`HasActiveDeployment` counts `queued`, blocking claim of a sole queued job). Evidence: `go test ./arsy/deployments/... -run <5 tests>` failed locally (see report).
- [x] [M] Fix claim semantics without weakening #149: added `HasInProgressDeployment` (preparing..verifying only) for worker claim selection; kept `HasActiveDeployment` (includes queued) for API enqueue gating.
- [x] [S] Fix test fallout: no httpapi/server fallout; `TestClaimNextQueuedDeploymentEnforcesAtomicity` (worker_test.go, two apps) passes under corrected semantics. #149 API 409/idempotency behavior unchanged.
- [x] [S] Ran targeted/full relevant tests and build; recorded in `report.md`.
- [x] [S] Commit and push; CI status checked/recorded.
- [x] [S] `report.md` populated.
