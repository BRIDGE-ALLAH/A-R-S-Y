# Design — Step 1 CI/deployments regression fix

## Context
Overlap-prevention (#149) introduced `HasActiveDeployment` that treats `queued` as active. `ClaimNextQueuedDeployment` uses it to skip candidates, so a lone queued job is considered "active for its app" and never claimed. The 5 regressions all depend on claiming/processing a single queued deployment.

## Decision
- Introduce `HasInProgressDeployment(app, applicationID)` that returns true only for non-terminal, non-queued states: `preparing, uploading, validating, deploying, verifying`.
- Keep `HasActiveDeployment` (includes `queued`) for API enqueue gating and any "is anything pending/active?" checks.
- `ClaimNextQueuedDeployment` uses `HasInProgressDeployment` when selecting a candidate, so it can claim a queued job when no execution is in progress for that app, but will not claim a second queued job for an app already executing.
- This preserves #149 (no concurrent executing deployments; API still blocks duplicate enqueue) and restores original worker/claim behavior for the single-queued case.

## Alternatives considered
- Make `HasActiveDeployment` exclude `queued` everywhere: rejected because API create should still prevent piling duplicate queued deployments/idempotency races (weakens #149 enqueue semantics).
- Revert #149: rejected.
- Change only tests: rejected (tests reflect valid pre-existing behavior).

## Risks
- If `HasInProgressDeployment` is misused in API paths, duplicate enqueues could occur; mitigate by keeping API on `HasActiveDeployment`.
