# Design: Per-deployment manifest snapshot

## Context

```
dashboard (Deployments page)
   │  POST /deployments/{id}/rollback
   ▼
arsy/httpapi/deployments.go  ── queues new deployment ──┐
   │  restores snapshot into application                │
   ▼                                                    ▼
arsy_applications.manifest_ciphertext         arsy_deployments (queue)
                                                        │ claimed by
                                                        ▼
                                       arsy/deployments Worker → Processor
                                                        │ uploads manifest
                                                        ▼
                                       arsy_deployments.manifest_ciphertext (snapshot)
```

## Decision: reuse `arsy/security` AES-256-GCM, mirror the application field

Alternatives considered:

1. **New collection `arsy_deployment_manifests`** (1:1 with deployments).
   Rejected: more migration + relation overhead for a single hidden blob; the
   deployments collection already has `manifest_checksum`, so co-locating the
   snapshot keeps the record self-describing.
2. **Store snapshot on `arsy_applications` history** — rejected, that is the
   current broken model.
3. **Chosen: hidden `manifest_ciphertext` on `arsy_deployments`.** Minimal,
   additive, matches the existing `arsy_applications.manifest_ciphertext`
   convention, and keeps the "one deployment = one immutable revision" invariant.

## Snapshot capture point

Capture happens in the processor after the manifest is resolved and validated,
immediately before upload. Rationale:

- The API/queue path does not need the secret key beyond what it already uses.
- Legacy queued rows (pre-migration) simply capture at process time.
- Capture is idempotent: if `manifest_ciphertext` is already set, do not
  overwrite (protects against re-processing replaying a changed app manifest).

## Rollback semantics

Rollback restores the *previous deployment's snapshot* to the application and
then queues a new deployment. This keeps the existing state machine and worker
unchanged while making the queued deployment actually deploy the old revision.

Trade-off kept honest:

- The restore mutates the application manifest. This is intended: the app now
  reflects the rolled-back revision, and the new deployment record snapshots it
  again for future rollbacks (a true history chain).
- If a future requirement needs immutable app versioning, the snapshot column is
  already the source of truth to build it on.

## Open questions

- Should snapshot capture also happen at enqueue time to cover `queued` rows that
  never get processed? Deferred: queued rows cannot affect a rollback until they
  succeed, and a succeeded row always passed through the processor.
