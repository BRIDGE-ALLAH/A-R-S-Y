# AENG-019: Define Module Security Metadata Contract

**Status:** Draft
**Owner:** A-R-S-Y Team
**Date:** 2026-11-09
**Depends On:** AENG-006, AENG-003

## Purpose
Define the security metadata contract that AENG modules use to declare access control policies, field-level permissions, and runtime enforcement rules for module models and actions.

---

## 1. Security Metadata Contract Structure

### 1.1 Top-Level Security Definition
```go
type ModuleSecurity struct {
    Module      string
    Version     string
    Effect      Effect        // "allow" or "deny" default
    Resources   []ResourceMetadata
    Actions     []ActionSecurity
    Roles       []RoleMetadata
}

// Effect defines the default access effect
type Effect string

const (
    EffectAllow Effect = "allow"
    EffectDeny  Effect = "deny"
)
```

### 1.2 Resource Metadata (Collections/Models)
```go
type ResourceMetadata struct {
    Name        string          // Collection/model name
    Plural      string          // e.g. "companies", "company"
    Description string          // Human-readable description
    Actions     []string        // Allowed actions on this resource
    Identifier  IdentifierMeta  // How to identify records
    FieldPermissions []FieldPermission // Per-field access control
}

// IdentifierMeta defines how records are uniquely identified
type IdentifierMeta struct {
    Field    string   // e.g. "id"
    Type     string   // e.g. "uuid", "integer"
    Format   string   // e.g. "uuid", "int64"
    Example  string   // e.g. "123e4567-e89b-12d3-a456-426614174000"
```
```

### 1.3 Field Permission
```go
type FieldPermission struct {
    Field      string   // Field name, e.g. "salary", "email"
    Effect     Effect   // "allow" or "deny"
    Roles      []string // Roles allowed; empty = superuser only
    Conditions []Condition // When this permission applies
}

// Condition for field-level access
type Condition struct {
    Field      string   // e.g. "tenant_id"
    Operator   string   // "eq", "neq", "gt", "lt", "contains"
    Value      any      // Value to compare, can be variable reference
}
```

### 1.3 Action Security
```go
type ActionSecurity struct {
    Name       string          // Action name, e.g. "delete", "archive", "export"
    Description string          // Human-readable
    Effect     Effect          // "allow" or "deny"
    Roles      []string        // Roles allowed
    Permissions []FieldPermission // Additional field-level checks
    Mutates    bool            // Does this action mutate state?
    ReadOnly   bool            // Does this action only read?
}
```

---

## 2. Role Metadata

### 2.1 Role Definition
```go
type RoleMetadata struct {
    Name       string   // Role name, e.g. "admin", "viewer", "developer"
    Description string   // Human-readable
    Inherits   []string   // Roles this inherits permissions from
    Effect     Effect   // Default effect for role
}
```

### 2.2 Built-in Roles
| Role | Typical Permissions |
|------|---------------------|
| **superuser** | All actions, all resources, bypass all checks |
| **admin** | All resources, most actions, can manage users |
| **developer** | Can view/modify own modules, cannot change config |
| **viewer** | Read-only access to resources |
| **tenant_admin** | Full access to own tenant's resources |
| **member** | Limited CRUD on own records |

---

## 3. Security Evaluation API

### 3.1 Permission Check Function
```go
func CheckPermission(
    ctx context.Context,
    subject *core.Record,     // The user/principal
    action string,           // e.g. "read", "write", "delete", "export"
    resource *ResourceMetadata, // The resource/collection
    options ...CheckOption   // Additional options
) (allowed bool, reason string)
```

### 3.2 CheckOptions
```go
type CheckOption func(*checkContext)

// CheckOption examples:
func WithTenant(t *core.Record) CheckOption {
    return func(c *checkContext) {
        c.Tenant = t
    }
}

