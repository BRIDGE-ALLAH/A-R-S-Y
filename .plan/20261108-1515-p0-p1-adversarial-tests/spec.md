# Spec (P0-P1 adversarial tests)

## REQ-005 #92 Adversarial isolation

### Test Coverage Requirements

#### 1. Wrong/Missing Tenant Selector
- **Case**: Request without `X-Arsy-Tenant` header or `tenant` query param
- **Endpoint**: `GET /engine/{model}/records` or other engine routes
- **Expected**: 400 or 401 status - tenant required error
- **AC**: Response must not expose secret key, DSN, or internal path data

#### 2. Cross-Tenant Access (Member A → Tenant B)
- **Case**: User authenticated as member of tenant-A attempts operations on tenant-B
- **Setup**: Create two active tenants (tenant-a, tenant-b); authenticate as member of tenant-a
- **Endpoint**: `GET /engine/{model}/records/{id}` with `X-Arsy-Tenant: tenant-b`
- **Expected**: 403 Forbidden or 404 Not Found - access denied to other tenant
- **AC**: No row from tenant-a should be visible/queryable when authenticated as tenant-b member

#### 3. Suspended Tenant Access
- **Case**: Tenant with status "suspended" attempts engine CRUD operations
- **Setup**: Create tenant with status "suspended"; create records within that tenant DB
- **Endpoint**: Any `POST/PATCH/DELETE /engine/{model}/records{id}` with suspended tenant header
- **Expected**: 409 Conflict - "tenant unavailable" or similar fail-closed response
- **AC**: Error message must not expose internal DB paths or DSNs

#### 4. Deleting Tenant Access
- **Case**: Tenant with status "deleting" attempts engine CRUD operations
- **Setup**: Create tenant with status "deleting"
- **Endpoint**: Any engine route with `X-Arsy-Tenant: <deleting-tenant-slug>`
- **Expected**: 409 Conflict - "tenant unavailable"
- **AC**: Same leakage protection as suspended tenant

#### 5. Inactive Membership Access
- **Case**: User with inactive membership status attempts tenant admin operations
- **Setup**: Create tenant, add membership with status "inactive"; authenticate as that inactive user
- **Endpoint**: `POST /tenants/{slug}/memberships` or admin operations
- **Expected**: 403 Forbidden - "tenant admin access denied" or similar
- **AC**: No membership management allowed for inactive users

#### 6. Path/Secret Leakage Checks
- **Case**: Various inputs that could cause path traversal or secret exposure
- **Endpoints**: All engine CRUD, tenant admin, membership operations
- **Inputs to test**:
  - `secret` field values in engine create/update
  - Special characters in tenant slugs
  - Path traversal patterns (`../`, etc.)
  - Very long field values
- **Expected**: 400 Bad Request or 409 Conflict for invalid inputs
- **AC**: Never return internal paths, DSNs, credential values in any error/DTO response

### Error Response Contract
All error responses across all endpoints must follow this contract:
```json
{"status": "error", "message": "human readable message only - no paths, no DSNs, no secrets"}
```
No fields like `dsn`, `path`, `credential`, `secret`, `db_url` etc. should ever appear in any response body.
