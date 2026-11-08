# Specification: Repository Improvement and Production Readiness

## Functional Requirements (FR)

### FR-001: Complete ORM and Persistence Layer
- Full CRUD operations must work with SQLite-per-tenant architecture
- All models must have Create, Read, Update, Delete operations
- Database migrations must be idempotent and safe to rerun
- Tenant isolation must be enforced at database level

### FR-002: Manifest Encryption and Redaction
- All manifests must be encrypted at rest using AES-256-GCM
- manifest_ciphertext field must store encrypted data
- Plaintext must never be exposed in API responses (redaction required)
- Decryption must occur only for authorized requests
- No plaintext leakage in any API endpoint

### FR-003: Audit Trail and Session Management
- All server operations must log who/when (who field, when timestamp)
- Host key revoke events must be logged with mismatch semantics
- Session invalidation must be comprehensive (who/when/invalidated_at)
- Audit logs must be queryable by tenant, host, event type
- Session lifecycle must be fully tracked

### FR-004: SSH Test Server Suite
- Test SSH server must support auth/exec operations reliably
- Host key fingerprint must be verifiable and mismatches detected
- Network-independent tests must pass reliably
- Start/stop/readiness commands must work consistently
- Tests must be deterministic and not depend on network conditions

### FR-005: Deployment Management with Progress Tracking
- Deployment progress must be tracked with bounded polling (max 3s)
- Deploy button must have sanitized, secure output
- Progress bar styles must be applied consistently
- Deploy operations must have proper rollback capability
- Job overlap prevention must work correctly

### FR-006: Issue Reconciliation with GitHub API
- GitHub issues must be reconciled against main+CI state
- Evidence comments must be added to tracked issues
- Duplicate/target issues must be closed/deduped according to criteria
- GitHub auth via token must work for operations

### FR-007: Tenant Isolation Enforcement
- Each tenant must have isolated data (SQLite-per-tenant architecture)
- Tenant context must be established early in request pipeline
- Cross-tenant data access must be prevented at all levels
- API endpoints must validate tenant access before operations

### FR-008: Complete API Contracts
- All API endpoints must return correct data with proper encryption/redaction
- Request validation and authentication must work for all endpoints
- Response serialization must include redaction of sensitive fields
- Error handling contracts must be consistent across all packages

## Non-Functional Requirements (NFR)

### NFR-001: Performance
- Critical API endpoints must respond within 200ms
- Database queries must use appropriate indexes
- E2E test suites must complete within reasonable time
- Bounded polling must not exceed 3-second timeout

### NFR-002: Security
- No plaintext secrets in source code
- Authentication tokens must be properly handled and validated
- API endpoints must validate input and authenticate users
- Data must be isolated per tenant (no cross-tenant leakage)
- Encryption must use AES-256-GCM with proper key management

### NFR-003: Maintainability
- Code must follow existing Go conventions
- No unnecessary abstractions or duplicate logic
- All new modules must have proper documentation
- Tests must cover new functionality with adequate coverage
- Code must be minimal and focused on actual business needs

### NFR-004: Reliability
- Application must handle graceful degradation when DB connections fail
- Must recover from transient errors without data corruption
- Database constraints must prevent invalid data
- Must not fake success responses or simulate operations

## API Contracts

### API-001: Manifest Endpoints (with encryption)
- POST /api/manifests - Create manifest (encrypt at save)
- GET /api/manifests - List manifests (ciphertext redacted in response)
- GET /api/manifests/{id} - Get manifest detail (ciphertext redacted)
- DELETE /api/manifests/{id} - Delete manifest
- PUT /api/manifests/{id} - Update manifest (encrypt)

### API-002: Audit Log Endpoints
- GET /api/audit-logs - Get audit log entries (queryable by tenant)
- GET /api/audit-logs/count - Get audit log count

### API-003: SSH Server Endpoints
- POST /api/ssh-servers - Create SSH test server
- GET /api/ssh-servers - List SSH servers
- POST /api/ssh-servers/{id}/start - Start SSH server
- POST /api/ssh-servers/{id}/stop - Stop SSH server
- GET /api/ssh-servers/{id}/commands - Execute command (auth/exec)

### API-004: Deployment Endpoints
- POST /api/deployments - Create deployment
- GET /api/deployments - List deployments
- GET /api/deployments/{id} - Get deployment detail
- POST /api/deployments/{id}/progress - Check deployment progress (bounded polling 3s)

### API-005: Tenant Endpoints
- POST /api/tenants - Create tenant
- GET /api/tenants - List tenants
- GET /api/tenants/{id} - Get tenant detail

## Data Contracts

### Model: Manifest
- ID (UUID)
- TenantID (string) - tenant isolation
- Ciphertext (bytea) - encrypted manifest data using AES-256-GCM
- Plaintext (virtual, never stored) - must be redacted from responses
- CreatedAt, UpdatedAt (timestamp)
- Version (int) - for optimistic locking

### Model: AuditLog
- ID (UUID)
- TenantID (string) - tenant isolation
- EventType (string) - e.g., deploy, revoke, session_invalidated
- Actor (string) - who performed the action
- Timestamp (timestamp) - when
- Details (JSON) - event-specific data (must not contain plaintext secrets)

### Model: SSHServer
- ID (UUID)
- TenantID (string) - tenant isolation
- HostKeyFingerprint (string) - verifiable fingerprint
- Status (string) - e.g., running, stopped, error
- CreatedAt, UpdatedAt (timestamp)

### Model: Deployment
- ID (UUID)
- TenantID (string) - tenant isolation
- ApplicationName (string)
- Status (string) - e.g., pending, in_progress, completed, failed
- Progress (int) - 0-100 percentage
- CreatedAt, UpdatedAt (timestamp)

### Model: Tenant
- ID (UUID)
- Name (string)
- CreatedAt, UpdatedAt (timestamp)

## Error Scenarios

### ER-001: Unauthorized Access
- 401 when no authentication token provided
- 403 when authenticated user lacks tenant access
- Error response must not leak implementation details

### ER-002: Validation Errors
- 422 when required fields missing or invalid
- Error details must include field-specific messages
- Must validate input before business logic execution

### ER-003: Not Found
- 404 when requested resource does not exist
- Must not leak implementation details
- Must not expose database schema details

### ER-004: Conflict
- 409 when operation conflicts with current state
- e.g., host key mismatch, duplicate deployment, overlapping jobs
- Error message must be descriptive and actionable

### ER-005: Encryption Error
- Failed to encrypt manifest at save
- Failed to decrypt manifest for authorized request
- Must provide meaningful error without exposing ciphertext

## Acceptance Criteria

- [ ] All API endpoints return correct data with proper encryption/redaction
- [ ] Tenant isolation is enforced across all operations (no cross-tenant leakage)
- [ ] Audit trail is comprehensive and queryable (who/when for all operations)
- [ ] SSH test server functions correctly in tests (auth/exec operations)
- [ ] Deployment progress polling works with 3s timeout and bounded retries
- [ ] Issue reconciliation completes with evidence comments added
- [ ] No plaintext leakage in any API response (manifest ciphertext never exposed)
- [ ] All new code has corresponding tests with coverage
- [ ] E2E tests pass for critical user workflows (auth, CRUD, deployment)
- [ ] Version bumped and production release prepared
- [ ] Database migrations are idempotent and safe to rerun
- [ ] Performance: critical endpoints within 200ms
- [ ] Security: no plaintext secrets, proper authz, tenant isolation
