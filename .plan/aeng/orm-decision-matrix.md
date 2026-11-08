# AENG-006: ORM Architecture Decision Matrix

## Options compared

| Criterion | A) PB Record/Collection dynamic API | B) dbx + thin static mapper (current direction) | C) Metadata-driven dynamic ORM |
|---|---|---|---|
| Type safety | Low at compile time (dynamic fields) | **High** for static Go models | Medium (runtime schema/types) |
| Dynamic models at runtime | **Strong** (collections/records) | Weak without codegen/generation | **Strong** if metadata+schema executor mature |
| Relations | PB relations/expand | Explicit mapping; join/preload patterns manual | Possible but complex/costly |
| Transactions | PB/app/dbx transaction support | dbx/UoW (`arsy/engine/uow`) | Same foundation required |
| Migrations | PB system migrations; dynamic schema ops | Explicit SQL/migration plans | Needs schema diff/executor (high risk) |
| SQLite fit | Native | **Good** (dbx, explicit SQL) | Feasible but migration complexity |
| PostgreSQL future | PB support differs; adapter work needed | Possible via dialect seams (`arsy/engine/query/compiler`, `persist`) | Harder; dual-dialect dynamic DDL |
| Maintenance cost | Low if staying within PB | **Moderate** (explicit code) | High (correctness, security, migrations) |
| Security surface | PB rules help; still need tenant scoping | Smaller if queries validated/allowlisted | Larger (dynamic SQL/schema paths) |

## Verified facts
- PB Record/dbx are available and used in-repo (`core`, `arsy/httpapi`, `arsy/engine/persist/mapper.go` uses dbx + `modernc.org/sqlite` patterns).
- AENG already has static-oriented building blocks: registry, fields, relations metadata, query AST/compiler, hooks, UoW, validation, persist store.
- Tenant isolation is implemented at DB-handle level (`arsy/tenancy`), not by PB multi-app tenancy.

## Assumptions (explicit)
- Future PostgreSQL support will require deliberate dialect abstraction; no driver added by default.
- Full dynamic business models (user-defined tables at runtime) are not required for initial production value.

## Recommendation (staged)
1. **Now**: continue **B + selective A** — reuse PB for control-plane/auth/admin where safe; use thin static mapper + validated query/compiler for engine CRUD; keep PB Record APIs unchanged.
2. **Next**: complete migration executor, seed runner, runtime security enforcement, view metadata validation using explicit contracts.
3. **Later (only with evidence)**: evaluate limited metadata-driven model generation behind strict schema migration/review gates; do not add GORM/Bun without measured need.
4. **PostgreSQL**: optional adapter spike behind datastore/compiler contracts; SQLite-only remains default.

## Unresolved decisions
- Canonical decimal/money representation across SQLite/PostgreSQL.
- Tenant-local migration format + executor safety gates (AENG-014/032).
- How much dynamic schema creation (if any) is exposed to modules/operators.
- Unified authorization composition: PB rules vs AENG security metadata vs tenant membership.

Documentation-only.
