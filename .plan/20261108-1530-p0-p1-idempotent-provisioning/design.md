# Design (P0-P1 idempotent provisioning)

## Architecture & Decisions

### Provisioner State Management

The provisioner workflow now has three states for each tenant:
1. **provisioning** - initial state after tenant creation, schema/migration in progress
2. **active** - tenant fully provisioned, safe for all operations
3. **failed** - provisioning failed, all operations rejected until manually reset

### State Transition Diagram

```
+------------------+     success     +------------------+
|      provisioning| ---------------> |       active     |
+------------------+                +------------------+
         ^                              |
         |      failure                 |
         | -----------------------------|
         v                              v
+------------------+     manual     +------------------+
|      failed      | <------------- |  (admin reset)   |
+------------------+                +------------------+
```

### Implementation Details

#### arsy/tenancy/provisioner.go
- Enhance existing Provisioner with status tracking
- On `Provision()` call:
  1. Check tenant status: if "active", return success (idempotent)
  2. If "failed", return error (or allow admin reset)
  3. If "provisioning" or blank, proceed with provisioning
  4. Open/create tenant database via provider
  5. Run pending migrations via runner
  6. On success: set status to "active", return nil
  7. On failure: set status to "failed", return error

#### Tenant Status Validation Middleware
- New `requireActiveTenant` middleware for engine routes
- New `requireActiveTenant` middleware for tenant admin routes
- Both check `tenancy.Status(record.GetString("status"))` 
- Return 409 Conflict if not "active"

#### API Endpoint Hardening
- All engine routes (`/engine/{model}/records`, `/engine/{model}/query`) 
  check tenant status before operation
- All tenant admin routes (`/tenants`, `/tenants/{slug}`, `/tenants/{slug}/memberships`)
  check tenant status before operation
- Membership add/list/update/delete check tenant status

#### Migration Idempotency
- `baas_tenant_memberships` migration already uses `ensureCollection` pattern
- Verify no other migrations need idempotency enhancement
- Use same pattern for any new provisioning-related collections

### Test Strategy

1. **Idempotency Test**: Provision same tenant twice, verify no duplication
2. **Retry Test**: Provision, fail, retry with same key, verify safe retry
3. **Status Transition Test**: Verify provisioning → active → failed → rejected flow
4. **Cross-Test**: Verify failed/provisioning tenants rejected in all API endpoints
5. **Leakage Test**: Verify error messages never expose internal paths/DSNs

### Files to Modify

1. `arsy/tenancy/provisioner.go` - add status tracking and idempotent provisioning
2. `arsy/migrations/baas_tenant_memberships.go` - verify idempotency (ensureCollection)
3. `arsy/httpapi/tenants_admin.go` - add status checks in all tenant operations
4. `arsy/httpapi/engine.go` or middleware - add requireActiveTenant checks
5. `arsy/httpapi/engine_test.go` - add idempotent provisioning test cases
6. `arsy/tenancy/membership.go` - update RoleAllowsWrite/etc for status checks
