# Task — MVP1-055 Inspect Deployment and Container Status (Issue #57)

Status: IN_PROGRESS
Scope: feature
Issue: #57
Dependencies: MVP1-053 (done — job claiming)

## Checklist

- [x] [M] Add `ParseComposePS` and `ContainerStatus` to `arsy/deployments/inspect.go`, accepting JSON array and newline-delimited `docker compose ps --format json` output and normalizing states (running, exited, restarting, unhealthy, missing). Acceptance: parser handles all five states plus ports/image fields; invalid JSON and empty output return errors.
- [x] [S] Add `MissingServices(manifest, services)` to `arsy/deployments/inspect.go` comparing manifest services against parsed containers, sorted. Acceptance: absent manifest services are reported with state `missing`.
- [x] [M] Add `InspectDeploymentStatus(app, deploymentID, secretKeyEnv)` to `arsy/deployments/inspect.go`: load the deployment record by `deployment_id`, load application/server, decrypt the SSH key, run the existing read-only `Executor.InspectComposeManifest`, and return `DeploymentInspection` with deployment fields, per-service status, and overall health. Acceptance: invalid ID, unknown deployment, and undecryptable credential return errors; happy path uses only existing read-only SSH operations.
- [x] [S] Refactor `AllComposeContainersRunning` in `arsy/deployments/processor.go` to reuse `ParseComposePS`. Acceptance: existing tests still pass; running-but-unhealthy containers are rejected.
- [x] [M] Add focused tests in `arsy/deployments/inspect_test.go` covering running, exited, restarting, unhealthy, missing, ports/image parsing, and `InspectDeploymentStatus` error paths using `tests.NewTestApp()` fixtures. Acceptance: `go test ./arsy/deployments/...` passes.
- [x] [S] Run `gofmt -l` on changed files, `go build ./...`, and `go test ./arsy/deployments/...`; record exact results in `report.md`. Acceptance: all green.
- [x] [S] Tick this checklist, finalize `report.md`, commit as `feat(deployments): inspect deployment and container status (#57)`, and push to `origin/main`. Acceptance: commit pushed, CI triggered.
