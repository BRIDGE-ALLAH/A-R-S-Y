# AENG-018: Define Idempotent Module Seed Data Format

**Status:** Draft
**Owner:** A-R-S-Y Team
**Date:** 2026-11-09
**Depends On:** AENG-006, AENG-001

## Purpose
Define the format for idempotent module seed data that AENG modules use to initialize database state. Seed data must be safe to run multiple times without duplication or side effects.

---

## 1. Design Principles

### 1.1 Idempotency Requirements
| Principle | Description |
|-----------|-------------|
| **No Duplicates** | Running seed N times produces identical state |
| **Safe Re-run** | Intermediate states are valid at all times |
| **Partial Failure** | One record failure doesn't block others |
| **Environment-Aware** | Different data per environment (dev/test/prod) |

### 1.2 Format Requirements
- **Declarative:** JSON/YAML or Go manifest (preferred)
- **Versioned:** Tied to module version
- **Environment Tags:** `dev`, `test`, `production`
- **Dependencies:** Seed ordering based on collection dependencies

---

## 2. Seed Data Format (Go Manifest)

### 2.1 Top-Level Structure
```go
type Seed struct {
    Module      string        // Module name/identifier
    Version     string        // Module version (e.g., "v1.2.3")
    Environment string        // "dev", "test", "production"
    Dependencies []string     // Seed IDs this depends on
    Records     []SeedRecord  // Records to create/update
}

// SeedRecord represents one record to seed
type SeedRecord struct {
    Collection string            // PocketBase collection name (or module model name)
    Conditions []Condition      // When to skip/conditions to check
    Fields     map[string]any   // Field values
    Upsert     bool             // Use UPSERT (INSERT ... ON CONFLICT) instead of INSERT
    DependsOn  []string         // Other seed record IDs this depends on
}

// Condition defines when a record should be skipped
type Condition struct {
    Field      string            // Collection field to check
    Operator   string            // "exists", "not_exists", "eq", "neq", "gt", "lt"
    Value      any               // Value to compare against
}
```

### 2.2 Example: Company + Department Seed
```go
var CompanySeed = &Seed{
    Module:     "base",
    Version:    "v1.0.0",
    Environment: "production",
    Records: []SeedRecord{
        {
            Collection: "companies",
            Fields: map[string]any{
                "name":  "Acme Corp",
                "plan":  "pro",
                "status": "active",
            },
        },
        {
            Collection: "departments",
            Conditions: []Condition{
                {Field: "name", Operator: "eq", Value: "Engineering"},
            },
            Fields: map[string]any{
                "name":        "Engineering",
                "description": "Software development team",
            },
            DependsOn: []string{"companies"}, // ensures company exists first
        },
    },
}
```

### 2.2 Conditional Seed (Only in Test Environment)
```go
var TestSeed = &Seed{
    Environment: "test",
    Records: []SeedRecord{
        {
            Collection: "settings",
            Conditions: []Condition{
                {Field: "key", Operator: "eq", Value: "test_mode"},
            },
            Fields: map[string]any{
                "value": true,
                "metadata": map[string]any{
                    "bypass_limit": true,
                },
            },
        },
    },
}
```

---

## 3. Idempotency Guarantees

### 3.1 Insert + Upsert Strategy
| Scenario | Action |
|----------|--------|
| Record does not exist | `INSERT INTO collection` |
| Record exists, fields match | No-op, skip |
| Record exists, fields differ | `UPDATE collection SET ...` |

### 3.2 Condition-Based Skipping
```go
// Only create if "superuser" collection is empty
{
    Collection: "superusers",
    Conditions: []Condition{
        {Field: "id", Operator: "eq", Value: "1"},
    },
    Fields: map[string]any{
        "id": "1",
        "email": "admin@example.com",
        "role": "superuser",
    },
}
```

### 3.3 Dependency Ordering
- Seeds are sorted topologically by `DependsOn`
- Circular dependencies return error at module load time
- Dependencies resolved before seed execution

---

## 4. Seed Execution Strategy

### 4.1 Per-Tenant Execution
```go
func RunSeeds(ctx context.Context, tenant *core.Record, seeds []*Seed) error {
    // 1. Resolve dependencies
    ordered := resolveDependencies(seeds)
    
    // 2. Per-tenant: check existing state
    for _, seed := range ordered {
        if !shouldExecute(seed, tenant) {
            continue // already seeded or env mismatch
        }
        
        // 3. Execute with transaction
        tx, err := tenantDB.BeginTxx(ctx, nil)
        if err != nil { return err }
        
        // 4. For each record in seed:
        for _, rec := range seed.Records {
            if !conditionsMet(ctx, rec.Conditions, tenantDB) {
                continue // skip this record
            }
            
            if rec.Upsert {
                // INSERT ... ON CONFLICT (or PocketBase Save)
                upsertRecord(ctx, rec, tenantDB)
            } else {
                // INSERT only, error if exists
                insertRecord(ctx, rec, tenantDB)
            }
        }
        
        // 5. Commit
        if err := tx.Commit(); err != nil {
            tx.Rollback()
            return err
        }
    }
    return nil
}
```

