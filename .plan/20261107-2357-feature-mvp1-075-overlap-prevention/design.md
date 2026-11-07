# Design — MVP1-075 Prevent Overlapping Deployments

## Context

MVP-1 deployments execute over SSH to release directories per application (`/opt/arsy/apps/{slug}/releases/{deployment_id}`). Concurrent deployments for the same application can race over the lifecycle state, manifest uploads, and service state, causing corruption or undefined behavior. The worker already uses atomic transactions; we need to add application-scoped semantics.

## Decisions

- D1: **Defense in depth**: reject at both API creation time (early) and worker claim time (worker-scoped safety). API catches most cases; claiming safeguard catches edge races.
- D2: **Idempotency key** is caller-optional and tied to application context. Retries with the same key return the original deployment without creating duplicates. Different applications can use the same key.
- D3: **Stale-lock recovery** runs once at worker startup, not continuously. A 1-hour threshold is conservative for MVP-1. Stale deployments are marked `failed` immediately, freeing the application lock.
- D4: **State machine unchanged**. Transitions and constants remain identical; logic simply rejects new queued deployments for locked applications.
- D5: Alternative: dedicated `deployment_locks` table (rejected — overcomplicates; non-terminal state is sufficient).

## Trade-offs

- **No SSE/push**: clients polling the progress endpoint will notice the lock is released after terminal state → acceptable for MVP-1 bounded polling model.
- **1-hour stale threshold**: could be configurable later; fixed value is fine for MVP-1.

## Open Questions

- None; overlap prevention is self-contained and does not affect other features.
