# Spec (P0-P1 idempotent provisioning)

## REQ-006 #89 Idempotent provisioning

### Provisioner Contract
The Provisioner must satisfy these requirements:

1. **Idempotent Provisioning**
   - `Provision(ctx, tenantID)` running twice with same tenantID must produce identical result
   - If tenant database already exists, only run pending migrations (not re-create)
   - No duplicate records or schema objects on repeated provisioning

2. **Status Management**
   - Initial state: tenant created with status "provisioning"
   - On successful schema + migration: status → "active"
   - On failure: status → "failed", no active exposure
   - Failed tenants must be rejected in all subsequent operations

3. **Retry Semantics**
   - Same idempotency key on retry must be safe (no duplication)
   - Different keys on same tenant may create new provisioning session
   - Retry must check current tenant status before proceeding

4. **No Partial Active Exposure**
   - Tenants with status "provisioning" or "failed" must be rejected in:
     - Engine CRUD routes (GET/PATCH/POST/DELETE)
     - Tenant admin routes (GET/PATCH/POST/DELETE /tenants/{slug})
     - Membership operations
   - Error response: 409 Conflict with "tenant unavailable" message
   - Error messages must not expose internal paths or secrets

5. **Provider Interface**
   - Reuse existing `arsy/tenancy/provisioner.Provider` interface
   - `Provision(ctx, tenantID)` returns error on failure
   - Provider responsible for opening/creating datastore

### Implementation Requirements

#### arsy/tenancy/provisioner.go Enhancements
```go
type Provisioner struct {
    provider Provider
    runner   *MigrationRunner
}

// NewProvisioner creates a Provisioner using the given provider and migrations.
func NewProvisioner(provider Provider, migrations []Migration) *Provisioner

// Provision opens (or creates) the tenant database and applies pending migrations.
// Returns ErrInvalidTenantID for a blank tenantID.
// On success: sets tenant status to "active" 
// On failure: sets tenant status to "failed", returns error
func (p *Provisioner) Provision(ctx context.Context, tenantID string) error
```

#### baas_tenant_migrations Migration
- Must use ensureCollection pattern (already in baas_memberships.go)
- Idempotent: if collection exists, just verify indexes and fields
- No destructive schema rewrites

#### API Endpoint Hardening
- All tenant operations must check tenant status before proceeding
- Suspended/deleting/provisioning/failed tenants: return 409 Conflict
- Error responses: no internal path/DSN/secret leakage

### Error Response Contract (consistent with REQ-005)
```json
{"status": "error", "message": "tenant unavailable"}
```
or
```json
{"status": "error", "message": "tenant is not active"}
```
Never expose dsn, path, credential, secret fields.
