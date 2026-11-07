# Design: MVP1 Remaining Issues

## Architecture & Decisions

### Module Boundaries

**arsy/deployments/**
- Core deployment logic isolated from HTTP API and SSH details
- State machine in state.go handles all transitions
- Worker in worker.go manages polling and processing
- Compose validation in compose.go enforces MVP-1 policy
- SFTP handling in sftp.go with atomic upload pattern
- Logging in logs.go with bounded message sizes
- Processor in processor.go bridges deployment records and execution

**arsy/servers/**
- SSH connection management and configuration
- Test SSH server suite for local development
- Credential rotation infrastructure

**arsy/httpapi/**
- REST API endpoints for deployment control
- Progress view WebSocket/SSE endpoints
- Request/response validation

**arsy/config/**
- Configuration profiles including optional Docker integration
- Feature flags for experimental features

**ars/security/**
- AES-GCM encryption utilities
- Key management for manifest encryption

### Chosen Alternatives

1. **State-based prevention** (Issue #149): Using the existing state machine to prevent overlapping deployments rather than adding a separate mutex. This is more consistent with the existing architecture and leverages the transition rules already defined in state.go.

2. **AES-GCM encryption** (Issue #142): Using the existing rs/security/encryption.go utilities rather than implementing a custom cipher. This ensures consistency and leverages already-audited cryptographic code.

3. **Opt-in Docker profile** (Issue #70): Making Docker integration opt-in rather than default. This minimizes risk and allows gradual adoption.

4. **Deterministic testing** (Issue #69): Writing tests that don't depend on timing. This ensures CI reliability and fast feedback.

### Trade-offs

- **Issue #142 (Encryption)**: Adds complexity to SFTP flow but provides security at rest. Decryption on read adds a small performance overhead.
- **Issue #143 (Credential rotation)**: Requires careful handling of active sessions but prevents credential staleness.
- **Issue #68 (SSH test server)**: New infrastructure adds maintenance burden but enables local testing without external dependencies.

### Open Questions

1. For Issue #142: What key derivation approach should be used for per-deployment encryption?
2. For Issue #143: Should rotation be manual or automatic on key expiry?
3. For Issue #64: Should progress use WebSocket, SSE, or polling?
