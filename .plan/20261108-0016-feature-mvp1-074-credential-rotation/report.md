# Report — MVP1-074 Credential Rotation

Status: COMPLETE

## Completed

- `arsy/servers/rotate.go`: `RotateServerCredential` — loads server record, loads encryption key, encrypts new PEM key, atomically replaces `credential_ciphertext`, invalidates client cache entry.
- `arsy/httpapi/servers_rotation.go`: `rotateServerCredential` handler for `POST /servers/{id}/rotate-credential` behind `requireSuperuser`.
- `arsy/servers/rotate_test.go`: 3 focused tests covering nil app, unknown server, and empty key validation.
- `arsy/migrations/mvp1.go`: Fixed `idempotency_key` unique index to use `WHERE idempotency_key != ''` (SQLite partial unique index).

## Verification Evidence

- `go build ./...` → exit 0.
- `go test ./arsy/httpapi/... ./arsy/servers/...` → all tests pass (exit 0).
- Rotation rejects nil app, unknown server, and empty key; confirmed in tests.
