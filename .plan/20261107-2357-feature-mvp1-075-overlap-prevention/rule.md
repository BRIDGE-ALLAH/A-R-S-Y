# Rules — MVP1-075 Prevent Overlapping Deployments

1. Two deployments for the same application cannot be in non-terminal states simultaneously; reject at POST and at claim time (defense in depth).
2. Idempotency keys are caller-supplied and optional; retries with the same key return the original deployment, not a new one.
3. Active-deployment locks are released only when the deployment reaches a terminal state (succeeded/failed/cancelled).
4. Stale locks (started > 1 hour ago, still in non-terminal state) are marked failed on worker startup, not during normal polling.
5. Different applications can deploy concurrently without interference.
6. No new dependencies; no changes to the state machine constants or transitions (only add logic around them).
7. All error messages are user-friendly and API-consistent.
