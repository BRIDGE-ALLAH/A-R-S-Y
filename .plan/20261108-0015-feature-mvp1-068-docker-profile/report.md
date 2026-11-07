# Report — MVP1-068 Docker Profile

Status: COMPLETE

## Completed

- `arsy/config/config.go`: Added `DockerTests` bool flag; set from `ARSY_ENABLE_DOCKER_TESTS` env var.
- `tests/docker_profile.go`: `IsDockerEnabled`, `SpinUpTestContainer`, `Cleanup` helpers.
- `tests/docker_test.go`: 2 tests (opt-in check, cleanup safety).

## Verification Evidence

- `go build ./...` → exit 0.
- `go test ./tests/... -run Docker` → 2 tests pass (gracefully skip when not enabled).
