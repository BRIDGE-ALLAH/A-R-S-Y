# Rules (P0-P1 completion)
1. Preserve existing user/unrelated WIP; never reset/overwrite uncommitted work outside task scope.
2. No fake/demo production behavior, no invented test results/tags/releases/deploys.
3. Tenant authorization is server-side only; never trust client tenant IDs for access grants.
4. API responses/DTOs/errors must not expose filesystem paths, DSNs, credentials, raw secrets, or secret references.
5. Suspended/deleting/unavailable tenants and deactivated memberships must fail closed.
6. Add/update regression tests for every P0/P1 security/correctness fix.
7. Keep PocketBase/upstream behavior intact; do not modify `core/` unless a task explicitly requires and evidence proves necessity.
8. No destructive production operations. Push/tag/GitHub release/production deploy require explicit user authorization and actual evidence.
9. Use repo-relative paths in artifacts; keep diffs focused.
10. Run relevant verification before completion claims; report PASS/FAIL/NOT RUN/BLOCKED honestly.
