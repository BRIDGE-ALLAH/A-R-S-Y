# Rules

1. Preserve all existing API behavior and tests; additive changes only.
2. Never expose secrets, credentials, or raw manifest ciphertext in responses.
3. Tenant DB paths must stay under configured tenant root and use validated IDs/slugs.
4. Engine CRUD remains superuser-gated and `allowed_fields`-enforced until per-tenant authz exists.
5. Do not commit/push/tag/release/deploy. Keep changes focused; record deviations in report.
6. Add meaningful tests for new behavior; `go build ./...` and `go test ./arsy/...` must pass.
