# Report — MVP1-055 Deployment Inspection

Status: COMPLETE

## Completed

- `arsy/deployments/inspect.go` — `ParseComposePS` (JSON array + NDJSON), state normalization (`running`, `exited`, `restarting`, `unhealthy`, `missing`, plus passthrough), ports from `Publishers` or legacy `Ports`, `MissingServices`, `overallServiceState`, and `InspectDeploymentStatus(ctx, app, deploymentID, secretKeyEnv)` reusing the read-only SSH compose inspect.
- `arsy/deployments/processor.go` — `AllComposeContainersRunning` now delegates to `ParseComposePS`; removed the duplicate ad-hoc parser.
- `arsy/deployments/inspect_test.go` — parser state matrix (running, exited, restarting, unhealthy, missing, unknown passthrough, invalid/empty inputs), missing-service comparison, and `InspectDeploymentStatus` error paths (invalid ID, unknown deployment, undecryptable credential).

## Verification Evidence

- `go test ./arsy/deployments/...` → `ok github.com/bridge-allah/a-r-s-y/arsy/deployments` (exit 0), includes the pre-existing `TestAllComposeContainersRunning` suite on the refactored parser.
- `go build ./...` → exit 0.
- `go vet ./arsy/deployments/...` → exit 0.
- `gofmt -l` on changed files → no real diffs (only CRLF working-copy noise; CI checks out LF).

## Deviations

- Parser keeps rows that have a state but no service name (legacy test inputs); rows with neither are skipped. Documented in `design.md` trade-offs.
- `InspectDeploymentStatus` accepts `ctx` as its first parameter for cancellation support; not foreseen in the original issue wording.

## Remaining Risks / Follow-ups

- HTTP endpoint exposure for inspection is deferred to issue #64.
