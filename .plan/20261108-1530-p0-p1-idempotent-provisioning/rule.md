# Rules (P0-P1 idempotent provisioning)
1. Provisioner must be idempotent: running twice with same tenantID must produce same result (no duplication)
2. Safe retry semantics: same idempotency key on retry should not duplicate tenant or migration state
3. Tenant status transition: provisioning → active (only on success), provisioning → failed (on error)
4. No partial active exposure: failed/provisioning tenants must be rejected in all API endpoints (GET/PATCH/POST/DELETE)
5. Error responses must not expose internal paths, DSNs, credentials, or secret references
6. Use existing arsy/tenancy/provisioner.go as base for implementation
7. All existing tests must still PASS after modifications
8. Migration baas_tenant_memberships must be idempotent (ensureCollection pattern)
9. Tenant admin operations (create/list/view/rename/suspend/resume) must fail closed for non-active tenants
10. Report any test failures as BLOCKED and document deviations
