# Design (P0-P1 adversarial tests)

## Architecture & Decisions

### Test Structure
- **Engine Isolation Tests**: `arsy/httpapi/engine_isolation_test.go`
  - Uses existing `engineFixture` from `engine_test.go`
  - Creates two active tenants, authenticates as member of one
  - Tests cross-tenant access rejection
  - Tests suspended/deleting tenant rejection
  - Tests missing/wrong tenant selector

- **Membership Status Tests**: `arsy/tenancy/membership_isolation_test.go`
  - Tests inactive membership access rejection
  - Tests suspended/deleting tenant admin access rejection

### Key Design Decisions

1. **Use Existing Fixtures**: Leverage `engineFixture` from `arsy/httpapi/engine_test.go` for test infrastructure
   - `newEngineFixture(t, authenticated bool)` creates test app with registry
   - `createTenant(t, slug, status)` creates tenant with optional DB
   - `do(t, method, path, tenant, body)` helper for HTTP calls

2. **Test Isolation Pattern**:
   - Each test creates its own tenants and cleans up
   - Use `t.Cleanup()` or `defer app.Cleanup()` for resource management
   - No shared state between test cases

3. **Fail-Closed Pattern** (per REQ-005):
   - Suspended/deleting tenants: return 409 Conflict
   - Missing/wrong tenant: return 400/401/403
   - Inactive memberships: return 403 Forbidden
   - All error responses: no internal data leakage

4. **Test Coverage Scope**:
   - Cross-tenant isolation (primary focus)
   - Membership state enforcement
   - Error/DTO leakage protection
   - Tenant selector validation

### Files Touched
- New: `arsy/httpapi/engine_isolation_test.go`
- New: `arsy/tenancy/membership_isolation_test.go`
- Modified: None (existing tests should remain passing)
