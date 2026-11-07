# Specification — MVP1-062 Deploy Action and Progress View

## Requirements

- REQ-001: `POST /deployments` with `{"application_id"}` creates a deployment record with `status=queued`, a generated 32-hex `deployment_id`, and `previous_deployment_id` pointing at the application's latest succeeded deployment when one exists.
- REQ-002: The stored manifest is validated before queueing; invalid manifests return HTTP 400 and never create a record.
- REQ-003: Unknown application returns 404; malformed body returns 400.
- REQ-004: `GET /deployments/{id}/progress` returns `{deployment_id, state, percent, step}` where percent is deterministic per state: queued=0, preparing=10, uploading=30, validating=50, deploying=70, verifying=90, succeeded=100, failed=0, cancelled=0; unknown deployment IDs return 404.
- REQ-005: The console Deployments page can trigger a deployment for any application and shows live progress bars for active deployments, stopping polling when all deployments are terminal.

## Data Contracts

- `POST /deployments` request: `{"application_id": string}`; response 201: `{"deployment_id", "application_id", "status", "previous_deployment_id,omitempty"}`.
- `GET /deployments/{id}/progress` response 200: `{"deployment_id", "state", "percent", "step"}`.

## Error Scenarios

- ERR-001: Missing/blank `application_id` → 400.
- ERR-002: Unknown application → 404.
- ERR-003: Stored manifest fails `ValidateCompose` → 400, no record created.
- ERR-004: Unknown deployment ID on progress → 404.
- ERR-005: Unauthenticated caller → rejected by `requireSuperuser`.

## Acceptance Criteria

- AC-001: UI can drive queued → succeeded via the new endpoints with bounded polling.
- AC-002: `go test ./arsy/httpapi/...` passes including new focused tests.
- AC-003: `npm run build` in `dashboard/` succeeds.
