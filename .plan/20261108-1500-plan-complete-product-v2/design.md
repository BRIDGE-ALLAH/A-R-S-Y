 # Design: Architecture and Decisions

 ## Context Diagram

 ```
 +--------------------+       +--------------------+       +--------------------+
 |   Frontend (UI)    |<----->|   HTTP API Layer   |<----->|   Backend Services |
 +--------------------+       +--------------------+       +--------------------+
 | - Routes           |       | - Routes           |       | - Domain Models    |
 | - Components       |       | - Endpoints        |       | - Business Logic   |
 | - Forms            |       | - Validation       |       | - Repositories     |
 | - State Mgmt       |       | - Auth/Authz       |       | - Services         |
 +--------------------+       +--------------------+       +--------------------+
           ^                           ^                           ^
           |                           |                           |
           |                           |                           |
 +--------------------+       +--------------------+       +--------------------+
 |   Database (SQLite)  |       |   GitHub API       |       |   External Services|
 +--------------------+       +--------------------+       +--------------------+
 | - Tenant data      |       | - Issue operations |       | - Notification     |
 | - Manifest data    |       | - Repo operations  |       | - Monitoring       |
 | - Audit logs       |       |                    |       | - Metrics          |
 +--------------------+       +--------------------+       +--------------------+
 ```

 ## Module Boundaries

 ### arsy/tenancy
 - SQLite-per-tenant architecture
 - Each tenant gets isolated database
 - Migrations per tenant with lifecycle management
 - Membership management (who belongs to which tenant)

 ### arsy/engine
 - Core domain model and business logic
 - Error handling and mapper stubs
 - Manifest parsing, resolution, and validation
 - Registry model descriptors

 ### arsy/manifest
 - Manifest parse/resolve/validate pipeline
 - AES-256-GCM encryption at save
 - Ciphertext/plaintext redaction in responses
 - No plaintext leakage in any API endpoint

 ### arsy/deployments
 - Deployment management (compose, worker, jobs)
 - Progress tracking with bounded polling
 - SCP/SFTP file transfer
 - Container orchestration

 ### arsy/servers
 - SSH test server for integration testing
 - Host key management and fingerprint verification
 - Auth/exec command execution
 - Network-independent test fixtures

 ### arsy/httpapi
 - REST API endpoints
 - Request validation and authentication
 - Response serialization with redaction
 - Error handling contracts

 ## Chosen Alternatives with Rationale

 ### SQLite-per-Tenant vs PostgreSQL-per-Tenant
 - Chosen: SQLite-per-tenant for simpler isolation and easier testing
 - Rationale: Faster test cycles, simpler deployment, adequate for scope
 - Deferred: PostgreSQL migration for multi-tenant scaling

 ### AES-256-GCM vs Other Encryption
 - Chosen: AES-256-GCM for manifest encryption
 - Rationale: Standard, secure, Go-native support, authenticated encryption
 - Avoided: RSA (too heavy for data), CBC (no authentication)

 ### In-Memory SSH Server vs Persistent SSH Service
 - Chosen: In-memory test server for unit/integration tests
 - Rationale: Network-independent, faster, deterministic
 - Avoided: Real SSH daemon (complex, flaky in tests)

 ## Trade-offs

 1. SQLite simplicity vs PostgreSQL scalability
    - Acceptable for current scope, can migrate later

 2. Encryption performance vs security
    - AES-256-GCM overhead is negligible for manifest size

 3. In-memory SSH vs real SSH daemon
    - Tests are faster and more reliable, production uses real SSH when needed

 ## Open Questions

 1. When to migrate from SQLite to PostgreSQL for tenant scaling?
 2. What external services need integration (notifications, monitoring)?
 3. How to handle graceful degradation when DB connections fail?

 ## Module Interactions

 +--------------------+         +----------------------+
 | tenancy            | <--->  | engine (persist mapper)|
 +--------------------+         +----------------------+
           |                           |
           v                           v
 +--------------------+         +----------------------+
 |   manifest         | <--->  |   registry (model desc)|
 +--------------------+         +----------------------+
           |                           |
           v                           v
 +--------------------+         +----------------------+
 |   deployments      | <--->  |   servers (SSH test)   |
 +--------------------+         +----------------------+

 ## Design Principles

 1. KISS - Keep implementation simple, avoid unnecessary complexity
 2. DRY - Don't repeat business rules across modules
 3. Explicit over implicit - Contracts and types must be clear
 4. Fail fast - Errors should be early and obvious
 5. Testable - All components must be unit-testable
 6. Secure by default - Encryption, authz, isolation built-in

 ## Migration Path

 Phase 1: Current SQLite-per-tenant architecture
 Phase 2: Add PostgreSQL support if/when scaling needed
 Phase 3: Optional: Real SSH service integration for production
