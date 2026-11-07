# Report — Step 3: Audit Trail, Host-Key Revoke, Session Invalidation

Status: PHASE 1 COMPLETE (audit trail + revoke mechanics implemented)

## Completed (Phase 1)

### Migrations
- [x] Added `arsy_audit_log` collection: event_type, server (relation), changed_by, metadata (JSON), created (auto).
- [x] Added `arsy_ssh_sessions` collection: session_id, server (relation), auth_method, created (auto), expires_at, invalidated_at.
- [x] Added `host_key_revoked` boolean field to `arsy_servers` (default: false).

### Audit Trail Helpers (arsy/servers/audit.go)
- [x] `LogAuditEvent(app, eventType, serverID, changedByID, metadata)` — creates immutable audit record.
- [x] `GetServerAuditLog(app, serverID, eventType, limit, offset)` — retrieves audit entries with filtering.

### Host-Key Revoke (arsy/servers/revoke.go)
- [x] `RevokeServerHostKey(app, serverID, changedByID)` — marks `host_key_revoked=true`, logs event, invalidates sessions.

### Session Invalidation (arsy/servers/sessions.go)
- [x] `CreateSession(app, sessionID, serverID, authMethod, expiresAt)` — creates session record.
- [x] `GetSession(app, sessionID)` — validates session (not invalidated, not expired).
- [x] `InvalidateSessions(app, serverID)` — marks all server sessions as invalidated.
- [x] `CleanupExpiredSessions(app)` — deletes sessions invalidated >7 days ago.

### HTTP Endpoints (arsy/httpapi/servers.go)
- [x] `GET /servers/{id}/audit-log?limit=N&offset=M&event_type=TYPE` — list audit events.
- [x] `POST /servers/{id}/host-key/revoke` — revoke and invalidate.

### Integration
- [x] `rotateServerCredential` logs `credential_rotated` event after successful rotation.
- [x] `createDeployment` checks `host_key_revoked` before queuing; returns 409 if true.

## Verification Evidence

- **Build**: `go build ./arsy/...` → exit 0 ✅
- **Tests**: `go test ./arsy/httpapi/... ./arsy/servers/... ./arsy/deployments/... -count=1` → all pass ✅
- **Format**: `go fmt ./arsy/...` applied to all changed files ✅
- **Commit**: `fa91f00` pushed to main ✅
- **Schema**: Migrations register 2 new collections + 1 new field on existing collection ✅
- **Audit immutability**: `LogAuditEvent` only creates records (no update/delete) ✅
- **Host-key check**: Deployment rejects with 409 if `host_key_revoked=true` ✅

## Remaining (Phase 2+)

### Session Refactoring (Large, Future Phase)
- [ ] Replace in-memory `sessionsMap` in `arsy/servers/ssh.go` with DB-backed GetSession calls.
- [ ] Update worker in `arsy/deployments/processor.go` to validate sessions via DB.
- [ ] Add background cleanup job to delete old invalidated sessions.

### Audit Log Queries (Nice-to-have)
- [ ] Expose audit log in dashboard/UI for operational visibility.
- [ ] Add filtering by date range for large logs.

## Deviations
- None. Implemented per spec for Phase 1 (audit trail + host-key revoke + basic session store).

## Risks / Next Steps
- **Session refactoring risk**: Large change; requires careful testing of worker session logic (defer to Phase 2).
- **Next**: Implement frontend endpoints (#64) and SSH test suite (#68/#69), or continue Session Phase 2 if needed.
