# Task — MVP1-074 SSH Credential Rotation (Issue #143)

Status: IN_PROGRESS (skeleton)
Scope: feature
Issue: #143

## Checklist

- [x] [M] Add `RotateServerCredential(app, serverID, newPrivateKey, secretKeyEnv)` in `arsy/servers/rotate.go`: encrypt new, replace atomically, invalidate cache.
- [x] [S] Add `POST /servers/{id}/rotate-credential` endpoint in `arsy/httpapi/servers_rotation.go` with auth guard.
- [x] [S] Add 3 tests in `arsy/servers/rotate_test.go`: nil app, unknown server, empty key.
- [x] [S] Fix idempotency unique index to use `WHERE idempotency_key != ''` for SQLite compatibility.
- [x] [S] Build passes, all httpapi + servers tests green.
- [x] [S] Commit as `feat(servers): add SSH credential rotation (issue #143)`, push.
