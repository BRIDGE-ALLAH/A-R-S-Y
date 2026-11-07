# Spec — Step 1 CI/deployments regression fix

## Requirements
- REQ-001: `ClaimNextQueuedDeployment` claims the oldest queued deployment for an application **only when that application has no in-progress deployment** (preparing, uploading, validating, deploying, verifying). A queued-but-not-yet-claimed deployment does not block its own claim.
- REQ-002: `ClaimNextQueuedDeployment` still skips queued deployments whose application has an in-progress deployment (preserves #149: no overlap of executing deployments).
- REQ-003: `POST /deployments` (API) still rejects with 409 when the application has any active deployment including queued (prevents duplicate enqueue/racing requests), except idempotency key returning the original deployment/result.
- REQ-004: Atomicity guarantee preserved: concurrent claims for the same single queued job yield exactly one success and one `ErrNoQueuedDeployments` (or equivalent no-job outcome) as before.
- REQ-005: The 5 previously failing tests pass: `TestClaimNextQueuedDeploymentIsAtomic`, `TestUpdateDeploymentStateUsesTransitionRules`, `TestWorkerProcessesDeploymentThroughVerification`, `TestWorkerPersistsFailureAndCancellation`, `TestWorkerMarksProcessorCancellationAsCancelled`.

## Acceptance / Evidence
- AC-001: `go test ./arsy/deployments/...` exits 0.
- AC-002: `go test ./arsy/httpapi/...` exits 0 (no #149 regression).
- AC-003: CI run for the fix commit is checked and recorded (green required for full closure; if still failing for unrelated reasons, record exact cause).
