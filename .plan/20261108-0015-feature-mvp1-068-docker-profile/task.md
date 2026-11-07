# Task — MVP1-068 Docker Integration Profile (Issue #70)

Status: IN_PROGRESS
Scope: feature
Issue: #70
Dependencies: MVP1-053/055/057 (ready)

## Checklist

- [x] [S] Add `ARSY_ENABLE_DOCKER_TESTS` config flag to `arsy/config/config.go` (opt-in, default false).
- [x] [S] Create `tests/docker_profile.go` helper: container spin-up, port allocation, cleanup (no broad prune).
- [x] [S] Add 2-3 focused tests in `tests/docker_test.go`: opt-in check, cleanup safety.
- [x] [S] Build passes, tests pass when Docker enabled or skipped gracefully.
- [x] [S] Finalize and commit as `feat(tests): add opt-in Docker integration profile (issue #70)`, push.
