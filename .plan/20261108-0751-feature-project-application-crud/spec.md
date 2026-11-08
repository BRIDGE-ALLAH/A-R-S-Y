# Spec

## Routes
- `GET /api/arsy/v1/projects/{id}`
- `PATCH /api/arsy/v1/projects/{id}` (name/description; slug immutable)
- `DELETE /api/arsy/v1/projects/{id}` (409 if has applications)
- `GET /api/arsy/v1/applications/{id}` (no manifest/secret fields)
- `PATCH /api/arsy/v1/applications/{id}` (name; optional compose_manifest -> re-encrypt & replace; reject while deployment active)
- `DELETE /api/arsy/v1/applications/{id}` (409 if active deployment)

## Acceptance
- 404 for missing records.
- 409 for guardrail violations.
- PATCH/DELETE mutate actual DB records.
- Responses are allow-listed DTOs.
- Focused tests pass; `go test ./arsy/...` and `go build ./...` pass.
