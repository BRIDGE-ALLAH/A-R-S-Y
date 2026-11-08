# Design — Step 2 manifest encryption/redaction

## Exposure audit (initial)
- `arsy/httpapi/projects.go`: `createApplication` stores request `compose_manifest` plaintext; DTO omits it (good) but record keeps plaintext.
- `arsy/httpapi/deployments.go`: `rollbackDeployment` returns `previous_manifest: prevRecord.GetString("summary")` (misleading name but summary-only; must stay sanitized). No manifest body returned.
- `arsy/deployments/processor.go`: decrypts ciphertext when present (good). Validation errors may echo details — must sanitize.
- `dashboard/src/main.js`, `arsy/console/dist/...`: send `compose_manifest` input (expected); must not display returned manifest.
- `arsy/httpapi/deployments_api_test.go` fixtures store plaintext manifests (test-only).

## Decisions
- Thread `config.Config` into `registerProjectAndAppRoutes`/`createApplication(cfg)`; validate with `deployments.ValidateCompose`, encrypt via `deployments.EncryptManifest` when key loadable; set `manifest_ciphertext`, clear/neutralize stored `compose_manifest` (e.g. empty string) to prevent plaintext at rest when encrypted. If encryption fails because key not configured, reject with clear non-secret error for create requiring secrets-at-rest? Or keep compat — decide in implementation with least surprise; never leak secrets.
- Add shared redaction helper for summaries/errors/log fields: strip `arsy:v1:*` tokens and cap/sanitize; ensure handlers never include manifest fields.
- Update processor validation path to return generic error without embedding manifest content.
