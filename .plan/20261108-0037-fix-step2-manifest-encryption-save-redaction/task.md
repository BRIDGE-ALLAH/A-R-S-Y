# Task — Step 2 (#142): Manifest encryption on save + redaction

Status: IN_PROGRESS
Scope: fix/security (goal step 2)
Goal: `Long-run 1-7...` step 2

## Checklist

- [ ] [M] Audit all `compose_manifest` / `manifest_ciphertext` exposure points (HTTP DTOs, list/detail, logs, diagnostics, console/dashboard). Record findings.
- [ ] [M] Encrypt manifest at save: `createApplication` (and any update path) validates, stores `manifest_ciphertext` via `EncryptWithEnv(cfg.SecretKeyEnv)`, and does not persist/return plaintext secrets; keep structural validation working. Wire `config.Config` into the route.
- [ ] [M] Processor/deploy uses decrypted manifest only in execution path (already present) — verify no plaintext leak to logs/errors/DTOs.
- [ ] [M] Redact `compose_manifest` and `arsy:v1:` ciphertext from API responses, deployment log/list endpoints, and error messages (never return manifest body/ciphertext in DTOs).
- [ ] [S] Add/adjust tests: no-plaintext-leakage (create/list/get/deploy logs), encryption-at-save behavior (ciphertext stored when key configured; safe fallback/behavior when not), backward compatibility.
- [ ] [S] `go build ./...`, relevant tests green; record evidence in `report.md`.
- [ ] [S] Commit, push, check CI.
