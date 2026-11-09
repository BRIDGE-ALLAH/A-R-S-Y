# A-R-S-Y Product Architecture Decision Record

**Task:** ARSY-002 — Define A-R-S-Y product architecture and public contract  
**Status:** In Progress  
**Depends on:** ARSY-001 (completed)  
**Generated:** 2026-11-09

---

## 1. Product Layers (Bottom-up)

```
┌─────────────────────────────────────────────────────────────────┐
│  LAYER 7: Developer Experience (DX)                            │
│  - CLI (arsy), SDKs (Go/TS), Code Generation, AI Assistant     │
└─────────────────────────────────────────────────────────────────┘
                              ▲
┌─────────────────────────────────────────────────────────────────┐
│  LAYER 6: Full-Stack App Builder                               │
│  - Metadata-driven UI (Forms, Lists, Kanban, Calendar)         │
│  - Studio (Admin Console), Real-time Collaboration             │
└─────────────────────────────────────────────────────────────────┘
                              ▲
┌─────────────────────────────────────────────────────────────────┐
│  LAYER 5: Manifest-Driven Module Runtime                       │
│  - Module Registry, Lifecycle (install/upgrade/rollback)       │
│  - Declarative Views, Actions, Menus, Workflows                │
│  - Seed Data, Deterministic Bootstrap                          │
└─────────────────────────────────────────────────────────────────┘
                              ▲
┌─────────────────────────────────────────────────────────────────┐
│  LAYER 4: Query Engine & API Platform                          │
│  - Validated Query AST, Dialect Compiler (SQLite/PG)           │
│  - Relation-Aware Querying, Eager Loading                      │
│  - Generated CRUD APIs + OpenAPI                               │
│  - API Keys, Quotas, Rate Limits, Webhooks                     │
└─────────────────────────────────────────────────────────────────┘
                              ▲
┌─────────────────────────────────────────────────────────────────┐
│  LAYER 3: Security & Enterprise Tenancy                        │
│  - Server-Derived Model/Field Authorization                    │
│  - Record-Level Policies, Tenant Isolation (RLS)               │
│  - Audit Trail, Secret Handling, Security Diagnostics          │
└─────────────────────────────────────────────────────────────────┘
                              ▲
┌─────────────────────────────────────────────────────────────────┐
│  LAYER 2: Database Abstraction & Migrations                    │
│  - Database Adapter Contract (SQLite, PostgreSQL)              │
│  - Tenant-Aware Provider Router                                │
│  - Deterministic Schema Planner/Executor                       │
│  - Versioned Migration Lifecycle & Recovery                    │
└─────────────────────────────────────────────────────────────────┘
                              ▲
┌─────────────────────────────────────────────────────────────────┐
│  LAYER 1: Intelligent ORM Engine (Core)                        │
│  - Canonical Model Descriptor (identity, fields, relations)    │
│  - Field Types, Defaults, Validators, Normalization            │
│  - Relation Metadata & Lifecycle (belongs-to, has-many, M2M)   │
│  - Validation/Default/Normalization Pipeline                   │
│  - Unit of Work, Lifecycle Hooks, Transaction Management       │
│  - Repository & Persistence Mapper (typed CRUD, bulk, upsert)  │
└─────────────────────────────────────────────────────────────────┘
                              ▲
┌─────────────────────────────────────────────────────────────────┐
│  LAYER 0: Infrastructure Foundation                            │
│  - Tools: Hooks, Cron, Filesystem, Mailer, Security, Tokens    │
│  - Subscriptions (Realtime), Router, Store, Logger             │
│  - SQLite (modernc.org/sqlite), PostgreSQL (pgx) Drivers       │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. Public Concepts & Terminology (A-R-S-Y Native)

| Concept | A-R-S-Y Term | Legacy PocketBase Term (INTERNAL ONLY) |
|---------|-------------|----------------------------------------|
| Application | **App** | PocketBase |
| Add-on Package | **Module** | (none) |
| Data Schema | **Model** | Collection |
| Schema Attribute | **Field** | Field |
| Cross-Model Link | **Relation** | Relation |
| Data Row | **Record** | Record |
| Multi-Tenant Partition | **Tenant** | (implicit via databases) |
| Access Control | **Role / Policy** | Rules (createRule, viewRule, etc.) |
| UI Definition | **View** | View |
| Executable Operation | **Action** | (none) |
| Business Process | **Workflow** | (none) |
| Admin User | **Owner / Admin** | Superuser |
| Setup Flow | **Bootstrap / Onboarding** | Installer |
| Data Directory | **arsy_data** | pb_data |
| Auth Collection | **_admins** | _superusers |
| Install Route | **/arsyinstall** | /pbinstall |

---

## 3. Legacy PocketBase Concepts — MUST NOT LEAK

| Legacy Concept | Containment Strategy |
|----------------|---------------------|
| `Collection` type | Internal to `core/collection_model.go`; public API uses `Model` |
| `Record` type | Internal to `core/record_model.go`; repositories return typed structs |
| `createRule`/`viewRule`/`updateRule`/`deleteRule` | Internal; exposed via `Role`/`Policy` metadata |
| `_superusers` collection | Renamed to `_admins` in all user-facing APIs |
| `pb_data` directory | Renamed to `arsy_data` in config, CLI, docs |
| `pbinstall` route | Renamed to `arsyinstall` |
| `DefaultInstallerEmail` constant | Renamed to `DefaultBootstrapEmail` |
| `PocketBase` in error messages | Replaced with `A-R-S-Y` |
| `Superuser` in CLI/UI | Replaced with `Owner` or `Admin` |
| Direct `dbx` usage | Wrapped in `core.App.DB()`; never exposed |
| Direct SQLite/PostgreSQL access | Only via Database Adapter Contract |

---

## 4. Package Ownership & Dependency Direction

```
A-R-S-Y DOMAIN (public API surface)          INFRASTRUCTURE (implementation)
─────────────────────────────────────        ──────────────────────────────
arsy/aeng/          ←──────────────────        core/ (ADAPTED)
arsy/engine/        ←──────────────────        core/ (ADAPTED)
arsy/tenancy/       ←──────────────────        core/ (ADAPTED), tools/*
apis/               ←──────────────────        core/ (ADAPTED), arsy/*
ui/                 ←──────────────────        apis/ (generated), arsy/*
cmd/                ←──────────────────        core/, arsy/*, apis/
tools/*             ←──────────────────        (no domain deps)
```

### Dependency Rules (Enforced by CI)
- **Domain packages** (`arsy/*`, `apis/`, `ui/`, `cmd/`) **MUST NOT** import concrete drivers (`modernc.org/sqlite`, `github.com/jackc/pgx`)
- **Domain packages** **MUST NOT** import `dbx` directly — use `core.App.DB()`
- **Infrastructure packages** (`core/`, `tools/*`, `migrations/`) **MAY** import drivers
- **Adapter packages** (`arsy/tenancy/pgadapter/`) implement `DatabaseAdapter` interface
- **One registry**: `arsy/engine/registry.Registry` is the single metadata source of truth

---

## 5. Versioning Policy

| Artifact | Versioning Scheme | Compatibility |
|----------|------------------|---------------|
| **Module Manifest** | SemVer (v1.0.0) | Major = breaking; Minor = additive; Patch = fix |
| **Public Go API** | Go module version (v1.x.y) | Semantic Import Versioning |
| **REST API** | URL versioning (`/api/v1/`) | Major in URL; additive within version |
| **OpenAPI Spec** | Tied to REST API version | Generated from model descriptors |
| **Database Adapters** | Tied to Go module version | Interface stability per major |
| **CLI Commands** | Tied to Go module version | Flags stable within major |
| **Database Migrations** | Timestamp + semantic name | Immutable once applied |
| **UI Assets** | Embedded in binary version | Built from source at release |

---

## 6. Extension Points & Lifecycle Contracts

### 6.1 Module Manifest Extension Points
```json
{
  "models": [...],           // Extend: add models
  "fields": [...],           // Extend: add field types
  "relations": [...],        // Extend: define relations
  "views": [...],            // Extend: declarative UI
  "actions": [...],          // Extend: executable operations
  "workflows": [...],        // Extend: business processes
  "permissions": [...],      // Extend: role/policy rules
  "seeds": [...],            // Extend: bootstrap data
  "migrations": [...]        // Extend: schema changes
}
```

### 6.2 Lifecycle Contracts
| Stage | Hook | Context | Must Not |
|-------|------|---------|----------|
| **Pre-Install** | `validateManifest` | Manifest only | Mutate state |
| **Install** | `runMigrations` | TX + Adapter | Commit outside TX |
| **Post-Install** | `runSeeds` | TX + Repository | Bypass auth |
| **Enable** | `registerModels` | Registry | Modify other modules |
| **Disable** | `unregisterModels` | Registry | Delete data |
| **Upgrade** | `runMigrations` + `runSeeds` | TX + Adapter | Downgrade schema |
| **Rollback** | `revertMigrations` | TX + Adapter | Leave partial state |
| **Uninstall** | `cleanup` | TX + Adapter | Delete other modules' data |

### 6.3 Adapter Contract (DatabaseAdapter)
```go
type DatabaseAdapter interface {
  // Connection
  Connect(config AdapterConfig) (Connection, error)
  Close() error
  Ping() error
  
  // Schema
  Introspect() (Schema, error)
  Plan(desired Schema) (MigrationPlan, error)
  Execute(plan MigrationPlan) error
  
  // Tenancy
  ProvisionTenant(id string) error
  DeprovisionTenant(id string) error
  SetTenantContext(conn Connection, tenantID string) error
  
  // Capabilities
  Capabilities() AdapterCapabilities
}
```

### 6.4 Query AST Extension
```go
type QueryAST interface {
  // Core (always supported)
  Filter(FilterExpr)
  Sort(SortExpr)
  Project(FieldList)
  Paginate(Offset, Limit)
  
  // Relations (opt-in via adapter capability)
  Expand(RelationPath)
  FilterRelation(RelationPath, FilterExpr)
  
  // Extensions (adapter-specific)
  Extension(name string, args map[string]any)
}
```

---

## 7. Architecture Decision Record

| Decision | Rationale | Evidence |
|----------|-----------|----------|
| **Manifest-first** | Deterministic, AI-draftable, policy-enforceable | ARSY-070 spec |
| **Single registry** | No competing metadata sources; validation at load | ARSY-030, ARSY-036 |
| **Adapter pattern** | Domain independent of SQLite/PG; testable | ARSY-040, ARSY-044 |
| **Tenant-aware providers** | Isolation at connection level; no cross-tenant leaks | ARSY-041, ARSY-051 |
| **Field-level auth** | Default deny; projection before serialization | ARSY-050 |
| **RLS for PG** | Database-enforced isolation; no app bypass | ARSY-044 |
| **Generated CRUD** | No hand-written routes; consistent auth/validation | ARSY-062 |
| **Declarative UI** | Metadata-driven; no client script execution | ARSY-072, ARSY-080 |
| **Circuit-breaker logging** | Audit failures never block requests | AENG-045 (implemented) |

---

## 8. Acceptance Checklist (ARSY-002)

- [x] Product layers defined (7 layers + infrastructure)
- [x] Public concepts documented with terminology mapping
- [x] Legacy concepts containment strategy defined
- [x] Package ownership and dependency direction specified
- [x] Versioning policy for all artifacts defined
- [x] Extension points and lifecycle contracts specified
- [ ] Architecture diagrams published (context, container, package)
- [ ] Public vocabulary and compatibility policy documented
- [ ] New packages can be placed unambiguously

**Evidence location:** `docs/architecture/a-r-s-y-architecture.md`