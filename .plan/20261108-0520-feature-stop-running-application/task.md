# Task: Stop a running application (complete vertical slice)

Scope: `.plan/20261108-0520-feature-stop-running-application/`

Goal: expose the existing remote `docker compose down` capability
(`servers.Executor.StopComposeManifest`) through a real API endpoint and
lifecycle transition, so an operator can bring a running application down. Today
the capability exists in the executor but is unreachable from the product.

Acceptance criteria:

- `POST /api/arsy/v1/applications/{id}/stop` exists, superuser-only.
- On success: remote `compose down` runs against the newest succeeded
  deployment's release path; `lifecycle_state` becomes `stopped`; an audit event
  is written.
- Guards: application must be `running`; server must have a pinned host key and
  not be revoked; newest succeeded deployment must exist; no active deployment.
- No fake success: remote command failure returns 502 and lifecycle stays
  `running`.
- Frontend: project card shows a "Stop" action for running applications.
- `go build ./...` and `go test ./arsy/... -count=1` pass.

## Tasks

- [x] T1 (M) `arsy/httpapi/stop_application.go`: handler + route registration.
  Files: `arsy/httpapi/stop_application.go`, `arsy/httpapi/routes.go`.
  Deps: none.

- [x] T2 (M) `arsy/deployments/stop.go`: `StopApplication(ctx, app, applicationID,
  secretKeyEnv) (state string, err error)` implementing guard checks, SSH exec of
  `StopComposeManifest`, and lifecycle/audit update.
  Files: `arsy/deployments/stop.go`.
  Deps: T1.

- [x] T3 (S) Frontend: `dashboard/src/main.js` stop button + handler; rebuild dist.
  Files: `dashboard/src/main.js`, `arsy/console/dist/**`.
  Deps: T1.

- [x] T4 (M) Tests: `arsy/httpapi/stop_application_test.go` covering guards
  (not running → 409, revoked host key → 409, no succeeded deployment → 409,
  active deployment → 409) and the happy path with a fake SSH layer.
  Files: `arsy/httpapi/stop_application_test.go`.
  Deps: T1, T2.

- [x] T5 (S) `report.md` evidence.
