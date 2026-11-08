# Spec: Per-deployment manifest snapshot

## Problem

`arsy_deployments` stores only `manifest_checksum` (a hash). The Compose manifest
itself lives on `arsy_applications.manifest_ciphertext`. Editing an application
manifest therefore changes what `rollback` re-applies, so "rollback" cannot
restore the previously deployed revision. This is a real product gap, not a
cosmetic one.

## Requirements

### REQ-001 (snapshot storage)
Each deployment must persist the exact Compose manifest bytes that were uploaded
to the target host, encrypted at rest.

- Storage: `arsy_deployments.manifest_ciphertext` (hidden TextField, Max 100000).
- Encoding: `arsy/security` AES-256-GCM (`arsy:v1` prefix).
- The stored plaintext equals the uploaded bytes (byte-for-byte).

### REQ-002 (snapshot capture)
The snapshot is written before the manifest is uploaded, during the processing
lifecycle, so any deployment that reaches `uploading` has a snapshot. Capture is
idempotent: re-processing does not overwrite an existing snapshot.

### REQ-003 (historical rollback)
`POST /api/arsy/v1/deployments/{id}/rollback`:

- Requires source deployment in a terminal state (unchanged).
- Requires the source deployment to have `previous_deployment_id` (unchanged,
  else `409`).
- Loads the previous deployment's snapshot and restores it into the application
  (re-encrypt to `manifest_ciphertext`, update checksum) before queuing.
- If the previous deployment has no snapshot, returns `409` and leaves the
  application manifest untouched.

### REQ-004 (contract stability)
Response payload keys and status codes for create/redeploy/rollback are
unchanged. No manifest material is returned.

### NFR-001
- Migration is additive and reversible; existing rows keep working (empty
  snapshot allowed).
- No new dependencies.
- `go build ./...` and `go test ./arsy/...` stay green.

## Data contract

| Field | Type | Notes |
|-------|------|-------|
| `arsy_deployments.manifest_ciphertext` | text (hidden) | `arsy:v1` ciphertext; empty for legacy rows |
| `arsy_deployments.manifest_checksum` | text | sha256 hex of uploaded manifest (existing) |

## Error scenarios

| Scenario | Endpoint | Status |
|----------|----------|--------|
| Source not terminal | rollback | 409 |
| No previous deployment | rollback | 409 |
| Previous deployment record missing | rollback | 404 |
| Previous deployment has no snapshot | rollback | 409 |
| Snapshot decrypt failure | rollback | 500 |
| Host key revoked / overlap | rollback | 409 |

## Acceptance criteria

- AC-1: Deployment record has non-empty `manifest_ciphertext` after processing
  begins; decrypting it yields the uploaded manifest.
- AC-2: Rollback after editing the application manifest restores the previous
  snapshot into the application record (asserted by test).
- AC-3: Rollback returns 409 when the previous deployment lacks a snapshot.
- AC-4: `go test ./arsy/... -count=1` passes.
