# Plan: Project & application CRUD vertical slices

- [x] Audit missing CRUD routes for projects/applications and confirm DTOs never expose secrets
- [x] Implement GET/PATCH/DELETE project and GET/PATCH/DELETE application API handlers with guardrails
- [x] Register routes and keep superuser-only auth consistent
- [x] Add API tests for CRUD, redaction, validation, and deletion guards
- [x] Wire dashboard list/detail actions to the real endpoints and rebuild console dist
- [x] Run focused + full `go test ./arsy/...` and `go build ./...`
- [x] Update report with evidence and residual gaps
