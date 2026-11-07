# Design — MVP1-062 Deploy Action and Progress View

## Context

The deployments API already exposes listing, detail, logs, stop, redeploy, and rollback, but nothing can create a deployment: the redeploy handler even instructs callers to use a `POST /deployments` endpoint that does not exist yet. The dashboard lists deployments but cannot start one or show progress.

## Decisions

- D1: Enqueue-only creation. `POST /deployments` validates the stored manifest, generates the ID with `deployments.NewDeploymentID`, links the previous succeeded deployment, and saves `status=queued`. The existing claiming worker performs all real transitions — no new concurrency surface.
- D2: Progress is a pure projection of the record state via a static percent/step map; no extra polling infrastructure (SSE/websockets) for MVP-1. The UI polls the lightweight endpoint on a bounded interval and stops when no active states remain.
- D3: The Deploy action lives in both the Projects page (per-application button) and the Deployments page (application picker), both calling the same endpoint.
- D4: Alternatives considered: direct worker invocation from the handler (rejected — bypasses queue semantics), SSE subscription (rejected — MVP-1 keeps the surface small; bounded polling satisfies the issue).

## Open Questions

- None; credential rotation (#143) and overlap prevention (#149) remain separate issues.
