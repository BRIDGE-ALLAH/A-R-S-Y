# Spec — Step 3: Audit Trail, Host-Key Revoke, Session Invalidation

## Requirements

### R1: Audit Trail (Immutable Log of Credential/Key Events)

- **R1.1**: Every credential rotation must log an audit event with (event_type='credential_rotated', server_id, changed_by, timestamp, fingerprint_of_new_key).
- **R1.2**: Every host-key revocation must log an audit event with (event_type='host_key_revoked', server_id, changed_by, timestamp).
- **R1.3**: Audit log entries are immutable; no updates or deletes allowed (only creation).
- **R1.4**: Audit log queryable via `GET /servers/{id}/audit-log?limit=N&offset=M&event_type=credential_rotated`.
- **R1.5**: Audit log entries include server ID, event type, changed-by superuser ID, timestamp, and JSON metadata.

### R2: Host-Key Revoke Semantics

- **R2.1**: Each server record has a `hostKeyRevoked` boolean field (default: false).
- **R2.2**: Revoking a host-key immediately sets `hostKeyRevoked = true` and logs revocation event.
- **R2.3**: Any attempt to create a deployment for a server with `hostKeyRevoked = true` returns 409 Conflict with message "host key revoked".
- **R2.4**: Revoking a host-key invalidates all active SSH sessions for that server.
- **R2.5**: Admin can revoke a key via `POST /servers/{id}/host-key/revoke` (no key rotation, just invalidation).

### R3: Session Invalidation (Persistent Store)

- **R3.1**: SSH sessions stored in persistent `arsy_ssh_sessions` collection, not in memory map.
- **R3.2**: Session record schema: `{session_id, server_id, auth_method, created_at, expires_at, invalidated_at}`.
  - `invalidated_at = NULL` → session is valid.
  - `invalidated_at != NULL` → session was explicitly invalidated (revocation/rotation).
  - `expires_at < now() && invalidated_at = NULL` → session is expired (stale).
- **R3.3**: Worker claiming a deployment must check session validity: reject if `invalidated_at != NULL` or `expires_at < now()`.
- **R3.4**: Credential rotation or host-key revocation calls `InvalidateSessions(serverID)`, which sets `invalidated_at = now()` on all sessions for that server.
- **R3.5**: Background cleanup deletes `arsy_ssh_sessions` records with `invalidated_at != NULL AND invalidated_at < now() - 7 days`.

## API Contracts

### Audit Trail Endpoints

**GET /servers/{id}/audit-log**
- Query params: `limit` (1-100, default 50), `offset` (default 0), `event_type` (optional: credential_rotated, host_key_revoked).
- Response: `{items: [{event_type, server_id, changed_by, created_at, metadata}], limit, offset}`.
- Auth: superuser only.

**POST /servers/{id}/host-key/revoke**
- Request body: `{}` (no parameters).
- Response: `{status: "host_key_revoked", server_id, revoked_at}`.
- Auth: superuser only.
- Side effects: sets `hostKeyRevoked=true`, logs audit event, invalidates all sessions.

### Deployment Constraint

**POST /api/arsy/v1/deployments** (existing)
- If application server has `hostKeyRevoked=true`, return 409 with `{status: "error", message: "host key revoked; cannot deploy"}`.

## Data Model Changes

### Migration: `arsy_audit_log` Collection
- `id` (text, primary key)
- `event_type` (text, indexed): credential_rotated, host_key_revoked
- `server_id` (text, relation to `arsy_servers`, indexed)
- `changed_by` (text, relation to `core_superusers`, nullable)
- `created_at` (datetime, auto-set, indexed)
- `metadata` (json): stores event-specific data (new_fingerprint, reason, etc.)

### Migration: `arsy_servers` Fields Added
- `hostKeyRevoked` (boolean, default: false, indexed)

### Migration: `arsy_ssh_sessions` Collection
- `id` (text, primary key, UUID)
- `server_id` (text, relation to `arsy_servers`, indexed)
- `auth_method` (text): key | password
- `created_at` (datetime, auto-set)
- `expires_at` (datetime, nullable, indexed)
- `invalidated_at` (datetime, nullable, indexed): when session was revoked/rotated

## Error Handling

- **Audit log unavailable**: soft-fail (log to stderr but do not block credential rotation).
- **Session store unavailable**: soft-fail for backward compatibility with map-based fallback (deprecated).
- **Revoke on unknown server**: return 404.
- **Revoke when already revoked**: idempotent (return 200, log no new event).

## Testing

- Unit tests for audit event creation, session creation/invalidation, revoke logic.
- Integration tests for credential rotation → audit log creation.
- Deployment rejection when host key revoked.
- Session invalidation persists across server restart (or simulated via DB query).
- No regressions: all existing httpapi + servers + deployments tests remain green.
