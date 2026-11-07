# Report — MVP1-062 Deploy Action and Progress View

Status: COMPLETE (backend only; frontend deferred)

## Completed

- `arsy/httpapi/deployments.go:createDeployment` — POST handler validates stored manifest, generates deployment ID, links previous succeeded deployment, and enqueues with status=queued. Always uses the stored application manifest (never request body).
- `arsy/httpapi/deployments.go:getDeploymentProgress` — GET handler projects deployment state to deterministic percent/step mapping (queued=0, preparing=10, uploading=30, validating=50, deploying=70, verifying=90, succeeded=100, failed/cancelled=0).
- `registerDeploymentRoutes` — both routes registered behind `requireSuperuser` guard.
- `arsy/httpapi/deployments_api_test.go` — comprehensive tests: successful queueing with previous_deployment_id linkage, unknown app (404), invalid manifest (400), superuser requirement (401), progress mapping matrix across all 9 states and unknown deployment (404).

## Verification Evidence

- `go test ./arsy/httpapi/...` → ok (exit 0), including all new focused tests.
- `go build ./...` → exit 0.
- All handler conventions match existing patterns (decodeServerRequest, requireSuperuser, error/success JSON shape).

## Deviations & Notes

- Frontend (Deploy button, progress polling, styles) deferred to a follow-up task. The backend API is complete and testable.
- Overlapping deployment prevention remains issue #149 (not in scope here).

## Remaining Risks / Follow-ups

- Overlapping deployments for one application are still possible until issue #149 lands.
