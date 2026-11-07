# Specification — MVP1-075 Prevent Overlapping Deployments

## Requirements

- REQ-001: `POST /deployments` rejects (409 Conflict) if the target application has any deployment in a non-terminal state (queued, preparing, uploading, validating, deploying, verifying).
- REQ-002: If `idempotency_key` is provided in the request and matches a recent deployment for the same application, return 200 with that deployment (idempotent). If key matches a deployment for a *different* application, treat as a new request.
- REQ-003: `ClaimNextQueuedDeployment` skips queued deployments for applications with active (non-terminal) deployments already in progress.
- REQ-004: On worker startup, `RecoverStaleDeploymentLocks` marks any deployment started > 1 hour ago still in a non-terminal state as `failed` with the message "Deployment lock recovered after process restart."
- REQ-005: Successful deployment creation (queued) returns 201 with `deployment_id`, `idempotency_key` (if provided), `status`, and `previous_deployment_id`.

## Data Contracts

- Migration: add `idempotency_key` field (Text, max 128, nullable, unique index per deployment). 
- `POST /deployments` request: `{"application_id", "idempotency_key"?: string}`.
- `POST /deployments` response 201: `{"deployment_id", "idempotency_key"?, "status", "previous_deployment_id"?}`.
- Response 409 (Conflict): `{"status": "error", "message": "deployment in progress for this application"}`.

## Error Scenarios

- ERR-001: Active deployment for application → 409.
- ERR-002: Unknown application → 404 (existing behavior).
- ERR-003: Invalid manifest → 400 (existing behavior).
- ERR-004: Unauthenticated → 401 (existing behavior).
- ERR-005: Idempotency key collision (different application) → treated as new request.

## Acceptance Criteria

- AC-001: Concurrent `POST /deployments` for same app, one succeeds (201) and one fails (409).
- AC-002: Idempotent retry with same key returns original deployment (200 or 201 depending on test design).
- AC-003: `go test ./arsy/deployments` and `./arsy/httpapi` both pass, including new focused tests.
- AC-004: Build succeeds.
