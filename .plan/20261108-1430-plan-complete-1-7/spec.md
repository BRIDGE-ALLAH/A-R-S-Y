## Technical Specification: Items 1-7 Deep Insight Plan

### REQ-001: CI Green (Step 1)
**Requirement**: All tests in `arsy/deployments/` must pass without weakening overlap-prevention guarantees from #149.
**Acceptance**: `go test ./arsy/deployments/...` exits 0.

### REQ-002: Manifest Encryption (Step 2)
**Requirement**: Manifest must be encrypted at save using AES-256-GCM. Ciphertext and plaintext must not appear in any API response, DTO, list, detail, log, or diagnostic.
**Acceptance**: Test no-plaintext-leakage passes.

### REQ-003: Acceptance #143 (Step 3)
**Requirement**: Audit trail for rotate/revoke (who/when), host-key revoke & mismatch semantics (no silent update), comprehensive session invalidation.
**Acceptance**: All related tests pass.

### REQ-004: Frontend #64 (Step 4)
**Requirement**: Deploy button, bounded 3s polling, progress bar/styles, sanitized output, dashboard build/test.
**Acceptance**: Dashboard builds and tests pass.

### REQ-005: SSH Test Suite #68/#69 (Step 5)
**Requirement**: Deterministic SSH test-server suite with compose/worker tests fixture/fake executor, success/failure/timeout/duplicate/recovery, network-independent.
**Acceptance**: All SSH tests pass without network dependency.

### REQ-006: Issue Reconciliation (Step 6)
**Requirement**: Verify issues #50,51,53,54,55,56,64,68,69,70,142,143,149 against main+CI, comment evidence, close/dedup per criteria.
**Acceptance**: Evidence comments added, issues closed appropriately.

### REQ-007: BaaS/AENG (Step 7)
**Requirement**: Audit foundations (arsy/engine, tenancy), prioritize #131/#103 then #138/#139/#140/#91 per dependency.
**Acceptance**: Foundations audited, priority issues addressed.

### API Contracts
- POST /deployments - Create deployment
- GET /deployments/{id}/progress - Poll progress

### Data Contracts
- Manifest: encrypted at rest, redacted in transit
- Audit log: arsy_audit_log collection
- Session store: arsy_ssh_sessions collection

### Error Scenarios
- Overlap prevention must not break #149 guarantees
- Host-key mismatch must not silently update
- Session invalidation must be comprehensive
