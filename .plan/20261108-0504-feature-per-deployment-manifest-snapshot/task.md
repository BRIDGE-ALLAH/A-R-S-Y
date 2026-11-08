# Task: Per-deployment manifest snapshot + true historical rollback

Scope: `.plan/20261108-0504-feature-per-deployment-manifest-snapshot/`

Goal: Store the exact Compose manifest deployed by each deployment so that
`rollback` restores the previously deployed manifest instead of re-applying the
current (possibly changed) application manifest. This closes the documented
rollback limitation without any migration-time data loss.

Acceptance criteria (all must pass):

- `go build ./...` exits 0.
- `go test ./arsy/... -count=1` exits 0.
- New column `manifest_ciphertext` (hidden) exists on `arsy_deployments`.
- Each successful claim/persist of a deployment stores an encrypted snapshot of
  the exact manifest bytes that were uploaded.
- `POST /deployments/{id}/rollback` restores the previous deployment's snapshot
  into the application before queuing the new deployment; if the previous
  deployment has no snapshot, the endpoint returns 409 with a clear message.
- Existing redeploy/rollback/create contract tests still pass.

## Tasks

- [x] T1 (S) Add migration `arsy/migrations/deployment_snapshot.go` adding hidden
  `manifest_ciphertext` TextField (Max 100000) to `arsy_deployments`; register
  under version `20261108050000_arsy_deployment_snapshot.go`; `down` removes the
  field only. Also add `manifest_checksum` is already present (no-op).
  Files: `arsy/migrations/deployment_snapshot.go`.
  Deps: none.
  Verify: `go test ./arsy/migrations/... -count=1`.

- [x] T2 (M) Add snapshot helpers in `arsy/deployments/snapshot.go`:
  `SnapshotManifest(app, deploymentID, plaintext, secretKeyEnv) error` (encrypt +
  save onto the deployment record), `LoadDeploymentManifest(app, deploymentID,
  secretKeyEnv) (string, error)` (returns snapshot plaintext or "" if absent),
  and `ApplyManifestSnapshot(app, applicationID, plaintext) error` (re-encrypt
  and store onto the application record with cache).
  Files: `arsy/deployments/snapshot.go`.
  Deps: T1.
  Verify: unit test in T4.

- [x] T3 (M) Wire snapshot persistence into the deployment lifecycle:
  - `arsy/deployments/processor.go`: after manifest is resolved/validated,
    call `SnapshotManifest` so the deployment stores what it uploads (idempotent
    if already present).
  - `arsy/httpapi/deployments.go` `rollbackDeployment`: load the previous
    deployment snapshot; if present, restore it into the application via
    `ApplyManifestSnapshot` before queuing; return 409 when the snapshot is
    absent. Keep response contract (`status`,`deployment_id`,`application_id`,
    `rollback_to`,`previous_deployment_id`).
  Files: `arsy/deployments/processor.go`, `arsy/httpapi/deployments.go`.
  Deps: T2.
  Verify: T4 tests + `go test ./arsy/httpapi/... -count=1`.

- [x] T4 (M) Tests:
  - `arsy/deployments/snapshot_test.go`: encrypt/round-trip, absent snapshot,
    apply-to-application.
  - Extend `arsy/httpapi/projects_and_deployments_test.go`: rollback uses the
    previous deployment snapshot (assert application manifest restored to the
    snapshot value) and 409 when previous has no snapshot.
  Files: `arsy/deployments/snapshot_test.go`,
  `arsy/httpapi/projects_and_deployments_test.go`.
  Deps: T3.
  Verify: `go test ./arsy/deployments/... ./arsy/httpapi/... -count=1`.

- [x] T5 (S) Update `report.md` with evidence (commands, exit codes, PASS/FAIL).
  Files: `report.md`.
  Deps: T4.
