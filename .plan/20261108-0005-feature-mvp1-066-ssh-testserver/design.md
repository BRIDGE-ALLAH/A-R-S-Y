# Design — MVP1-066 SSH Test Server

## Context

Integration tests for SSH connections need a live SSH server without requiring external infrastructure. A lightweight embedded test server enables deterministic, isolated testing.

## Decisions

- D1: Use `gliderlabs/ssh` (lightweight, embeddable, MIT licensed) for the test server.
- D2: Fixture returns listener address dynamically (avoid port conflicts in parallel tests).
- D3: Mock command handler returns fixed responses; real command execution deferred to integration/e2e tests.
- D4: Host key generated per-server for determinism; fingerprint computed for test assertions.

## Trade-offs

- Embedded server adds a test dependency; acceptable for local test suite.
- Mock commands limit what can be tested; real-world tests use actual servers.
