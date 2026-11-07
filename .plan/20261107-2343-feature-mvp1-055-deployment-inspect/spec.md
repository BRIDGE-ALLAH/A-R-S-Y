# Specification — MVP1-055 Deployment Inspection

## Requirements

- REQ-001: Parse `docker compose ps --format json` output in both JSON-array and newline-delimited JSON forms.
- REQ-002: Normalized per-service state must be one of: `running`, `exited`, `restarting`, `unhealthy`, `missing`, `paused`, `created`, `dead`, or the raw lowercased state when unknown. A `Health` value of `unhealthy` overrides a `running` container state.
- REQ-003: A service listed in the Compose manifest but absent from `docker compose ps` output is reported with state `missing`.
- REQ-004: `InspectDeploymentStatus` returns deployment identity (`deployment_id`), record fields (`status`, `summary`, `manifest_checksum`, `previous_deployment_id`, `started_at`, `finished_at`), application slug, per-service container statuses, and an `overall` verdict (`healthy` when every service is running, `down` when any service is exited/missing/dead or none exist, `degraded` otherwise).
- REQ-005: Container inspection reuses the stored manifest path (`ManifestPath(applicationSlug, deploymentID)`) and the existing read-only compose inspect SSH operation.

## Data Contracts

- `ContainerStatus{service, image, state, health, ports}` — ports prefer `Publishers` entries (`host:published`) and fall back to the raw `Ports` string.
- `DeploymentInspection{deployment_id, application_slug, status, summary, manifest_checksum, previous_deployment_id, started_at, finished_at, overall, services[]}`.

## Error Scenarios

- ERR-001: Malformed deployment ID → error before any lookup.
- ERR-002: Unknown `deployment_id` → "deployment not found" error.
- ERR-003: Undecryptable SSH credential → error, no SSH connection attempted.
- ERR-004: SSH command failure or non-zero exit → error with no fabricated status.
- ERR-005: Empty or invalid `docker compose ps` JSON → error from the parser.

## Acceptance Criteria

- AC-001: Parser tests cover running, exited, restarting, unhealthy, and missing states (issue body requirement).
- AC-002: `go test ./arsy/deployments/...` passes with the new tests.
- AC-003: `go build ./...` passes.
