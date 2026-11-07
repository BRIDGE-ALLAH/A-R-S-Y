# Task — MVP1-066 Add Local SSH Test-Server Suite (Issue #68)

Status: IN_PROGRESS (skeleton only)
Scope: feature
Issue: #68
Dependencies: MVP1-035/036/037/038 (ready)

## Checklist

- [ ] [M] Create `arsy/servers/testserver.go` with `TestSSHServer` helper: embedded SSH daemon (gliderlabs/ssh or similar), deterministic test keys, fixture port allocation, mock command execution. Acceptance: server starts/stops cleanly in tests.
- [ ] [M] Add `TestSSHServer` fixture to `arsy/servers/testserver_test.go`: basic SSH connect, auth success/fail, command output capture. Acceptance: 3-4 focused tests pass, no external network.
- [ ] [S] Document test server usage in `arsy/servers/doc.go` for integration tests.
- [ ] [S] Run `go build ./...`, `go test ./arsy/servers/...`, record results in `report.md`.
- [ ] [S] Finalize and commit as `feat(servers): add local SSH test-server suite (issue #68)`, push.

**Note**: Full deterministic suite (host-key pinning, timeout, output limits edge cases) deferred to follow-up PR with expanded test coverage.
