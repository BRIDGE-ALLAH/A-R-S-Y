## Architecture & Decisions: Items 1-7 Deep Insight Plan

### Context
This plan addresses 7 sequential items from deep insight analysis:
1. CI green restoration (5 test regressions)
2. Manifest encryption security gap
3. Acceptance #143 (audit trail, host-key revoke, session invalidation)
4. Frontend #64 (Deploy button, polling, progress)
5. SSH test suite #68/#69 (deterministic, network-independent)
6. GitHub issues reconciliation (MVP1)
7. BaaS/AENG foundations (#131/#103 priority)

### Module Boundaries
- `arsy/deployments/` - CI fixes, worker tests
- `arsy/servers/` - audit, revoke, sessions
- `dashboard/src/` - frontend UI
- `.plan/` - artifacts per contract

### Chosen Alternatives
- **Manifest encryption**: AES-256-GCM at save, redact from all responses
- **Host-key revoke**: Explicit action only, no silent update
- **Session invalidation**: Comprehensive coverage via arsy_ssh_sessions collection
- **Frontend polling**: Bounded 3s interval with progress bar

### Trade-offs
- Step 4 frontend deferred initially (backend API complete)
- Step 7 BaaS deferred (foundations audit first)

### Open Questions
- Phase 2 session map → DB migration (deferred)
- BaaS #131/#103 implementation scope
