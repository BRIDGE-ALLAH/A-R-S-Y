# Task — MVP1-073 Encrypt Compose Manifests at Rest (Issue #142)

Status: IN_PROGRESS
Scope: feature
Issue: #142

## Checklist

- [x] [S] Add `manifest_ciphertext` field (Hidden, Max 100000) to `arsy_applications` schema in migration.
- [x] [M] Add `arsy/deployments/manifest.go` with `EncryptManifest(plaintext, secretKeyEnv)` and `DecryptManifest(ciphertext, secretKeyEnv)` helpers (backward-compat passthrough).
- [x] [S] Update processor: decrypt `manifest_ciphertext` before upload when present; fallback to raw `compose_manifest`.
- [x] [M] Add tests in `arsy/deployments/manifest_test.go` covering empty rejection and passthrough behavior.
- [x] [S] Build passes, manifest tests pass. API handler encryption hook deferred to follow-up (requires hook integration).
- [x] [S] Commit as `feat(deployments): add manifest encryption helpers (issue #142)`, push.
