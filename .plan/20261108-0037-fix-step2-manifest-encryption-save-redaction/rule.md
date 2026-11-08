# Rules — Step 2 manifest encryption/redaction

1. Never return `compose_manifest` plaintext secrets or `manifest_ciphertext` in normal API DTOs, list/detail, logs, diagnostics, or console payloads.
2. Encrypt at rest using existing `arsy/security` helpers only; decrypt only in deployment execution path immediately before upload/use.
3. Validation errors and deployment summaries must be sanitized (no secret values, no ciphertext).
4. Backward compatibility: legacy records with plaintext `compose_manifest` and no ciphertext must still deploy and must not leak when served; migrate-on-read/save only if safe and explicitly in scope (default: read-safe).
5. Preserve existing API contracts except removing secret leakage; do not break dashboard create flow (input field name can remain `compose_manifest`).
6. Conventional Commits; focused diff; all artifacts in English.
