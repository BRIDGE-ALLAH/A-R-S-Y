 # Specification: Complete Product Implementation

 ## Functional Requirements (FR)

 ### FR-001: Tenant Isolation
 - Each tenant must have isolated data (SQLite-per-tenant architecture)
 - Tenant context must be established early in request pipeline
 - Cross-tenant data access must be prevented

 ### FR-002: Manifest Encryption
 - All manifests must be encrypted at rest using AES-256-GCM
 - manifest_ciphertext field must store encrypted data
 - Plaintext must never be exposed in API responses
 - Decryption must occur only for authorized requests

 ### FR-003: Audit Trail
 - All server operations must log who/when (who field, when timestamp)
 - Host key revoke events must be logged with mismatch semantics
 - Session invalidation must be comprehensive (who/when/invalidated_at)
 - Audit logs must be queryable by tenant, host, event type

 ### FR-004: SSH Test Server
 - Test SSH server must support auth/exec operations
 - Host key fingerprint must be verifiable
 - Network-independent tests must pass reliably
 - Start/stop/readiness commands must work consistently

 ### FR-005: Deployment Management
 - Deployment progress must be tracked with bounded polling (max 3s)
 - Deploy button must have sanitized, secure output
 - Progress bar styles must be applied consistently
 - Deploy operations must have proper rollback capability

 ### FR-006: Issue Reconciliation
 - GitHub issues must be reconciled against main+CI state
 - Evidence comments must be added to tracked issues
 - Duplicate/target issues must be closed/deduped according to criteria
 - GitHub auth via token must be available for operations

 ## Non-Functional Requirements (NFR)

 ### NFR-001: Performance
 - Critical API endpoints must respond within 200ms
 - Database queries must use appropriate indexes
 - E2E test suites must complete within reasonable time

 ### NFR-002: Security
 - No plaintext secrets in source code
 - Authentication tokens must be properly handled
 - API endpoints must validate input and authenticate users
 - Data must be isolated per tenant

 ### NFR-003: Maintainability
 - Code must follow existing Go conventions
 - No unnecessary abstractions or duplicate logic
 - All new modules must have proper documentation
 - Tests must cover new functionality

 ## API Contracts

 ### API-001: Manifest Endpoints
 - POST /api/manifests - Create manifest (encrypt at save)
 - GET /api/manifests - List manifests (redact ciphertext)
 - GET /api/manifests/{id} - Get manifest detail (redact ciphertext)
 - DELETE /api/manifests/{id} - Delete manifest

 ### API-002: Audit Log Endpoints
 - GET /api/audit-logs - Get audit log entries
 - GET /api/audit-logs/count - Get audit log count

 ### API-003: SSH Server Endpoints
 - POST /api/ssh-servers - Create SSH test server
 - GET /api/ssh-servers - List SSH servers
 - POST /api/ssh-servers/{id}/start - Start SSH server
 - POST /api/ssh-servers/{id}/stop - Stop SSH server
 - GET /api/ssh-servers/{id}/commands - Execute command

 ### API-004: Deployment Endpoints
 - POST /api/deployments - Create deployment
 - GET /api/deployments - List deployments
 - GET /api/deployments/{id} - Get deployment detail
 - POST /api/deployments/{id}/progress - Check deployment progress

 ## Data Contracts

 ### Model: Manifest
 - ID (UUID)
 - TenantID (string)
 - Ciphertext (bytea) - encrypted manifest data
 - CreatedAt, UpdatedAt (timestamp)
 - Version (int)

 ### Model: AuditLog
 - ID (UUID)
 - TenantID (string)
 - EventType (string) - e.g., "deploy", "revoke", "session_invalidated"
 - Actor (string) - who performed the action
 - Timestamp (timestamp) - when
 - Details (JSON) - event-specific data

 ### Model: SSHServer
 - ID (UUID)
 - TenantID (string)
 - HostKeyFingerprint (string)
 - Status (string) - e.g., "running", "stopped", "error"
 - CreatedAt, UpdatedAt (timestamp)

 ## Error Scenarios

 ### ER-001: Unauthorized Access
 - 401 when no authentication token provided
 - 403 when authenticated user lacks tenant access

 ### ER-002: Validation Errors
 - 422 when required fields missing or invalid
 - Error details must include field-specific messages

 ### ER-003: Not Found
 - 404 when requested resource does not exist
 - Must not leak implementation details

 ### ER-004: Conflict
 - 409 when operation conflicts with current state
 - e.g., host key mismatch, duplicate deployment

 ## Acceptance Criteria

 [ ] All API endpoints return correct data with proper encryption/redaction
 [ ] Tenant isolation is enforced across all operations
 [ ] Audit trail is comprehensive and queryable
 [ ] SSH test server functions correctly in tests
 [ ] Deployment progress polling works with 3s timeout
 [ ] Issue reconciliation completes with evidence comments
 [ ] No plaintext leakage in any API response
 [ ] All new code has corresponding tests
 [ ] E2E tests pass for critical user workflows
 [ ] Version bumped and production release prepared
