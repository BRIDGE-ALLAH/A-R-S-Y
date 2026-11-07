# Hard Rules for Plan 1-7 Implementation

## Security Rules
1. Never log plaintext manifest data or encryption keys.
2. Redact ciphertext and plaintext from all API responses (DTO/list/detail/log/diagnostics).
3. Host-key revoke must be explicit action, never silent update on mismatch.
4. Session invalidation must comprehensively cover all active sessions for a user/host.
5. Audit trail must record who/when for all rotate/revoke operations.

## Quality Rules
6. Every P0/P1 fix must add a regression test.
7. Never weaken existing tests, lint, security checks, or CI gates.
8. Preserve unrelated user modifications, untracked files, and local data.
9. Keep changes focused and backward-compatible unless explicitly intended.
10. Network-independent tests must not rely on external services.

## Workflow Rules
11. Each step must have .plan/ folder artifacts per contract.
12. Commit/push after each step completion with evidence.
13. Close GitHub issues only after acceptance criteria and verification satisfied on default branch.
14. Do not delete files unless proven unused by static/dynamic imports, runtime discovery, build, CI, packaging, deployment, migrations, or external integrations.
