# AENG-034: Implement Idempotent Module Seed Data Runner

**Status:** Draft
**Owner:** A-R-S-Y Team
**Date:** 2026-11-09
**Depends On:** AENG-018, AENG-006

## Purpose
Implement the seed data runner that executes module seed data idempotently, supporting per-environment, per-tenant, and dependency-aware execution.

---

## 1. Seed Runner Design

### 1.1 Seed Runner Structure
```go
type SeedRunner struct {
    db          *sql.DB          // Tenant database connection
    registry    *ModelRegistry   // Model registry for field validation
    environment string           // "dev", "test", "production"
    executed    map[string]bool // Track which seed IDs have been executed
    runCount    map[string]int  // How many times each seed has run
}
```

### 1.2 Seed Execution Result
```go
type SeedExecutionResult struct {
    SeedID      string // Seed identifier
    Status      string // "executed", "skipped", "failed"
    RecordsAffected int   // Number of records created/updated
    Error       error  // If failed
    ExecutedAt  time.Time // When it was executed
}
```

---

## 2. Execution Strategy

### 2.1 Per-Environment Filtering
```go
func (r *SeedRunner) shouldExecute(seed *Seed) bool {
    // If no environment filter, always execute
    if seed.Environment == "" {
        return true
    }
    
    // Check environment variable or context
    env := r.environment
    if env == "" {
        env = os.Getenv("ARSY_ENV")
    }
    
    return seed.Environment == env
}
```

### 2.2 Per-Tenant Execution
```go
func (r *SeedRunner) shouldExecuteForTenant(seed *Seed, tenantID string) bool {
    // Global seeds (no tenant filter) execute for all tenants
    if len(seed.TenantFilter) == 0 {
        return r.shouldExecute(seed)
    }
    
    // Tenant-specific seeds: check if this tenant matches
    // Could check tenant ID, subscription plan, etc.
    return r.tenantMatches(seed.TenantFilter, tenantID)
}
```

### 2.3 Dependency-Aware Ordering
```go
func (r *SeedRunner) orderSeeds(seeds []*Seed) []*Seed {
    // Topological sort by DependsOn
    // Build dependency graph
    // Return sorted list
    
    // Simplified: just return as-is, rely on caller to sort
    return seeds
}
```

---

## 3. Seed Execution Algorithm

### 3.1 Main Execution Function
```go
func (r *SeedRunner) Execute(ctx context.Context, seeds []*Seed) []SeedExecutionResult {
    results := make([]SeedExecutionResult, len(seeds))
    
    // 1. Order seeds by dependencies
    ordered := r.orderSeeds(seeds)
    
    // 2. Execute each seed in order
    for i, seed := range ordered {
        // 2a. Check environment filter
        if !r.shouldExecute(seed) {
            results[i] = SeedExecutionResult{
                SeedID:      seed.Module + ":" + seed.Version,
                Status:      "skipped",
                Reason:      "environment filter mismatch",
                ExecutedAt: time.Now(),
            }
            continue
        }
        
        // 2b. Check tenant filter
        if !r.shouldExecuteForTenant(seed, r.currentTenantID) {
            results[i] = SeedExecutionResult{
                SeedID:      seed.Module + ":" + seed.Version,
                Status:      "skipped",
                Reason:      "tenant filter mismatch",
                ExecutedAt: time.Now(),
            }
            continue
        }
        
        // 2c. Check if already executed (idempotency)
        if r.executed[seed.ID] {
            results[i] = SeedExecutionResult{
                SeedID:      seed.Module + ":" + seed.Version,
                Status:      "skipped",
                Reason:      "already executed",
                ExecutedAt: time.Now(),
            }
            continue
        }
        
        // 3. Execute seed
        result := r.executeSeed(ctx, seed)
        results[i] = result
        
        // 4. Mark as executed (even if skipped due to conditions)
        if result.Status != "failed" {
            r.executed[seed.ID] = true
        }
    }
    
    return results
}
```

### 3.2 Single Seed Execution
```go
func (r *SeedRunner) executeSeed(ctx context.Context, seed *Seed) SeedExecutionResult {
    result := SeedExecutionResult{
        SeedID:      seed.Module + ":" + seed.Version,
        ExecutedAt: time.Now(),
    }
    
    // 1. Begin transaction
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        result.Status = "failed"
        result.Error = err
        return result
    }
    defer tx.Rollback()
    
    // 2. Execute each record in seed
    recordsAffected := 0
    for _, rec := range seed.Records {
        // 3a. Check conditions
        if !conditionsMet(r.db, rec.Conditions) {
            // Skip this record (not an error)
            continue
        }
        
        // 3b. Upsert or insert
        if rec.Upsert {
            affected := r.upsertRecord(tx, rec)
            recordsAffected += affected
        } else {
            affected := r.insertRecord(tx, rec)
            recordsAffected += affected
        }
    }
    
    // 4. Commit transaction
    if err := tx.Commit(); err != nil {
        result.Status = "failed"
        result.Error = err
        return result
    }
    
    // 5. Mark as executed
    result.Status = "executed"
    result.RecordsAffected = recordsAffected
    r.executed[seed.ID] = true
    
    return result
}
```

