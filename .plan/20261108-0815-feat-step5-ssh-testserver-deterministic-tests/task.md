# Task — Step 5: SSH Test-Server Suite + Deterministic Tests (#68/#69)

Status: COMPLETE
Scope: feature
Issues: #68, #69

## Checklist

### SSH Test-Server Suite (#68)
- [x] [M] Upgrade `arsy/servers/testserver.go`: deterministic SSH server with ed25519 host key, public-key auth, exec channel handling, canned command responses (readiness + compose). Network-local (127.0.0.1).
- [x] [S] Add tests in `arsy/servers/testserver_test.go`: start/stop, host-key fingerprint, SSH auth + exec, readiness command responses.
- [x] [S] Verify integration with existing SSH client and readiness flow (uses same NewSSHClient).

### Deterministic Compose/Worker Tests (#69)
- [x] [S] Worker tests cover: success (processes through verification), failure persistence, cancellation, timeout, atomic claiming. All use fake Process executors, no network.
- [x] [S] Compose validation tests exist (`compose_test.go`). Network-independent.
- [x] [S] All tests: `go test ./arsy/servers/... ./arsy/deployments/... -count=1` — PASS.

## Acceptance Criteria
- SSH test server supports auth/key/command execution deterministically.
- Worker tests cover success/failure/timeout/duplicate scenarios (fixture/fake executor, network-independent).
- All tests green.
