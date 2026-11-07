# Specification — MVP1-066 SSH Test Server

## Requirements

- REQ-001: `TestSSHServer` fixture accepts a listener address (or 0 for auto-port), public key, returns server and listener address.
- REQ-002: SSH connect with valid credentials succeeds; invalid credentials fail.
- REQ-003: Mock command execution returns structured output (stdout, stderr, exit code).
- REQ-004: Host-key fingerprint deterministic and inspectable for test validation.
- REQ-005: Server closes all connections and listeners on shutdown.

## Acceptance Criteria

- AC-001: 3+ focused tests in testserver_test.go covering connect, auth, command execution.
- AC-002: `go test ./arsy/servers/...` passes with new tests.
- AC-003: `go build ./...` succeeds.
