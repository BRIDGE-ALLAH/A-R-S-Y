# Rules

1. Never store a Compose manifest in plaintext. Snapshots must be encrypted with
   the configured `ARSY_SECRET_KEY` (AES-256-GCM via `arsy/security`).
2. Never return manifest plaintext or ciphertext in any API response.
3. The snapshot must record exactly the bytes uploaded to the host; do not
   re-render or normalize the manifest.
4. Rollback must fail closed: if the previous deployment has no snapshot, return
   `409` and do not modify the application manifest.
5. Do not weaken existing redaction, host-key, revocation, or overlap checks.
6. Migration must be additive and reversible (`down` drops only the new field).
7. Keep diffs focused; no drive-by refactors.
8. Preserve existing test contracts; add regression tests for the new behavior.
9. Do not commit, push, tag, or deploy.
10. Do not add dependencies.
