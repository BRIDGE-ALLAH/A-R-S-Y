# Spec — Step 5 (#68/#69)

## REQ-068: SSH Test-Server Suite
- Deterministic SSH server on 127.0.0.1 with ed25519 host key.
- Accepts public-key authentication.
- Handles exec requests with canned responses for readiness commands (uname, docker version, docker compose version).
- Supports arbitrary command success for compose operations.
- Provides HostKeyFingerprint() for pinning tests.

## REQ-069: Deterministic Compose/Worker Tests
- Worker tests: success flow, failure persistence, cancellation, timeout, atomic claim (duplicate prevention).
- All tests use fake Process callbacks (no external SSH/Docker/Compose).
- Compose validation tests cover manifest structure and policy.
- Network-independent, deterministic, no timing flakiness.
