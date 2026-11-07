# Design — MVP1-055 Deployment Inspection

## Context

`arsy/deployments` already owns the SSH deployment pipeline. `servers.Executor` exposes a closed set of read-only operations including `InspectComposeManifest` (`docker compose -f <path> ps --format json`). The MVP-1 backlog (issue #57) asks for a reusable inspection that answers "what is running right now" for a stored deployment.

## Decisions

- D1: New `arsy/deployments/inspect.go` holds the parser (`ParseComposePS`), the manifest-vs-actual comparison (`MissingServices`), and the orchestrator (`InspectDeploymentStatus`). The processor keeps deploying; inspection is read-only.
- D2: State normalization lives in one place; `AllComposeContainersRunning` refactors onto the parser so verification and inspection agree on semantics (unhealthy running containers are not healthy).
- D3: Inspection follows the same SSH sequence as the processor (load server record, decrypt key via `secretKeyEnv`, pin host key, 10s connect timeout, bounded executor) but never mutates records.
- D4: `missing` services are derived locally by comparing the stored Compose manifest against `ps` output — no extra SSH round-trip.
- D5: Alternatives considered: a new dedicated SSH operation for `docker compose ls` (rejected — more remote surface for no gain) and returning raw compose JSON to callers (rejected — leaks format details and pushes mapping to every consumer).

## Trade-offs

- The orchestrator performs one SSH connection per inspection; acceptable for MVP-1 console usage.
- Unknown future Compose states pass through lowercased rather than failing, so new Docker versions do not break inspection.

## Open Questions

- None blocking; HTTP exposure is intentionally deferred to issue #64 (Deploy action and progress view).
