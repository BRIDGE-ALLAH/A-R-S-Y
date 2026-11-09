# AENG-035: Enforce Model and Record Security Metadata at Runtime

**Status:** Draft
**Owner:** A-R-S-Y Team
**Date:** 2026-11-09
**Depends On:** AENG-019, AENG-006

## Purpose
Implement runtime enforcement of model and record security metadata, ensuring that all API operations respect the security contracts defined in module manifests. This bridges the gap between declarative metadata and actual HTTP handler enforcement.

---

## 1. Runtime Security Enforcement Architecture

### 1.1 Enforcement Middleware Chain
```go
func SecurityEnforcementMiddleware(r *ModelRegistry) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. Resolve tenant from request
            tenant := resolveTenant(r)
            if tenant == nil {
                http.Error(w, "tenant required", http.StatusBadRequest)
                return
            }
            
            // 2. Resolve user role
            userRole := resolveUserRole(r, tenant)
            
            // 3. Extract resource and action from path
            resource := resolveResource(r)
            action := resolveAction(r)
            
            // 3. Check security metadata
            allowed, reason := checkSecurityMeta(ctx, r, tenant, resource, action)
            if !allowed {
                http.Error(w, fmt.Sprintf("forbidden: %s", reason), http.StatusForbidden)
                return
            }
            
            // 4. Add security headers/context for handler
            r = r.WithContext(securityContext{tenant, userRole, action, resource})
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### 1.2 Security Check Function
```go
func checkSecurityMeta(
    ctx context.Context,
    r *http.Request,
    tenant *core.Record,
    resource *ResourceMeta,
    action string,
) (allowed bool, reason string) {
    // 1. Superuser bypass
    if r.Context().Value("is_superuser").(bool) {
        return true, "superuser bypass"
    }
    
    // 2. Get user role
    userRole := resolveUserRole(r.Context())
    
    // 3. Check resource-level permissions
    if resource != nil {
        // Check resource-level allow/deny
        allowed, ok := checkResourcePermission(userRole, resource, action)
        if !ok {
            return false, "no resource permission config"
        }
        if !allowed {
            return false, "resource permission denied"
        }
    }
    
    // 4. Check action-specific permissions
    if action != "" {
        allowed, ok := checkActionPermission(userRole, action)
        if !ok {
            return false, "no action permission config"
        }
        if !allowed {
            return false, "action permission denied"
        }
    }
    
    // 5. Default allow (if no restrictions configured)
    return true, "no restrictions configured"
}
```

---

## 2. Resource Permission Checking

### 2.1 Resource Permission Structure
```go
type ResourcePermission struct {
    Resource  string   // e.g. "companies", "employees"
    Actions   []string // e.g. ["read", "write", "delete"]
    Roles     []string // e.g. ["admin", "manager"]
    Conditions []Condition // When this permission applies
}
```

### 2.2 Check Resource Permission
```go
func checkResourcePermission(userRole string, resource *ResourceMeta, action string) (bool, string) {
    // Find applicable permissions for this resource+action
    for _, perm := range resource.Permissions {
        // Check if action matches
        actionMatches := contains(perm.Actions, action) || contains(perm.Actions, "*")
        
        // Check if role matches
        roleMatches := contains(perm.Roles, userRole) || contains(perm.Roles, "superuser")
        
        // Check conditions
        conditionsMet := true
        if len(perm.Conditions) > 0 {
            conditionsMet = checkConditions(perm.Conditions, r.Context())
        }
        
        if actionMatches && roleMatches && conditionsMet {
            // Effect: allow or deny
            if perm.Effect == EffectAllow {
                return true, ""
            } else if perm.Effect == EffectDeny {
                return false, "denied by resource permission"
            }
        }
    }
    
    // Default: deny if no matching permission found
    return false, "no matching resource permission"
}
```

---

## 3. Action Permission Checking

### 3.1 Action Permission Structure
```go
type ActionPermission struct {
    Action      string   // e.g. "archive", "delete", "export"
    Effect      Effect   // "allow" or "deny"
    Roles      []string // Roles allowed
    FieldPerms  []FieldPermission // Additional field-level checks
}
```

### 3.2 Check Action Permission
```go
func checkActionPermission(userRole string, action *Action) (bool, string) {
    // Look up action permission from module security
    // ...
    
    // Check roles
    allowed := false
    for _, role := range action.Permission.Roles {
        if userRole == role || contains(userRole, "superuser") {
            allowed = true
            break
        }
    }
    
    // Check field permissions if action has them
    if allowed && len(action.Permission.FieldPerms) > 0 {
        // Check each field permission
        for _, fp := range action.Permission.FieldPerms {
            if !checkFieldPermission(fp, r.Context()) {
                allowed = false
                break
            }
        }
    }
    
    if allowed {
        return true, ""
    }
    return false, "action denied by role or field permissions"
}
```

---

## 4. Field-Level Security Enforcement

### 4.1 Field Permission Structure
```go
type FieldPermission struct {
    Field       string   // e.g. "salary", "email", "ssn"
    Effect      Effect   // "allow" or "deny"
    Roles       []string // Roles allowed
    Conditions  []Condition // When permission applies
}
```

### 4.2 Enforce Field-Level Security in Handlers
```go
// Before returning record data, filter fields based on user role
func filterRecordFields(record *core.Record, userRole string, fieldPerms []FieldPermission) *core.Record {
    if len(fieldPerms) == 0 {
        return record // no field restrictions
    }
    
    filtered := core.NewRecord(record.Collection)
    
    for _, field := range record.Fields {
        // Check if this field has a permission restriction
        hasPerm := false
        for _, fp := range fieldPerms {
            if fp.Field == field.Name {
                hasPerm = true
                // Check conditions
                if !conditionsMet(r.Context(), fp.Conditions) {
                    // Condition not met; depend on effect
                    if fp.Effect == EffectDeny {
                        // Don't include this field
                        hasPerm = true // mark as restricted
                        break
                    }
                }
                // If effect is allow and condition met, include
                if fp.Effect == EffectAllow && conditionsMet(r.Context(), fp.Conditions) {
                    hasPerm = true
                    break
                }
            }
        }
        
        // Include field only if permitted
        if hasPerm {
            filtered.Set(field.Name, record.Get(field.Name))
        }
    }
    
    return filtered
}
```

### 4.2 Example: Salary Field Restriction
```go
// Company record view for different roles
func CompanyViewForRole(record *core.Record, userRole string) *core.Record {
    fieldPerms := []FieldPermission{
        {
            Field:   "salary",
            Effect:  EffectDeny,
            Roles:   []string{"viewer"},
        },
        {
            Field:   "salary",
            Effect:  EffectAllow,
            Roles:   []string{"admin", "manager"},
        },
        {
            Field:   "name",
            Effect:  EffectAllow,
            Roles:   []string{"viewer", "admin", "manager"},
        },
        {
            Field:   "email",
            Effect:  EffectAllow,
            Roles:   []string{"admin", "manager", "viewer"},
        },
    }
    
    return filterRecordFields(record, userRole, fieldPerms)
}
```

---

## 5. API Rule Generation from Metadata

### 5.1 Generate PocketBase API Rules
```go
func GeneratePocketBaseAPIRules(security *ModuleSecurity) []core.APIRule {
    rules := []core.APIRule{}
    
    for _, res := range security.Resources {
        // View rule - who can see this collection
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

### 5.2 Generate View Rule
```go
func generateViewRule(res *ResourceMeta) string {
    clauses := []string{}
    
    // Add field permissions
    for _, fp := range res.FieldPermissions {
        if fp.Effect == EffectDeny {
            clauses = append(clauses, fmt.Sprintf(`%s is null`, fp.Field))
        }
    }
    
    // Add role-based clauses
    for _, role := range res.Roles {
        clauses = append(clauses, fmt.Sprintf(`role is %s`, role))
    }
    
    if len(clauses) == 0 {
        return "true" // default allow
    }
    
    return strings.Join(clauses, " && ")
}
```

### 5.3 Generate Create/Update/Delete Rules
```go
func generateCreateRule(res *ResourceMeta) string {
    // Similar pattern: check roles + field permissions
    clauses := []string{}
    
    for _, fp := range res.FieldPermissions {
        if fp.Effect == EffectDeny && strings.Contains(fp.Field, "password|ssn|secret") {
            clauses = append(clauses, fmt.Sprintf(`%s is null`, fp.Field))
        }
    }
    
    // Allow if any role matches
    for _, role := range res.Roles {
        clauses = append(clauses, fmt.Sprintf(`role is %s`, role))
    }
    
    if len(clauses) == 0 {
        return "true"
    }
    
    return strings.Join(clauses, " && ")
}
```

---

## 6. Runtime Security Metadata Validation

### 5.1 Pre-Flight Validation
```go
func ValidateSecurityMeta(r *ModelRegistry) error {
    var errors []string
    
    // 1. Check all resource fields exist in models
    for _, mod := range r.getModules() {
        for _, res := range r.getSecurity(mod.ID) {
            for _, fp := range res.FieldPermissions {
                if !fieldExistsInModel(fp.Field, r.getModel(res.Name)) {
                    errors = append(errors, fmt.Sprintf("field %s not found in model %s", fp.Field, res.Name))
                }
            }
        }
    }
    
    // 2. Check role inheritance is acyclic
    if hasCircularRoleInheritance(r) {
        errors = append(errors, "circular role inheritance detected")
    }
    
    // 3. Check all referenced actions exist
    if !allActionsExist(r) {
        errors = append(errors, "referenced actions not found in module manifests")
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("security metadata validation failed: %v", errors)
    }
    
    return nil
}
```

### 5.2 Example Validation Error
```
security metadata validation failed: [field salary not found in companies, circular role inheritance: admin -> manager -> admin]
```

---

## 6. Integration with Existing HTTP API

### 6.1 Existing Handler Pattern
Existing handlers in `arsy/httpapi/` can be wrapped with security enforcement:

```go
// Before: bare handler
func CompanyListHandler(w http.ResponseWriter, r *http.Request) {
    // ... get companies, return JSON
}

// After: wrapped with security
func CompanyListHandler secured(w http.ResponseWriter, r *http.Request) {
    // Security middleware already ran - permissions checked
    // ... get companies, apply field filtering, return JSON
}
```

### 6.2 Wrapper Function
```go
func WrapHandlerWithSecurity(handler http.HandlerFunc, r *ModelRegistry) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Security enforcement already happened via middleware
        // Just execute the handler
        handler(w, r)
    }
}
```

---

## 7. Open Issues

1. **Performance overhead:** Evaluating security metadata on every request could be slow; should we cache per-tenant?
2. **Condition evaluation:** Complex conditions with subqueries may need special handling
3. **Superuser configurability:** Should superuser bypass be configurable or always on?
4. **Action chaining:** Can one action trigger another action's security checks?
5. **Audit logging:** Should security checks be logged for audit purposes?

---

## 6. References

- [AENG-019 Module Security Metadata](aeng-019-module-security-metadata.md)
- [AENG-006 ORM Architecture](aeng-006-orm-architecture.md)
- [PocketBase API Rules](https://pocketbase.io/docs/api-rules/)
- [OAuth2 Scopes and Claims](https://oauth.net/2/scopes/)