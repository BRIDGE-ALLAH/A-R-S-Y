# Design

Reuse existing patterns from `arsy/httpapi/projects.go` and `servers.go`:
- `decodeServerRequest` for JSON body size/shape.
- `deployments.HasActiveDeployment` to block mutating/removing apps with active work.
- `deployments.EncryptManifest` for manifest updates (store only ciphertext, clear plaintext field).
- Project delete checks `arsy_applications` by `project`; application delete checks active deployments.
- Frontend uses existing `apiFetch` and refreshes lists; destructive actions require confirmation.
