# AENG-006: ORM Architecture Decision Matrix

**Status:** Draft
**Owner:** A-R-S-Y Team
**Date:** 2026-11-09

## Purpose
Evaluate and decide on the ORM/persistence architecture for AENG modules. Compare approaches for mapping Go models to PocketBase collections and tenant databases.

---

## 1. Requirements

| Requirement | Priority | Description |
|-------------|----------|-------------|
| Type-safe Go models | P0 | Compile-time field access, no string keys |
| Multi-tenant isolation | P0 | Per-tenant database routing |
| Schema migration per module | P0 | Versioned, idempotent, rollback-safe |
| Relations (has-one, has-many, many-to-many) | P1 | Declarative, lazy/eager loading |
| Computed/derived fields | P1 | Views, virtual fields, actions |
| Query builder | P1 | Type-safe filters, pagination, sorting |
| Seed data management | P1 | Idempotent, environment-aware |
| Performance | P2 | Connection pooling, query optimization |
| Zero external dependencies | P2 | Minimal deps beyond PocketBase |

---

## 2. Approach Comparison

### Option A: PocketBase Native (Current)
**Use `core.Record` + `dbx` query builder directly**

| Pros | Cons |
|------|------|
| ✅ Zero dependencies | ❌ No type safety (string field names) |
| ✅ Full PocketBase feature access | ❌ Manual relation handling |
| ✅ Works with all field types | ❌ No schema migration management |
| ✅ Mature, battle-tested | ❌ Verbose boilerplate |

**Verdict:** Baseline - insufficient for AENG module system

---

### Option B: Code Generation (sqlc-style)
**Generate Go structs + methods from schema definitions**

| Pros | Cons |
|------|------|
| ✅ Full type safety | ❌ Requires codegen step |
| ✅ Custom query methods | ❌ Schema drift risk |
| ✅ Fast execution | ❌ Migration sync complexity |
| ✅ No runtime reflection | ❌ Generated code maintenance |

**Tools:** `sqlc`, `sqlboiler`, custom generator

---

### Option C: Reflection-Based ORM (GORM-style)
**Struct tags + runtime reflection**

| Pros | Cons |
|------|------|
| ✅ Familiar API | ❌ Reflection overhead |
| ✅ Auto-migration | ❌ PocketBase schema mismatch |
| ✅ Rich relation support | ❌ Not designed for multi-tenant |
| ✅ Large ecosystem | ❌ Heavy dependency |

**Verdict:** Poor fit - PocketBase schema model differs significantly from SQL tables

---

### Option D: Hybrid - Declarative Schema + Generated Accessors ⭐ RECOMMENDED
**Define schema in Go (manifest), generate type-safe accessors, use PocketBase for persistence**

| Pros | Cons |
|------|------|
| ✅ Single source of truth (Go manifest) | ❌ Codegen pipeline needed |
| ✅ Type-safe models + PocketBase compatibility | ❌ Initial setup cost |
| ✅ Module-level migrations from manifest | ❌ Learning curve |
| ✅ Relations, computed fields, actions as metadata | ✅ Can be built incrementally |
| ✅ Seed data from manifest | |

**Architecture:**
```
Module Manifest (Go) → Codegen → Type-safe Models + Migration SQL + Seed Data
```

---

## 3. Recommended Architecture: Declarative Manifest

### 3.1 Module Manifest (Single Source of Truth)
```go
// Module defines an AENG module
type Module struct {
    Name        string
    Version     string
    Models      []Model
    Relations   []Relation
    Views       []View
    Actions     []Action
    Seeds       []Seed
}

// Model maps to a PocketBase collection
type Model struct {
    Name        string          // Collection name
    TableName   string          // SQLite table name
    Fields      []Field
    Indexes     []Index
    APIRules    APIRules
}

// Field defines a model field
type Field struct {
    Name        string
    Type        FieldType       // Text, Number, Bool, Relation, etc.
    Required    bool
    Unique      bool
    Default     any
    Options     FieldOptions    // Max, Min, Pattern, Values, RelationTarget
    Computed    bool            // Derived field, not stored
    ComputedFn  string          // Function name for computation
}
```

### 3.2 Codegen Outputs
| Output | Purpose |
|--------|---------|
| `models_gen.go` | Type-safe Go structs with field constants |
| `migrations_gen.go` | PocketBase migration up/down functions |
| `seeds_gen.go` | Idempotent seed data functions |
| `queries_gen.go` | Type-safe query builders per model |
| `registry_gen.go` | Module registration with relations/views/actions |

### 3.3 Runtime Components
| Component | Responsibility |
|-----------|----------------|
| `ModelRegistry` | Load module manifests, provide metadata |
| `MigrationExecutor` | Apply module migrations per tenant |
| `SeedRunner` | Execute seed data idempotently |
| `QueryBuilder` | Type-safe query construction |
| `RelationLoader` | Eager/lazy load relations |
| `ActionExecutor` | Execute declared actions |