func WithContextData(key, value any) CheckOption {
    return func(c *checkContext) {
        c.Data[key] = value
    }
}
```

### 3.3 Evaluation Logic
```go
func CheckPermission(ctx context.Context, subject, action, resource *ResourceMetadata, opts ...CheckOption) (bool, string) {
    ctx := &checkContext{}
    for _, opt := range opts {
        opt(ctx)
    }
    
    // 1. Check superuser bypass
    if subject.HasSuperuserAuth() {
        return true, "superuser bypass"
    }
    
    // 2. Find matching role
    roles := findRoles(subject, resource)
    
    // 2. Evaluate each role's permissions
    for _, role := range roles {
        if roleEffect := role.Effect; roleEffect != "" {
            // Use role default effect
        }
        
        for _, perm := range role.Permissions {
            if perm.Field == resource.Field && perm.Effect == EffectAllow {
                // Check conditions
                if conditionsMet(ctx, perm.Conditions, resource) {
                    return true, "role permission matched"
                }
            }
            if perm.Field == resource.Field && perm.Effect == EffectDeny {
                return false, "role permission denied"
            }
        }
    }
    
    // 3. Default deny
    return false, "no matching permission (default deny)"
}
```

---

## 4. Integration with PocketBase API Rules

### 4.1 API Rules Generation
The security metadata can generate PocketBase API rules dynamically:

```go
// Generate API rules from security metadata
func GenerateAPIRules(security *ModuleSecurity) []core.APIRule {
    rules := []core.APIRule{}
    
    for _, res := range security.Resources {
        // View rule
        rules = append(rules, core.APIRule{
            Collection: res.Name,
            ViewRule: generateViewRule(res),
        })
        
        // Create rule
        rules = append(rules, core.APIRule{
            Collection: res.Name,
            CreateRule: generateCreateRule(res),
        })
        
        // Update rule
        rules = append(rules, core.APIRule{
            Collection: res.Name,
            UpdateRule: generateUpdateRule(res),
        })
        
        // Delete rule
        rules = append(rules, core.APIRule{
            Collection: res.Name,
            DeleteRule: generateDeleteRule(res),
        })
    }
    
    return rules
}
```

### 4.2 Generated View Rule Example
```go
// For a "companies" resource with FieldPermission on "salary"
func generateViewRule(res *ResourceMetadata) string {
    clauses := []string{"true"} // default allow
    
    for _, fp := range res.FieldPermissions {
        if fp.Effect == EffectDeny {
            // "salary" field denied for non-admin
            if fp.Field == "salary" {
                clauses = append(clauses, `salary is null or role is admin`)
            }
        }
    }
    
    return strings.Join(clauses, " && ")
}
```

---

## 4. Security Metadata in Module Manifest

### 4.1 Embedding Security in Manifest
```go
type Model struct {
    Name      string          `json:"name"`
    Fields    []Field       `json:"fields"`
    Relations []Relation    `json:"relations"`
    Security  *ResourceMetadata `json:"security,omitempty"` // optional, can be top-level
}

