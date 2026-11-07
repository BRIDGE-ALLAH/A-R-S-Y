# Task — Step 3: Audit Trail, Host-Key Revoke, Session Invalidation (Issue #143)

Status: READY
Scope: feature
Issue: #143

## Checklist

### Audit Trail (Credential/Host-Key Events)

- [ ] [M] Add `arsy_audit_log` collection migration in `arsy/migrations/mvp1.go`: event type (rotated/revoked), server_id, changed_by, timestamp, metadata_json.
- [ ] [M] Add `LogAuditEvent(app, eventType, serverID, changedByID, metadata)` in `arsy/servers/audit.go`: creates audit record, soft-fails if collection unavailable.
- [ ] [S] Call `LogAuditEvent` in `arsy/httpapi/servers.go:rotateServerCredential` after atomic rotate; log new key fingerprint + timestamp.
- [ ] [S] Call `LogAuditEvent` in future host-key revoke endpoint (reserve for Step 3b).
- [ ] [S] Add `GET /servers/{id}/audit-log` endpoint in `arsy/httpapi/servers.go`: list last 100 audit entries, filter by event type.
- [ ] [S] Add tests in `arsy/servers/audit_test.go`: log event, query by server, handle missing collection gracefully.

### Host-Key Revoke Semantics

- [ ] [S] Add `hostKeyRevoked` boolean field to `arsy_servers` in migration (default: false).
- [ ] [S] Add `RevokeServerHostKey(app, serverID)` in `arsy/servers/revoke.go`: marks record, logs revocation event, clears SSH sessions.
- [ ] [M] Update `createDeployment` in `arsy/httpapi/deployments.go`: check `hostKeyRevoked` before queuing deployment; reject with 409 if true.
- [ ] [S] Add `POST /servers/{id}/host-key/revoke` endpoint: guard with `requireSuperuser`, return 200 on success.
- [ ] [S] Add tests: revoke blocks future deployments, audit log records revocation event.

### Session Invalidation (SSH Sessions Map → Persistent Store)

- [ ] [M] Add `arsy_ssh_sessions` collection to `arsy/migrations/mvp1.go`: session_id (UUID), server_id, auth_method (key/password), created_at, expires_at, invalidated_at (NULL = valid).
- [ ] [L] Refactor `arsy/servers/ssh.go`: replace `sessionsMap` with DB-backed sessions; add `CreateSession`, `GetSession`, `InvalidateSessions(serverID)`, `ExpiredSessionCleanup`.
- [ ] [M] Update deployment worker in `arsy/deployments/processor.go` to use `GetSession` + check `invalidated_at` before claiming connection.
- [ ] [S] Call `InvalidateSessions(serverID)` when host-key revoked or credential rotated.
- [ ] [S] Add background cleanup job (or manual cleanup endpoint) to delete expired sessions older than 7 days.
- [ ] [S] Add tests: session creation/lookup, invalidation blocks use, cleanup.

## Dependencies
- Step 2 complete (manifest encryption); session work may depend on manifest deployment flow.
- No external dependencies; work is internal refactoring + audit/revoke features.

## Acceptance Criteria
- [x] Audit trail records all credential/key rotation and revocation events.
- [x] Host-key revoke immediately blocks new deployments.
- [x] SSH sessions stored persistently, invalidation persists across restarts.
- [x] All tests pass: `go test ./arsy/httpapi/... ./arsy/servers/... ./arsy/deployments/...`.
- [x] No test regressions from refactoring session map → DB.
