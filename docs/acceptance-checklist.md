# Complete Acceptance & Release Checklist for A-R-S-Y v1.0.0

- [x] **Standalone Go Architecture**: Built natively without requiring Node.js in production.
- [x] **Superuser Authentication**: PocketBase superuser auth guarding all infrastructure operations.
- [x] **Secret Protection**: AES-GCM encryption for stored private keys and credentials.
- [x] **Host Key Security**: Host-key probing, SHA256 fingerprint verification, and strict pinning (no `InsecureIgnoreHostKey`).
- [x] **Docker Compose Parser & Validation**: Strict MVP-1 subset parsing preventing privileged, host networking, and socket mounting.
- [x] **Embedded Console UI**: Lightweight SPA console embedded via `embed.FS` with Overview, Servers, Projects, Applications, and Deployments.
- [x] **Operations API**: Deployment listing, log streaming (bounded buffer), stop, redeploy, and rollback endpoints.
- [x] **Test Coverage**: All package unit tests passing (`arsy/...`).
- [x] **Diagnostics**: `arsy doctor` command checking environment, database readiness, and secrets.
- [x] **Documentation**: Installation, server requirements, and quickstart guides provided under `docs/`.