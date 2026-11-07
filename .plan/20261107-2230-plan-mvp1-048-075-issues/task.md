# MVP1 Remaining Issues Task List

## Overview
This task list contains the 16 remaining MVP1 issues to be implemented.
All issues are sized: S (Small), M (Medium), L (Large), XL (Extra Large).
Size tags indicate estimated complexity, not absolute effort.

### Already Completed (5 of 21)
- #50 [MVP1-048] Secure SFTP manifest upload - Atomic temp-file + rename pattern, 1MB validation limit, test coverage
- #51 [MVP1-049] Validate Compose on target host - YAML parsing and MVP-1 policy enforcement
- #53 [MVP1-051] Implement persistent deployment job claiming - ClaimNextQueuedDeployment with PocketBase transactions
- #54 [MVP1-052] Implement background deployment worker lifecycle - Worker struct with ProcessNext/Run methods
- #55 [MVP1-053] Execute Docker Compose deployments - SSH-based execution in compose.go

## Remaining 16 Issues

### Issue #57: Inspect deployment and container status
- **File**: rsy/deployments/processor.go
- **Task**: Inspect deployment and container status
- **Size**: M (Medium)
- **Description**: Add functionality to inspect deployment status and container state. This involves integrating with SSH to check running containers and their status.
- **Dependencies**: Requires SSH connection infrastructure already in rsy/servers/
- **Acceptance Criteria**:
  - Function can retrieve deployment status from PocketBase
  - Function can query container status via SSH
  - Returns structured status information

### Issue #64: Build Deploy action and progress view
- **File**: rsy/httpapi/routes.go
- **Task**: Build Deploy action and progress view
- **Size**: M (Medium)
- **Description**: Add HTTP API endpoints and frontend components for triggering deployments and viewing progress. This includes:
  - POST /deploy endpoint to initiate deployment
  - GET /deploy/progress endpoint to check deployment status
  - UI components to display deployment progress in real-time
- **Dependencies**: Requires worker integration and state tracking
- **Acceptance Criteria**:
  - API endpoints accept deployment requests
  - Progress can be monitored via API
  - Frontend displays deployment status

### Issue #68: Add local SSH test-server suite
- **File**: rsy/servers/
- **Task**: Add local SSH test-server suite
- **Size**: M (Medium)
- **Description**: Create a test SSH server infrastructure for local development and testing. This enables testing of SSH-based deployment without requiring external servers.
- **Dependencies**: None (new infrastructure)
- **Acceptance Criteria**:
  - Test SSH server starts and listens on available port
  - Can authenticate with test credentials
  - Supports basic SSH command execution

### Issue #69: Add deterministic Compose and worker tests
- **File**: rsy/deployments/*_test.go
- **Task**: Add deterministic Compose and worker tests
- **Size**: M (Medium)
- **Description**: Write comprehensive test files that cover:
  - Compose validation edge cases
  - Worker ProcessNext and Run method behavior
  - State transition correctness
  - SFTP upload atomicity
- **Dependencies**: Existing code in compose.go, worker.go, logs.go
- **Acceptance Criteria**:
  - All test files pass (go test ./arsy/deployments/...)
  - Coverage includes error cases and boundary conditions
  - Deterministic (no time-dependent tests)

### Issue #70: Add opt-in Docker integration profile
- **File**: rsy/config/
- **Task**: Add opt-in Docker integration profile
- **Size**: S (Small)
- **Description**: Add configuration profile that enables optional Docker integration. This should be opt-in and not affect existing functionality.
- **Dependencies**: Existing config infrastructure
- **Acceptance Criteria**:
  - New config option added without breaking changes
  - Can be enabled/disabled via configuration
  - Documented in config README

### Issue #142: Encrypt Compose manifests at rest
- **File**: rsy/deployments/sftp.go, ars/security/encryption.go
- **Task**: Encrypt Compose manifests at rest
- **Size**: L (Large)
- **Description**: Implement encryption for Compose manifests before storage. This involves:
  - Integrating AES-GCM encryption from rs/security/encryption.go
  - Modifying SFTP storage to encrypt before write, decrypt on read
  - Key management for encryption/decryption
- **Dependencies**: rs/security/encryption.go already has AES-GCM utilities
- **Acceptance Criteria**:
  - Manifests encrypted before SFTP upload
  - Decryption works correctly on download
  - No plaintext secrets persisted

### Issue #143: Add safe SSH credential rotation
- **File**: rsy/servers/, ars/security/
- **Task**: Add safe SSH credential rotation
- **Size**: M (Medium)
- **Description**: Implement SSH credential rotation mechanism that:
  - Rotates keys without service disruption
  - Maintains session continuity during rotation
  - Securely handles old and new credentials during transition
- **Dependencies**: rsy/servers/ SSH connection management
- **Acceptance Criteria**:
  - Credential rotation completes without losing active connections
  - Old credentials can be revoked securely
  - New credentials accepted after rotation

### Issue #149: Prevent overlapping deployments
- **File**: rsy/deployments/worker.go, arsy/deployments/state.go
- **Task**: Prevent overlapping deployments
- **Size**: M (Medium)
- **Description**: Add logic to the worker and state machine to prevent concurrent/deployments from overlapping. This includes:
  - State-based prevention of new deployments while one is active
  - Queue management for pending deployments
  - Proper handling of cancellation and retries
- **Dependencies**: Existing state machine in state.go, Worker in worker.go
- **Acceptance Criteria**:
  - Worker rejects new deployment if one is already running
  - Pending deployments queue correctly
  - No two deployments in non-quued state simultaneously

## Implementation Priority
1. **High**: #57, #64 (API and core functionality)
2. **Medium**: #68, #69, #143, #149 (infrastructure and safety)
3. **Low**: #70, #142 (feature enhancements)

## Notes
- All tasks follow Conventional Commits format
- Tests must be added for P0/P1 fixes
- Preserve existing tests and user modifications
- Use repo-relative paths in all documentation
