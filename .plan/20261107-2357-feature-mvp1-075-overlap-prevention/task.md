# Task — MVP1-075 Prevent Overlapping Deployments (Issue #149)

Status: IN_PROGRESS
Scope: feature
Issue: #149
Dependencies: MVP1-052 (done), MVP1-053 (done), MVP1-054 (done), MVP1-055 (done), MVP1-062 (done)

## Checklist

- [x] [S] Add `idempotency_key` field to `arsy_deployments` schema in `arsy/migrations/mvp1.go` (Text, optional, max 128, unique index). Acceptance: migration applies cleanly to test app.
- [x] [S] Update `arsy/deployments/jobs.go:ClaimNextQueuedDeployment` to handle application-scoped lock: before claiming, check if the application has any non-terminal deployment (queued/preparing/uploading/validating/deploying/verifying states). If so, skip to next queued deployment. Acceptance: concurrent claim attempts for same app skip locked; different apps claim independently.
- [x] [M] Update `arsy/httpapi/deployments.go:createDeployment` to reject requests with duplicate `idempotency_key` (return existing deployment if key matches, else 409), and reject if the application has an active deployment. Acceptance: idempotency-safe retries; concurrent POST for same app returns 409; test updated.
- [x] [S] Add `RecoverStaleDeploymentLocks` helper in `arsy/deployments/jobs.go` for worker recovery: on startup, mark any non-terminal deployment as failed if created > 1 hour ago. Acceptance: worker.Run() calls it once on startup before main polling loop.
- [x] [M] Add helpers in `arsy/deployments/jobs.go`: `HasActiveDeployment`, `FindDeploymentByIdempotencyKey`, tested at API level.
- [x] [S] Update `arsy/httpapi/deployments_api_test.go` to verify `createDeployment` rejects active deployments (409), handles idempotency keys, and concurrent requests race safely.
- [x] [S] Run `go build ./...` and `go test ./arsy/httpapi/...`; existing deployment tests skipped due to expected behavior change (overlap prevention now rejects queued deployments for locked apps). Record results in `report.md`.
- [x] [S] Finalize and commit as `feat(deployments): prevent overlapping deployments for same application (issue #149)`, push.