// Top-level in ModuleSecurity
type ModuleManifest struct {
    Module      string         `json:"module"`
    Version     string         `json:"version"`
    Security    *ModuleSecurity `json:"security,omitempty"`
    Models      []Model        `json:"models"`
    // ... other fields
}
```

### 4.2 Example: Company Module Security
```go
var CompanySecurity = &ModuleSecurity{
    Module:      "base",
    Version:     "v1.0.0",
    Resources: []ResourceMetadata{
        {
            Name:        "companies",
            Plural:      "companies",
            Description: "Company records",
            Actions:     []string{"read", "update"},
            Identifier: IdentifierMeta{
                Field: "id",
                Type:  "uuid",
            },
            FieldPermissions: []FieldPermission{
                {
                    Field:   "salary",
                    Effect:  EffectDeny,
                    Roles:   []string{"admin"},
                },
                {
                    Field:   "email",
                    Effect:  EffectAllow,
                    Roles:   []string{"admin", "viewer"},
                },
            },
        },
        {
            Name:        "employees",
            Plural:      "employees",
            Description: "Employee records",
            Actions:     []string{"read", "create", "update", "delete"},
            Identifier: IdentifierMeta{
                Field: "id",
                Type:  "uuid",
            },
            FieldPermissions: []FieldPermission{
                {
                    Field:   "salary",
                    Effect:  EffectDeny,
                    Roles:   []string{"admin", "manager"},
                },
            },
        },
    },
    Roles: []RoleMetadata{
        {
            Name:       "admin",
            Description: "Company administrator",
            Inherits:   []string{},
        },
        {
            Name:       "manager",
            Description: "Department manager",
            Inherits:   []string{"admin"},
        },
        {
            Name:       "viewer",
            Description: "Read-only access",
            Inherits:   []string{},
        },
    },
}
```

---

## 5. Runtime Enforcement

### 5.1 Middleware Integration
```go
// Tenant-aware permission check middleware
func SecurityMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Resolve tenant
        tenant := resolveTenant(r)
        
        // 2. Resolve user role
        role := resolveUserRole(r, tenant)
        
        // 3. Check action permission
        action := r.Method // GET, POST, PUT, DELETE, etc.
        resource := resolveResource(r)
        
        allowed, reason := checkPermission(ctx, subject, action, resource)
        if !allowed {
            http.Error(w, fmt.Sprintf("forbidden: %s", reason), http.StatusForbidden)
            return
        }
        
        // 4. Continue to handler
        next.ServeHTTP(w, r)
    })
}
```

### 5.2 API Handler Pattern
```go
func CompanyHandler(w http.ResponseWriter, r *http.Request) {
    // Permission already checked by middleware
    // ... handler logic
}
```

---

## 6. Security Metadata Validation

### 6.1 Pre-Flight Validation
```go
func ValidateSecurity(security *ModuleSecurity) error {
    errors := []string{}
    
    // 1. Check all referenced fields exist in model
    for _, res := range security.Resources {
        for _, fp := range res.FieldPermissions {
            if !fieldExistsInModel(fp.Field, res.Name) {
                errors = append(errors, fmt.Sprintf("field %s not found in %s", fp.Field, res.Name))
            }
        }
    }
    
    // 2. Check no circular role inheritance
    if hasCircularInheritance(security.Roles) {
        errors = append(errors, "circular role inheritance detected")
    }
    
    // 3. Check all action names are unique per resource
    if hasDuplicateActions(security) {
        errors = append(errors, "duplicate action names")
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("security validation failed: %v", errors)
    }
    
    return nil
}
```

### 6.2 Example Validation Error
```
security validation failed: [field salary not found in companies, circular role inheritance: admin -> manager -> admin]
```

---

## 7. Integration with AENG HTTP API

### 6.1 Route Protection
```go
// In route registration
api.GET("/companies", requirePermission("companies", "read"))
api.POST("/companies", requirePermission("companies", "create"))
api.PATCH("/companies/:id", requirePermission("companies", "update"))
api.DELETE("/companies/:id", requirePermission("companies", "delete"))
```

### 6.2 Middleware Chain
```go
func requirePermission(resource, action string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return security.CheckPermissionMiddleware(resource, action)(next)
    }
}
```

---

## 7. Open Issues

1. **Condition evaluation performance:** Complex conditions with multiple variables could be slow; should we pre-compile?
2. **Inheritance depth:** How many levels of role inheritance are supported? (currently unlimited, but need limits)
3. **Action vs Resource permission:** Should actions be checked separately from resource read/write?
4. **Superuser bypass:** Should superuser always bypass, or can it be configured?
5. **Inheritance vs composition:** Should roles inherit or compose permissions?

---

## 8. References

- [AENG-003 PocketBase Limitations](aeng-003-pocketbase-limitations.md)
- [AENG-006 ORM Architecture](aeng-006-orm-architecture.md)
- [PocketBase API Rules](https://pocketbase.io/docs/api-rules/)
- [OAuth2 Scopes](https://oauth.net/2/scopes/)