---

## 4. Multi-Tenant Integration

```
Tenant Request
      │
      ▼
resolveAuthorizedTenantDatastore() ───▶ tenant.DB() (sql.DB)
      │
      ▼
persist.New(registry, dbx.NewFromDB(tenantDB, "sqlite"))
      │
      ▼
Type-safe CRUD via generated accessors
```

### Key Points
- Each tenant has **own SQLite database** (not shared PocketBase DB)
- Module collections created **per tenant** via migrations
- Model registry shared (metadata), data isolated

---

## 5. Migration Strategy

| Aspect | Decision |
|--------|----------|
| Versioning | Per-module semantic version (v1.2.3) |
| Ordering | Dependency graph topological sort |
| Idempotency | `CREATE IF NOT EXISTS`, `ALTER IF NOT EXISTS` |
| Rollback | Down migration for each up |
| Data migration | Explicit migration steps with SQL |
| Tenant apply | Parallel per tenant, serialized per tenant |

---

## 6. Relation Handling

| Relation Type | Storage | Loading |
|---------------|---------|---------|
| Has-One | FK on child | Eager (join) or Lazy |
| Has-Many | FK on child | Lazy (separate query) |
| Many-to-Many | Junction table | Lazy or eager with limit |

**Generated Code:**
```go
// Company has many Employees
func (c *Company) Employees(q *QueryBuilder) ([]*Employee, error) {
    return q.Model("employee").Filter("company_id = ?", c.ID).All()
}
```

---

## 7. Computed Fields & Actions

### Computed Fields (Views)
```go
Field{
    Name: "full_name",
    Type: "string",
    Computed: true,
    ComputedFn: "computeFullName",
}
```

### Actions (Business Logic)
```go
Action{
    Name: "archive",
    Input: ArchiveInput,
    Output: ArchiveOutput,
    Handler: archiveCompany,
}
```

---

## 8. Decision Matrix Summary

| Criterion | Option A (Native) | Option B (Codegen) | Option C (Reflection) | Option D (Hybrid) |
|-----------|-------------------|---------------------|----------------------|-------------------|
| Type Safety | ❌ | ✅ | ✅ | ✅ |
| Multi-Tenant | Manual | Manual | Poor | ✅ Built-in |
| Migrations | Manual | Generated | Auto (risky) | ✅ From manifest |
| Relations | Manual | Generated | ✅ | ✅ Generated |
| Computed Fields | ❌ | ❌ | ❌ | ✅ |
| Actions | ❌ | ❌ | ❌ | ✅ |
| Seed Data | Manual | Generated | ❌ | ✅ |
| Dependencies | None | Codegen tool | Heavy (GORM) | Codegen tool |
| Learning Curve | Low | Medium | Medium | Medium |
| PocketBase Native | ✅ | ✅ | ❌ | ✅ |

---

## 9. Implementation Plan

### Phase 1: Manifest Definition (Week 1)
- [ ] Define `Module`, `Model`, `Field`, `Relation`, `View`, `Action`, `Seed` types
- [ ] Create example module manifest
- [ ] Validate against existing PocketBase collections

### Phase 2: Code Generator (Week 2)
- [ ] Build Go code generator (uses `text/template` or `gotemplate`)
- [ ] Generate: models, migrations, seeds, queries, registry
- [ ] Integrate with `go generate`

### Phase 3: Runtime Components (Week 3)
- [ ] `ModelRegistry` - load and serve manifests
- [ ] `MigrationExecutor` - apply per tenant
- [ ] `SeedRunner` - idempotent execution
- [ ] `QueryBuilder` - type-safe filters

### Phase 4: Integration (Week 4)
- [ ] Wire into `httpapi` engine routes
- [ ] Add tenant-scoped model access
- [ ] Tests: multi-tenant isolation, migrations, seeds

### Phase 5: Module Examples (Week 5)
- [ ] Base module (Company, User, etc.)
- [ ] CRM module example
- [ ] Documentation

---

## 10. Open Questions

1. **Codegen tool:** Custom vs `sqlc` vs `go:generate` with template?
2. **Migration format:** PocketBase migration functions vs raw SQL files?
3. **Relation loading:** Eager by default or explicit?
4. **View collections:** PocketBase view collections vs computed at query time?
5. **Action execution:** Synchronous HTTP handler vs async job queue?

---

## 11. References

- [PocketBase Core Types](https://pkg.go.dev/github.com/pocketbase/pocketbase/core)
- [DBX Query Builder](https://github.com/go-dbx/dbx)
- [sqlc](https://sqlc.dev/)
- [Ent ORM](https://entgo.io/) - for API inspiration
- [GORM](https://gorm.io/) - for API inspiration (not for use)