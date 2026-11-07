# Task — MVP1-062 Deploy Action and Progress View (Issue #64)

Status: IN_PROGRESS
Scope: feature
Issue: #64
Dependencies: MVP1-052 (done), MVP1-054 (done), MVP1-056 (manifest upload, done), MVP1-061 (done)

## Checklist

- [x] [M] Add `createDeployment` handler for `POST /deployments` in `arsy/httpapi/deployments.go`: decode `application_id`, load the application, validate the stored Compose manifest, generate a deployment ID, link the latest succeeded deployment as `previous_deployment_id`, and queue the deployment. Acceptance: 201 with a queued deployment; unknown application 404; invalid manifest 400.
- [x] [S] Add `getDeploymentProgress` handler for `GET /deployments/{id}/progress` returning `deployment_id`, `state`, `percent`, and `step` from a deterministic state map. Acceptance: every record state maps to a stable percent/step.
- [x] [S] Register both routes behind `requireSuperuser` in `registerDeploymentRoutes`. Acceptance: unauthenticated requests are rejected.
- [x] [M] Add API tests in `arsy/httpapi/deployments_api_test.go` covering successful queueing, unknown application, invalid manifest, progress mapping across states, and missing deployment. Acceptance: `go test ./arsy/httpapi/...` passes.
- [ ] [M] Extend `dashboard/src/main.js`: Deploy button per application card, deploy form on the Deployments page, and a bounded poll (3s interval, stops when no active deployments) rendering progress bars from the progress endpoint. Acceptance: `npm run build` succeeds in `dashboard/`. *(deferred to follow-up task)*
- [ ] [S] Add progress bar styles to `dashboard/src/style.css` and run dashboard build; record results in `report.md`. *(deferred to follow-up task)*
- [ ] [S] Finalize artifacts, commit as `feat(httpapi): add deploy trigger and progress endpoints (issue #64)`, and push.