### 4.2 Global vs Per-Tenant
| Scope | When to Use |
|-------|-------------|
| **Global** | Core data: admin users, default settings, lookup tables |
| **Per-Tenant** | Tenant-specific: initial configurations, onboarding data |

### 4.2 Environment Filtering
```go
// Skip seeds not matching current environment
func shouldExecute(seed *Seed, ctx context.Context) bool {
    if seed.Environment == "" {
        return true // no env filter = always execute
    }
    // Check environment variable or context key
    env := ctx.Value("environment") // or os.Getenv("ARSY_ENV")
    return seed.Environment == env.(string)
}
```

---

## 5. Seed Data Examples

### 5.1 Default/Production Seed
```go
var DefaultSeed = &Seed{
    Module:      "base",
    Version:     "v1.0.0",
    Environment: "production",
    Records: []SeedRecord{
        // Default admin user
        {
            Collection: "_superusers", // or custom admin collection
            Fields: map[string]any{
                "email":     "admin@example.com",
                "password":  "$2a$14$...", // bcrypt hash
                "verified":  true,
            },
        },
        // Default settings
        {
            Collection: "settings",
            Fields: map[string]any{
                "key": "site_name",
                "value": "A-R-S-Y Platform",
            },
        },
        {
            Collection: "settings",
            Fields: map[string]any{
                "key": "max_file_size",
                "value": 10, // 10 MB
            },
        },
    },
}
```

### 5.2 Test Seed (Minimal)
```go
var TestSeed = &Seed{
    Module:      "base",
    Version:     "v1.0.0",
    Environment: "test",
    Records: []SeedRecord{
        {
            Collection: "companies",
            Conditions: []Condition{
                {Field: "name", Operator: "eq", Value: "Acme Corp"},
            },
            Fields: map[string]any{
                "name": "Acme Corp",
            },
        },
    },
}
```

---

## 6. Integration with Migration Executor

### 6.1 Seed Execution Order
```
1. Run module migrations (create schema)
2. Run seeds (insert default data)
3. Both are versioned together (v1.2.3)
4. Re-running: migrations are idempotent, seeds are idempotent
```

### 6.2 Version Bumping
| Change Type | Action |
|-------------|--------|
| New seed records added | Increment patch version |
| Seed logic changes | Increment minor version |
| Schema changes | Increment major version (also update migration) |

---

## 7. Seed Validation

### 7.1 Pre-Flight Checks
```go
func ValidateSeeds(seeds []*Seed) error {
    // 1. Check all collections exist
    for _, seed := range seeds {
        for _, rec := range seed.Records {
            if !collectionExists(rec.Collection) {
                return fmt.Errorf("collection %s does not exist", rec.Collection)
            }
        }
    }
    
    // 2. Check no circular dependencies
    if hasCircularDep(seeds) {
        return errors.New("circular seed dependencies detected")
    }
    
    // 3. Validate field types against schema
    if !validatesAgainstSchema(seeds) {
        return errors.New("seed field types don't match schema")
    }
    
    return nil
}
```

### 7.2 Runtime Checks
- Each seed record's `Conditions` evaluated against current DB state
- Failed conditions = record skipped (not error)
- Failed upsert = error returned, transaction rolled back

---

## 8. Seed Execution API

### 7.3 CLI Command
```
arsy seed run --module base --environment production
arsy seed run --module base --environment test --skip-conditions
arsy seed status --show which seeds have been applied
```

### 7.4 Programmatic API
```go
// Run all seeds for a tenant
func RunAllSeeds(ctx context.Context, tenantID string) error {
    // Load module manifests
    manifests := loadModuleManifests()
    
    // Run per-tenant
    for _, mod := range manifests {
        if err := RunSeeds(ctx, tenantID, mod.Seeds); err != nil {
            return err
        }
    }
    return nil
}
```

---

## 8. Migration from Legacy Seed Format

### 7.1 Legacy (Pre-AENG)
```go
// Old format - no idempotency, no conditions
func seedLegacy() {
    db := getDB()
    db.MustExec(`INSERT INTO companies (name, plan, status) VALUES ('Acme Corp', 'pro', 'active')`)
    db.MustExec(`INSERT INTO settings (key, value) VALUES ('site_name', 'A-R-S-Y')`)
}
```

### 7.2 New Format Migration
```go
// New format with idempotency
func seedNew() {
    // Run existing arsy seed runner
    // Old seeds become "first-run" seeds
    // Subsequent runs are no-ops
}
```

**Migration Strategy:**
1. Existing data stays as-is
2. New seeds run idempotently on next deploy
3. After all tenants upgraded, remove legacy seed functions

---

## 8. Open Issues

1. **Condition operators:** Support for `contains`, `in`, `not_in` on text fields?
2. **Bulk seed records:** Can a single seed record create multiple records?
3. **Seed metadata:** Should we track `seed_id`, `run_count`, `last_run_at` in a seed tracking table?
4. **Environment detection:** Automatic vs explicit environment flag?
5. **Undo/unseed:** Is there a need to reverse seed data?

---

## 8. References

- [AENG-006 ORM Architecture](aeng-006-orm-architecture.md)
- [AENG-014 Schema Migration Plan](aeng-014-schema-migration-plan.md) - upcoming
- [PocketBase Migrations](https://pocketbase.io/docs/go-migrations/)