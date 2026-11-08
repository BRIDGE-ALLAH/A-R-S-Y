# Completion Report: Per-deployment manifest snapshot

## Status: COMPLETE

## Task results

- [x] T1 migration — `arsy/migrations/deployment_snapshot.go` adds hidden
  `manifest_ciphertext` (Max 100000) to `arsy_deployments`; reversible `down`
  removes only that field.
- [x] T2 snapshot helpers — `arsy/deployments/snapshot.go` adds
  `SnapshotManifest`, `LoadDeploymentManifest`, `ApplyManifestSnapshot`, and
  `ErrSnapshotUnavailable`.
- [x] T3 lifecycle wiring + rollback restore — processor captures the snapshot
  before upload (`arsy/deployments/processor.go`); `rollbackDeployment` restores
  the previous deployment's snapshot into the application and returns 409 when
  the snapshot is missing (`arsy/httpapi/deployments.go`).
- [x] T4 tests — `arsy/deployments/snapshot_test.go`,
  `arsy/migrations/deployment_snapshot_test.go`, and extended
  `arsy/httpapi/projects_and_deployments_test.go`.
- [x] T5 report — this file.

## Verification log

| # | Command | Result |
|---|---------|--------|
| 1 | `go build ./...` | PASS (exit 0) |
| 2 | `go test ./arsy/deployments/... ./arsy/httpapi/... ./arsy/migrations/... -count=1` | PASS (exit 0) |
| 3 | `go test ./arsy/deployments/... -run "Snapshot\|ApplyManifest" -count=1 -v` | PASS — 5/5 new tests ran (round-trip, idempotent, missing, restore, invalid Compose) |
| 4 | `go test ./arsy/httpapi/... -run "Rollback" -count=1 -v` | PASS — 1/1 (`TestRollbackRequiresPreviousSnapshot`) |
| 5 | `go test ./arsy/... -count=1` | PASS (all arsy packages) |
| 6 | `git diff --check` | PASS (exit 0; only LF→CRLF notices) |

Assertions proven by tests (not just HTTP status codes):

- `TestSnapshotManifestRoundTrip`: stored ciphertext decrypts to the exact
  uploaded bytes and the checksum is persisted.
- `TestSnapshotManifestIsIdempotent`: a re-process does not overwrite an
  existing snapshot.
- `TestDeploymentsOperationsEndpoints`: after rollback the application record's
  decrypted manifest equals the previous deployment's snapshot, and no plaintext
  manifest remains on the application.
- `TestRollbackRequiresPreviousSnapshot`: 409 when the previous deployment has
  no snapshot (fail-closed; the application manifest is not mutated).

## Deviations

- The processor test double used by the worker unit tests does not exercise the
  snapshot capture (it injects a fake `Process`). Snapshot capture is covered by
  the dedicated `snapshot_test.go` unit tests plus the rollback integration test.
- Existing contract test step 5 previously asserted a 201 rollback from a
  deployment that never had a manifest deployed. That expectation encoded the
  old (incorrect) behavior; the test now seeds a real snapshot and additionally
  asserts manifest restoration. A separate test covers the missing-snapshot case.

## Remaining risks / follow-ups

- Snapshot capture happens in the processor. A deployment that is queued but
  never processed has no snapshot and therefore cannot be a rollback source —
  by design (there is nothing deployed to roll back to).
- Pre-migration deployments created before this change have no snapshot; rollback
  from them returns 409 rather than guessing. This is intentional (no fake data).
