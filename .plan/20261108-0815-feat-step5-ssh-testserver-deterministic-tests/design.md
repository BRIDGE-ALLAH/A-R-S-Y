# Design — Step 5 (#68/#69)

## SSH Test Server (`arsy/servers/testserver.go`)
- Uses `golang.org/x/crypto/ssh` server-side APIs (already a dependency).
- Generates ed25519 host key per instance (deterministic within test, isolated per server).
- `PublicKeyCallback` accepts any key (auth focus is on the client/test, not server auth policy).
- `handleSession` parses exec payloads and returns canned responses via `cannedCommandResponse`.
- Canned responses match actual commands from `commandForOperation` in `remote.go`.

## Test Coverage
- `testserver_test.go`: lifecycle, fingerprint, auth+exec, readiness command responses.
- `client_test.go`: existing SSH client integration with `startTestSSHServer`.
- `worker_test.go`: 5 tests covering success/failure/cancellation/timeout/atomicity with fake executors.
- `compose_test.go`: validation coverage (pre-existing).
