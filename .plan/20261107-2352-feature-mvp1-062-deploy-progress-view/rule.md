# Rules — MVP1-062 Deploy Action and Progress View

1. Only superuser-authenticated requests may trigger deployments or read progress.
2. Never accept manifests from the request body; deployments always use the stored application manifest, validated with the existing `deployments.ValidateCompose`.
3. Deployment creation must only enqueue (`status = queued`); the existing worker owns all further state transitions.
4. Frontend polling is bounded: fixed 3s interval, timer stopped when the page is left or no active deployments remain; never unbounded recursive polling.
5. All rendered API values pass through the existing `escapeHTML` helper.
6. No new dependencies (Go or npm); no changes to the worker or state machine (overlap prevention is issue #149).
7. Keep responses consistent with existing handler conventions (`{status: "error", message}` for failures).
