# Report — Step 1 CI/deployments regression fix

Status: COMPLETE (local verification; CI checked after push)

## Completed
- [x] Plan folder created (`.plan/20261108-0033-fix-step1-ci-deployments-regression/`).
- [x] Diagnosed root cause: `HasActiveDeployment` treated `queued` as active; `ClaimNextQueuedDeployment` used it, so a sole queued deployment was never claimable.
- [x] `arsy/deployments/jobs.go`: added `HasInProgressDeployment` (preparing, uploading, validating, deploying, verifying). `ClaimNextQueuedDeployment` now skips candidates only when the app has an **in-progress** deployment. `HasActiveDeployment` (includes `queued`) retained for API enqueue gating (#149).

## Verification Evidence
- Pre-fix (local): `go test ./arsy/deployments/... -run "TestClaimNextQueuedDeploymentIsAtomic|TestUpdateDeploymentStateUsesTransitionRules|TestWorkerProcesses|TestWorkerPersists|TestWorkerMarks"` → **FAIL** (5 failures: `jobs_test.go:51,75`, `worker_test.go:33,55,98`).
- Post-fix: `go test ./arsy/deployments/... -count=1` → **ok** `github.com/bridge-allah/a-r-s-y/arsy/deployments 2.111s` (exit 0).
- Post-fix: `go test ./arsy/httpapi/... ./arsy/servers/... ./arsy/config/... ./tests/... -run Docker -count=1` → **ok** (`httpapi 1.475s [no tests to run]`, `servers 1.392s`, `config 1.311s [no tests to run]`, `tests 1.536s`).
- Post-fix: `go test ./arsy/migrations/... ./arsy/engine/... ./arsy/tenancy/... ./arsy/security/... -count=1` → **ok** (all listed packages pass; `arsy/engine/persist` has `[no test files]`).
- Post-fix: `go build ./...` → exit 0.
- CI (post-push): recorded below after `gh run` check.

## Deviations
- Worker claim uses `HasInProgressDeployment` (excludes `queued`) while API still uses `HasActiveDeployment` (includes `queued`). This is intentional to satisfy both #149 enqueue/idempotency protection and the ability to claim/execute a queued job.

## Remaining Risks / Follow-ups
- Full `go test ./...` was not run locally (time); CI is the authoritative full matrix. If CI fails for unrelated packages, record and address in subsequent steps.
- #149 overlap guarantees remain: API blocks enqueue when active (including queued) except idempotent return; worker never claims a queued job for an app with in-progress execution.
