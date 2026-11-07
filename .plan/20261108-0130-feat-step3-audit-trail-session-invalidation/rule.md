# Rules — Step 3 Audit Trail + Session Invalidation

## Hard Rules

1. **Audit log must be immutable once written**. No updates/deletes to `arsy_audit_log` records; soft-delete via separate flag if needed (do not add now).
2. **Session invalidation must be immediate and persistent**. Map-based state is not sufficient; all session state must be in DB.
3. **Host-key revoked blocks ALL operations**: deployments, SSH operations, redeployments, rollbacks. Use single `hostKeyRevoked` boolean check at deployment request entry.
4. **Backward compatibility**: Existing servers without `hostKeyRevoked` field default to `false` (not revoked); existing sessions without DB records are treated as invalid.
5. **No production data loss**: Migration must not drop existing server/deployment records; only add new fields/collections.
6. **Credential rotation and host-key revocation must log via `LogAuditEvent`**; do not log to stdout or logs (audit trail must be queryable).
7. **Test coverage**: Every new audit event and session invalidation path must have a regression test. Do not weaken existing test assertions.
8. **Session cleanup**: Expired sessions may be deleted after 7 days; cleanup must not remove active/recently invalidated sessions.

## Forbidden Actions

- Do not modify `arsy_deployments` record structure without migration.
- Do not add external audit service call (Splunk, DataDog, etc.); keep audit trail local.
- Do not use string concatenation for audit log events; use structured JSON metadata field.
- Do not remove the map-based `sessionsMap` until DB-backed sessions are fully tested and in use.
- Do not bypass `InvalidateSessions` when rotating credentials or revoking keys.
