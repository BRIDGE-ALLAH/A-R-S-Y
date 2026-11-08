# Report

Status: complete (scoped CRUD vertical slice).

## Changes
- Added real CRUD handlers/routes: `GET/PATCH/DELETE /api/arsy/v1/projects/{id}` and `GET/PATCH/DELETE /api/arsy/v1/applications/{id}` (`arsy/httpapi/projects.go`, registered in `registerProjectAndAppRoutes`).
- Guardrails: project delete blocked if it still has applications (409); application update/delete blocked while a deployment is active (409) via `deployments.HasActiveDeployment`; application manifest updates are re-encrypted (`deployments.EncryptManifest`) and never returned.
- DTOs remain allow-listed (`projectDTO`/`applicationDTO`); no `manifest_ciphertext`, plaintext manifest, or secrets exposed.
- Dashboard (`dashboard/src/main.js`) wired to real endpoints for edit/delete project, update manifest (explicit prompt; existing manifest never displayed), and delete application; console dist rebuilt (`index-tIHcTvYa.js`).
- Tests: `arsy/httpapi/projects_crud_test.go` covers CRUD, 404/409 guards, and manifest redaction; passes with existing CRUD tests.

## Evidence
- `go build ./arsy/httpapi/...` PASS (exit 0)
- `go test ./arsy/httpapi/... -run "TestProjectGetUpdateDeleteAndGuards|TestApplicationGetUpdateRedactsManifest|TestApplicationDeleteBlockedByActiveDeployment" -count=1` PASS (exit 0)
- `cd dashboard; npm run build` PASS (exit 0)
- `go build ./...` PASS (exit 0)
- `go test ./arsy/... -count=1` PASS (exit 0)

## Residual
- Real multi-tenant isolation for engine dynamic CRUD still partial.
- Real managed-host E2E / release/deploy remain blocked by credentials/infra/explicit authorization.
