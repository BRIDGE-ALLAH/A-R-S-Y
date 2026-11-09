# AENG-014: Define Schema Migration Plan Format for Application Modules

**Status:** Draft
**Owner:** A-R-S-Y Team
**Date:** 2026-11-09
**Depends On:** AENG-006, AENG-001

## Purpose
Define the schema migration plan format that AENG modules use to version and apply database schema changes across tenants. This is the migration executor's input format.

---

## 1. Migration Plan Structure

### 1.1 Top-Level Plan
```go
type MigrationPlan struct {
    Module      string        // Module identifier (e.g., "base", "crm")
    Version     string        // Module version (e.g., "v1.2.3")
    Description string        // Human-readable description
    Migrations  []Migration   // Ordered list of migrations
    Seeds       []*Seed       // Associated seed data
    DependsOn   []string      // Other module plan IDs this depends on
}
```

### 1.2 Migration Definition
```go
type Migration struct {
    ID          string        // Unique migration ID (e.g., "20261109-001")
    Version     string        // Migration version within module
    Description string        // What this migration does
    Up          MigrationFn   // Apply migration
    Down        MigrationFn   // Revert migration
    DependsOn   []string      // Migration IDs this depends on
    Tags        []string      // Tags: "add", "modify", "remove", "data", "schema"
    Safe        bool          // Can run concurrently with other migrations
}
```

### 1.3 Migration Function Signature
```go
type MigrationFn func(ctx context.Context, tenantDB *sql.DB, registry ModelRegistry) error

// Example: Add a field
upAddField := func(ctx context.Context, tenantDB *sql.DB, registry ModelRegistry) error {
    _, err := tenantDB.Exec(`ALTER TABLE companies ADD COLUMN plan TEXT`)
    return err
}
```

---

## 2. Migration Plan Example

### 2.1 Module: Base (v1.2.0)
```go
var BasePlan = &MigrationPlan{
    Module:      "base",
    Version:     "v1.2.0",
    Description: "Initial schema + default data",
    Migrations: []Migration{
        {
            ID:          "20261109-001",
            Version:     "001",
            Description: "Create companies table",
            Up:          upCreateCompanies,
            Down:        downCreateCompanies,
            Tags:        []string{"schema", "create"},
            Safe:        false, // first migration
        },
        {
            ID:          "20261109-002",
            Version:     "002",
            Description: "Add admin user",
            Up:          upSeedAdmin,
            Down:        downSeedAdmin,
            Tags:        []string{"data"},
            Safe:        true,
        },
        {
            ID:          "20261109-003",
            Version:     "003",
            Description: "Add departments table",
            Up:          upCreateDepartments,
            Down:        downCreateDepartments,
            DependsOn:   []string{"20261109-001"}, // depends on companies table
            Tags:        []string{"schema", "create"},
            Safe:        true,
        },
    },
    Seeds: []*Seed{DefaultSeed},
}
```

---

## 3. Migration Execution Strategy

### 3.1 Per-Tenant Execution
```go
func ExecutePlan(ctx context.Context, plan *MigrationPlan, tenantDB *sql.DB) error {
    // 1. Check current version
    currentVersion, err := getAppliedVersion(tenantDB, plan.Module)
    if err != nil { return err }
    
    // 2. If already at or past target version, skip
    if !shouldExecute(currentVersion, plan.Version) {
        return nil
    }
    
    // 3. Sort migrations by dependencies
    ordered := topologicalSort(plan.Migrations)
    
    // 3. Execute each migration in order
    for _, m := range ordered {
        if err := m.Up(ctx, tenantDB, nil); err != nil {
            // Log error but continue? Or abort?
            return fmt.Errorf("migration %s failed: %w", m.ID, err)
        }
        
        // 4. Record as applied
        recordAppliedMigration(tenantDB, m.ID, plan.Module)
    }
    
    // 5. Run associated seeds
    for _, seed := range plan.Seeds {
        if shouldExecuteSeed(seed, tenantDB) {
            if err := RunSeed(ctx, seed, tenantDB); err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

### 3.2 Version Tracking Table
| Column | Type | Description |
|--------|------|-------------|
| module | TEXT | Module identifier |
| migration_id | TEXT | Migration ID (e.g., "20261109-001") |
| applied_at | TEXT (RFC3339) | When it was applied |
| checksum | TEXT (SHA256) | Optional: verify migration wasn't tampered with |

### 3.3 Idempotency Guarantees
| Guarantee | Implementation |
|-----------|----------------|
| Run once | `WHERE NOT EXISTS (SELECT 1 FROM migrations WHERE module=? AND migration_id=?)` |
| Re-run safe | `INSERT OR IGNORE` or `INSERT ... ON CONFLICT DO NOTHING` |
| Rollback track | `down` migration recorded separately |
| Concurrency safe | `BEGIN EXCLUSIVE TRANSACTION` per migration |

---

## 3.4 Concurrency Model

### 3.4.1 Per-Tenant Locking
```go
// Per-tenant migration lock
func withMigrationLock(ctx context.Context, tenantID string, fn func(*sql.DB) error) error {
    db := getTenantDB(tenantID)
    // Use SQLite's built-in locking
    // Or: BEGIN EXCLUSIVE TRANSACTION
    return fn(db)
}
```

### 3.4.2 Module-Level Locking (Optional)
```go
// If multiple tenants might run same migration simultaneously
func withModuleLock(module string, fn func() error) error {
    // Use Redis lock or SQLite WAL lock
    // Only needed if same migration runs across tenants simultaneously
}
```

---

## 4. Migration Fn Signature Details

### 4.1 Available Context
| Parameter | Type | Description |
|-----------|------|-------------|
| `ctx` | `context.Context` | For timeout, cancellation |
| `tenantDB` | `*sql.DB` | Tenant's SQLite database connection |
| `registry` | `ModelRegistry` | Module metadata (models, relations, etc.) |
| `plan` | `*MigrationPlan` | The plan being executed (optional) |

### 4.2 Best Practices
- **Keep migrations short-running** (< 30 seconds)
- **Use transactions** for data migrations
- **Handle partial failures** gracefully
- **Log with context** (tenant ID, migration ID)
- **Never close** `tenantDB` in migration fn

---

## 5. Down Migration Contract

### 5.1 Down Migration Rules
| Rule | Description |
|------|-------------|
| **Best effort** | Down should revert Up, but may not be perfect |
| **Data loss** | Explicitly documented if data will be lost |
| **Dependencies** | Down must undo Up in reverse order |
| **Idempotency** | Down should be safe to re-run |

### 5.2 Example Down Migration
```go
// Up: Add plan column
upAddPlan := func(ctx context.Context, tenantDB *sql.DB, registry ModelRegistry) error {
    _, err := tenantDB.Exec(`ALTER TABLE companies ADD COLUMN plan TEXT`)
    return err
}

