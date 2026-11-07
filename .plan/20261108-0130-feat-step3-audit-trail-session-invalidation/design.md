# Design — Step 3: Audit Trail, Host-Key Revoke, Session Invalidation

## Architecture Context

A-R-S-Y currently manages SSH server credentials and deployment sessions in-memory. This design adds:
1. **Immutable audit log** of all credential/key events for compliance and debugging.
2. **Host-key revocation** semantics to immediately block ops if SSH key is compromised.
3. **Persistent session store** to survive process restarts and support concurrent workers.

## Module Organization

### `arsy/servers/audit.go` (new)
- `LogAuditEvent(app core.App, eventType string, serverID string, changedByID string, metadata map[string]any) error`
  - Finds `arsy_audit_log` collection, creates record, returns error (soft-fail).

### `arsy/servers/revoke.go` (new)
- `RevokeServerHostKey(app core.App, serverID string, changedByID string) error`
  - Marks server `hostKeyRevoked = true`, logs revocation event, calls `InvalidateSessions`.

### `arsy/servers/ssh.go` (refactor)
- Replace `sessionsMap sync.Map` with DB-backed sessions.
- Add `CreateSession(serverID, authMethod string, expiresAt time.Time) (sessionID string, error)`.
- Add `GetSession(sessionID string) (session *Session, error)` — checks `invalidated_at`, expiry.
- Add `InvalidateSessions(serverID string) error` — sets `invalidated_at = now()` for all server sessions.
- Add `CleanupExpiredSessions() error` — periodic cleanup (7 days).

### `arsy/httpapi/servers.go` (extend)
- Add `getServerAuditLog(e *core.RequestEvent) error` handler.
- Add `revokeServerHostKey(e *core.RequestEvent) error` handler (new endpoint).
- Update `rotateServerCredential` to call `LogAuditEvent` after atomic rotate.

### `arsy/httpapi/deployments.go` (extend)
- Update `createDeployment` to check `hostKeyRevoked` before queuing; reject 409 if true.

### `arsy/migrations/mvp1.go` (extend)
- Add `arsy_audit_log` collection definition.
- Add `arsy_ssh_sessions` collection definition.
- Add `hostKeyRevoked` field to `arsy_servers`.

## Key Design Decisions

### Audit Log as Separate Immutable Collection
- **Rationale**: Audit logs must be immutable for compliance. Storing them in a separate collection prevents accidental data modification and makes access patterns clear.
- **Trade-off**: Adds a collection; slightly more query overhead. Mitigated by indexing on `event_type`, `server_id`, `created_at`.
- **Alternative considered**: JSON field in servers table — rejected because it would require entire server record update for every event.

### Host-Key Revoke as Boolean Flag
- **Rationale**: Simplicity. A single boolean check at deployment request entry is fast and clear.
- **Trade-off**: Cannot distinguish between "never revoked" vs. "revoked then re-enabled." Acceptable for MVP; future enhancement: revocation history in audit log.
- **Alternative considered**: Revocation timestamp — rejected in favor of boolean for simplicity now.

### Session Invalidation (Persistent Store, Not Map)
- **Rationale**: Multi-worker deployments need shared session state. In-memory map not sufficient for stateless scaling. DB-backed avoids race conditions.
- **Trade-off**: Schema change + migration; more query overhead per session check. Mitigated by indexing, in-memory caching (future), or session reuse.
- **Backward compatibility**: Old sessions map entries gracefully ignored; all new sessions use DB.

### Soft-Fail for Audit Log Unavailable
- **Rationale**: Audit log is observability, not correctness. Credential rotation should not fail if audit collection missing.
- **Trade-off**: Silent partial failure. Logged to stderr for operator debugging.
- **Mitigation**: Migration ensures collection exists; operator alerts on migration failures.

### Session Expiration as Independent of Invalidation
- **Rationale**: Expiration is time-based staleness; invalidation is explicit security-based revocation. Both states must be checkable.
- **Trade-off**: Two timestamps per session. Simplified by single check: invalid if `invalidated_at != NULL` OR `expires_at < now()`.

### Background Cleanup (7-day retention)
- **Rationale**: Prevents unbounded growth of audit log. 7 days is arbitrary; operator can configure. Audit logs in external system for longer retention.
- **Trade-off**: If deployment is replayed from history after 7 days, session ID may not exist (benign—new session created).
- **Implementation**: Cron job or admin endpoint; details in Step 3b if needed.

## Deployment Order

1. **Phase 1**: Add collections to migration; add audit + revoke helpers; wire into credential rotation endpoint. (Step 3a in this plan)
2. **Phase 2**: Refactor session map → DB; wire session checks into worker. (Step 3b if needed; large refactor)
3. **Phase 3**: Add host-key revoke endpoint; add audit log query endpoint. (Step 3c if needed)

## Integration Points

- **Credential rotation** (`arsy/httpapi/servers.go:rotateServerCredential`): calls `LogAuditEvent` after save.
- **Deployment creation** (`arsy/httpapi/deployments.go:createDeployment`): checks `hostKeyRevoked` before queuing.
- **Worker claiming** (`arsy/deployments/processor.go`): validates session before use (future; may defer to Step 3b).
- **Background job** (new or existing heartbeat): calls `CleanupExpiredSessions()` daily.

## Open Questions / Risks

1. **Session reuse**: Should sessions be reusable across multiple deployments (long-lived) or per-deployment (short-lived)? Current design supports both; details per implementation.
2. **Audit log query performance**: If audit log grows large (100k+ records), queries may be slow without proper indexing. Indexed on `server_id`, `event_type`, `created_at` by design.
3. **Operator visibility**: How does operator know host key is revoked? Answer: audit log query or UI (future). For now, logged in audit trail.
4. **Backward compat for old sessions**: Map-based sessions coexist with DB sessions during rollout. When to delete map? Answer: after all active sessions cleared and 1 deploy cycle.
