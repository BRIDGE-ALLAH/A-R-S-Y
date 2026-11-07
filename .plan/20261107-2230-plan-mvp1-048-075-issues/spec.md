# Specification: MVP1 Remaining Issues

## Requirements (FR/NFR)

### Functional Requirements

**FR-057**: Deployment status inspection
- System shall provide function to retrieve deployment status from PocketBase
- System shall provide function to query container status via SSH
- Status retrieval shall return structured information including: deployment ID, step, level, message, timestamp

**FR-064**: Deploy action and progress view
- System shall expose POST /deploy endpoint to initiate deployment
- System shall expose GET /deploy/progress endpoint to check deployment status
- Frontend shall display real-time deployment progress

**FR-068**: SSH test-server suite
- System shall provide test SSH server that starts on available port
- Test SSH server shall support authentication with test credentials
- Test SSH server shall support basic SSH command execution

**FR-069**: Deterministic test coverage
- System shall have test coverage for Compose validation edge cases
- System shall have test coverage for Worker ProcessNext/Run methods
- System shall have test coverage for state transition correctness
- System shall have test coverage for SFTP upload atomicity
- All tests shall be deterministic (no time-dependent tests)

**FR-070**: Docker integration profile
- System shall provide opt-in Docker integration configuration
- Configuration shall be additive and not break existing functionality
- New config option shall be toggleable via config file

**FR-142**: Encrypt Compose manifests at rest
- System shall encrypt Compose manifests before SFTP storage
- System shall decrypt manifests on SFTP retrieval
- Encryption shall use AES-GCM from rs/security/encryption.go
- No plaintext secrets shall be persisted

**FR-143**: Safe SSH credential rotation
- System shall implement SSH credential rotation without service disruption
- Rotation shall maintain session continuity during transition
- System shall securely handle old and new credentials
- Rotation shall be revocable for old credentials

**FR-149**: Prevent overlapping deployments
- Worker shall reject new deployment if one is already running
- Pending deployments shall be queued correctly
- No two deployments shall be in non-queued state simultaneously
- State machine shall enforce transition rules to prevent overlaps

### Non-Functional Requirements

**NFR-001**: All implementations shall follow existing codebase conventions
**NFR-002**: Test coverage shall maintain or exceed current levels
**NFR-003**: No breaking changes to existing APIs or data contracts
**NFR-004**: All code shall pass go build ./... and go test ./...
**NFR-005**: Implementations shall be minimal and focused on the requirement