// Down: Remove plan column (SQLite doesn't support DROP COLUMN easily!)
downRemovePlan := func(ctx context.Context, tenantDB *sql.DB, registry ModelRegistry) error {
    // SQLite limitation: cannot ALTER TABLE DROP COLUMN
    // Options:
    // 1. Create new table, copy data, drop old (data loss on plan)
    // 2. Set default NULL, just mark as unused
    // 3. Keep column, just stop using it
    _, err := tenantDB.Exec(`PRAGMA deprecated_plan = TRUE`) // custom pragma
    return err
}
```

**Recommendation:** Document limitations in migration description; avoid down migrations for SQLite DDL changes.

---

## 5.3 Alternative: Version Bumping
Instead of Down migrations, use:
- Version bump (major = breaking schema change)
- Feature flags for new columns
- Dual-column period during transition

---

## 6. Migration Plan Versioning

### 6.1 Plan Version Format
```
v{major}.{minor}.{patch}
- Major: Breaking schema changes
- Minor: New migrations added
- Patch: Bug fixes, non-breaking changes
```

### 6.2 Migration from v1.0.0 → v1.1.0
1. Add new migration(s) to plan
2. Run `arsy migration up --target v1.1.0`
3. Record new version in tracking table
4. Old version entries remain (for rollback reference)

### 6.3 Cross-Module Dependencies
```go
var CRMPlan = &MigrationPlan{
    Module:      "crm",
    Version:     "v1.0.0",
    Description: "CRM module schema",
    DependsOn:   []string{"base:v1.2.0"}, // must run after base
}
```

Execution order: `base` → `crm` (topological sort across modules).

---

## 7. Migration CLI

### 7.1 Available Commands
```
arsy migration up           # Run all pending migrations
arsy migration up --target v1.2.3  # Run up to version
arsy migration status      # Show applied migrations
arsy migration down --target v1.0.0  # Rollback to version
arsy migration new --module base --description "Add table"  # Scaffold new migration
```

### 7.2 Output Format
```
$ arsy migration status
module    version    migration_id    applied_at
base      v1.2.0     20261109-001    2026-11-09T10:30:00Z
base      v1.2.0     20261109-002    2026-11-09T10:35:00Z
crm       v1.0.0     20261115-001    2026-11-15T14:20:00Z
```

---

## 8. Integration with Seed Data

### 7.1 Plan + Seed Coupling
```go
var BasePlan = &MigrationPlan{
    Module:      "base",
    Version:     "v1.2.0",
    Migrations:  []Migration{...},
    Seeds:       []*Seed{DefaultSeed, TestSeed},
}
```

Execution order:
1. Run all migrations (schema)
2. Run all seeds (data)
3. Both tracked in same version entry

### 7.2 Conditional Seeds
Seeds with `Environment: "test"` only run when `ARSY_ENV=test`.

---

## 8. Migration Plan Validation

### 8.1 Pre-Flight Checks
```go
func ValidatePlan(plan *MigrationPlan) error {
    // 1. Check all migration IDs are unique
    if hasDuplicateIDs(plan.Migrations) {
        return errors.New("duplicate migration IDs")
    }
    
    // 2. Check no circular dependencies
    if hasCircularDep(plan.Migrations) {
        return errors.New("circular migration dependencies")
    }
    
    // 3. Check FK references exist
    if !refsExist(plan.Migrations) {
        return errors.New("foreign key references don't exist")
    }
    
    // 4. Validate Up/Down symmetry
    if !symmetricUpDown(plan.Migrations) {
        return errors.New("up/down not symmetric")
    }
    
    return nil
}
```

### 8.2 Migration Graph
Visualize as DAG:
```
v1.0.0 → v1.1.0 → v1.2.0
     ↘         ↗
      v1.0.1  (hotfix)
```

---

## 9. Open Issues

1. **SQLite ALTER TABLE limitations:** How to handle DROP COLUMN, RENAME COLUMN?
2. **Data migration helpers:** Built-in helpers for common transforms (rename, split, etc.)?
3. **Migration testing:** How to test migrations without a real DB?
4. **Concurrent migrations:** Is per-tenant locking sufficient, or need module-level?
5. **Cross-module FK:** How to handle FK from module A to module B?

---

## 10. References

- [AENG-006 ORM Architecture](aeng-006-orm-architecture.md)
- [AENG-018 Seed Data Format](aeng-018-seed-data-format.md)
- [PocketBase Migrations](https://pocketbase.io/docs/go-migrations/)
- [SQLite ALTER TABLE](https://www.sqlite.org/lang_altertable.html)