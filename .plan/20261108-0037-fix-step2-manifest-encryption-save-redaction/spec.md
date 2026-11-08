# Spec — Step 2 manifest encryption/redaction

## Requirements
- REQ-001: On application create (and update, if present), validate Compose manifest, encrypt with configured `cfg.SecretKeyEnv`, store `manifest_ciphertext`; do not store raw secret-bearing manifest in a way exposed by normal reads (store sanitized/empty structural marker or keep plaintext only when encryption unavailable with explicit, safe behavior — prefer: if key configured encrypt; if not configured, behave compatibly but never claim encrypted). Ciphertext never returned.
- REQ-002: `applicationDTO` / list / get / console responses contain **no** `compose_manifest` body and **no** `manifest_ciphertext`.
- REQ-003: Deployment create/rollback/progress/log endpoints never return manifest body or ciphertext; `previous_manifest`/`log`/`summary` fields are sanitized.
- REQ-004: Processor decrypts only for execution; validation/Compose errors surfaced to callers/logs must not include secret values.
- REQ-005: Tests prove no plaintext leakage for a manifest containing a sentinel secret (`SECRET_TOKEN=...`) across create/list/get/log responses and stored API-visible fields.

## Acceptance
- AC-001: Relevant `arsy/httpapi`, `arsy/deployments` tests pass; build passes.
- AC-002: Evidence recorded; CI checked.