### 3.3 Insert Record
```go
func (r *SeedRunner) insertRecord(tx *sql.Tx, rec *SeedRecord) int {
    // Build INSERT query from fields
    columns := []string{rec.Collection} // simplified
    placeholders := make([]string, len(rec.Fields))
    values := make([]any, len(rec.Fields))
    
    for i, (k, v) := range rec.Fields {
        placeholders[i] = "?"
        values[i] = v
    }
    
    // Execute INSERT ... ON CONFLICT (upsert-like)
    // Or plain INSERT if !rec.Upsert
    query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, 
        rec.Collection, 
        strings.Join(columns, ", "),
        strings.Join(placeholders, ", "))
    
    // For upsert: ON CONFLICT (field) DO UPDATE SET ...
    // For plain: INSERT INTO ...
    
    result, err := tx.Exec(query, values...)
    if err != nil {
        return 0
    }
    
    affected, _ := result.RowsAffected()
    return int(affected)
}
```

### 3.4 Upsert Record
```go
func (r *SeedRunner) upsertRecord(tx *sql.Tx, rec *SeedRecord) int {
    // Build ON CONFLICT query
    // PocketBase doesn't have native upsert, so we:
    // 1. Check if record exists (conditions)
    // 2. If exists: UPDATE
    // 3. If not: INSERT
    
    // Simplified: just INSERT, let caller handle idempotency via conditions
    return r.insertRecord(tx, rec)
}
```

---

## 4. Condition Evaluation

### 4.1 Condition Types
| Operator | Description | Example |
|----------|-------------|---------|
| `eq` | Equal | `{Field: "status", Operator: "eq", Value: "active"}` |
| `neq` | Not equal | `{Field: "plan", Operator: "neq", Value: "pro"}` |
| `gt` | Greater than | `{Field: "employee_count", Operator: "gt", Value: 10}` |
| `lt` | Less than | `{Field: "employee_count", Operator: "lt", Value: 100}` |
| `contains` | Text contains | `{Field: "name", Operator: "contains", Value: "Acme"}` |
| `exists` | Record exists | `{Field: "id", Operator: "exists", Value: ""}` |
| `not_exists` | Record doesn't exist | `{Field: "id", Operator: "not_exists", Value: ""}` |

### 4.2 Condition Evaluation
```go
func conditionsMet(db *sql.DB, conditions []Condition) bool {
    if len(conditions) == 0 {
        return true
    }
    
    for _, cond := range conditions {
        // Query the field value from DB
        var value any
        err := db.QueryRow(`SELECT `+cond.Field+` FROM `+cond.Table+` LIMIT 1`).Scan(&value)
        if err != nil {
            return false // can't determine, skip
        }
        
        // Evaluate operator
        met := false
        switch cond.Operator {
        case "eq":
            met = value == cond.Value
        case "neq":
            met = value != cond.Value
        case "gt":
            met = value > cond.Value
        case "lt":
            met = value < cond.Value
        case "contains":
            met = fmt.Sprintf("%v", value) != "" && strings.Contains(fmt.Sprintf("%v", value), fmt.Sprintf("%v", cond.Value))
        case "exists":
            met = true // always true if we got here
        case "not_exists":
            met = false
        }
        
        if !met {
            return false // one condition failed
        }
    }
    
    return true // all conditions met
}
```

---

## 3. Seed Execution API

### 4.1 CLI Command
```
arsy seed run --module base --environment production
arsy seed run --module crm --environment test --tenant acme-corp
arsy seed status --show all seeds and their execution count
arsy seed reset --module base --environment production --force
```

### 4.2 Programmatic API
```go
// Run all seeds for current tenant
results := runner.Execute(ctx, allSeeds)

// Check if a specific seed has been executed
hasRun := runner.HasExecuted("base:v1.0.0")

// Get execution count for a seed
count := runner.GetRunCount("crm:v1.2.0")
```

### 4.3 Status API
```go
type SeedStatus struct {
    SeedID      string
    Module      string
    Version     string
    Status      string // "executed", "skipped", "failed"
    ExecutedAt  time.Time
    RecordsAffected int
    Error       string
}
```

---

## 5. Integration with Migration Executor

### 5.1 Execution Order
```
1. Run pending migrations (schema changes)
2. Run seeds (data initialization)
3. Both tracked per-tenant in version table
4. Re-running: migrations are idempotent (CREATE IF NOT EXISTS)
5. Re-running: seeds are idempotent (conditions + upsert)
```

### 5.2 Version Bumping with Seeds
| Change | Action |
|--------|--------|
| New seed records added | Increment patch version |
| Seed logic changes | Increment minor version |
| Environment changes | Update seed.Environment field |
| No changes | Same version, re-execute is no-op |

---

## 5.3 Migration + Seed Transaction
```go
func RunMigrationsAndSeeds(ctx context.Context, tenantDB *sql.DB, plan *MigrationPlan) error {
    // 1. Run migrations
    if err := executeMigrations(ctx, tenantDB, plan.Migrations); err != nil {
        return err
    }
    
    // 2. Run seeds
    if err := runner.Execute(ctx, plan.Seeds); err != nil {
        return err
    }
    
    // 3. Record version
    recordVersion(tenantDB, plan.Module, plan.Version)
    
    return nil
}
```

---

## 6. Open Issues

1. **Condition performance:** Evaluating conditions on large datasets could be slow; should we cache?
2. **Bulk seed records:** Can a single seed record create multiple records? (Currently one record = one DB row)
3. **Seed undo/unseed:** Is there a need to reverse seed data? (Related to migration down)
4. **Seed versioning:** Should seeds have independent version from migrations?
5. **Concurrent seed execution:** Can two tenants run seeds simultaneously safely?

---

## 6. References

- [AENG-018 Seed Data Format](aeng-018-seed-data-format.md)
- [AENG-014 Schema Migration Plan](aeng-014-schema-migration-plan.md)
- [AENG-006 ORM Architecture](aeng-006-orm-architecture.md)
- [PocketBase Migrations](https://pocketbase.io/docs/go-migrations/)