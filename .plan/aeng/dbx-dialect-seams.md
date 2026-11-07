# dbx Query and SQL Dialect Extension Points

**Issue**: #101 `[AENG-004] Map dbx query and SQL dialect extension points`  
**Created**: 2026-11-07  
**Scope**: Evidence-based mapping of `ganigeorgiev.com/pocketbase/dbx` v1.13.0 abstraction boundaries for future SQLite/PostgreSQL support without library replacement.

---

## 1. dbx.Builder Interface (Dialect Entry Point)

**Source**: `ganigeorgiev.com/pocketbase/dbx` (vendor or go.mod dependency)  
**Repository usage**: `core/db_builder.go`, `core/db.go`, `core/db_connect.go`, `core/app.go`

### Key Methods
| Method | Dialect-Dependent | Notes |
|---|---|---|
| `GeneratePlaceholder(int) string` | ✅ YES | PostgreSQL `$1`, SQLite `?` (positional) — core abstraction |
| `Quote(string) string` | ✅ YES | Identifier quoting: PostgreSQL `"name"`, MySQL `` `name` ``, SQLite `"name"` or `[name]` |
| `QuoteSimpleTableName(string) string` | ✅ YES | Safe table-name quoting |
| `QuoteSimpleColumnName(string) string` | ✅ YES | Safe column-name quoting |
| `Select(cols ...string) *SelectQuery` | ⚠️ PARTIAL | Query builder entry; row iteration/locking/pagination may differ |
| `Insert(table, cols) *Query` | ⚠️ PARTIAL | `RETURNING` support varies (PostgreSQL yes, SQLite 3.35+) |
| `Update(table, cols, where) *Query` | ⚠️ PARTIAL | Same as Insert for `RETURNING` |
| `Delete(table, where) *Query` | ⚠️ PARTIAL | Same |
| `NewQuery(sql string) *Query` | ❌ NO | Raw SQL entry; dialect responsibility on caller |
| `Model(interface{}) *ModelQuery` | ⚠️ PARTIAL | Reflection-based CRUD; `RETURNING` and type mapping |
| `Transactional(func(*Tx) error) error` | ❌ NO | Driver-level transaction, dbx wraps |

### Repository Wrapper: `dualDBBuilder` (`core/db_builder.go:14`)
Routes SELECT/WITH → `concurrentDB` (read replicas safe), writes → `nonconcurrentDB` (primary). Preserves single dialect per app instance.

---

## 2. Placeholder Generation

**Implementation**: `dbx.Builder.GeneratePlaceholder(i int) string`  
**Observed usage**: `core/db_builder.go:30` delegates to `concurrentDB`.

### Dialect Differences (Must Not Hide)
| Dialect | Placeholder Style | Example |
|---|---|---|
| **PostgreSQL** | Numbered `$n` | `SELECT * FROM users WHERE id = $1 AND age > $2` |
| **SQLite** | Positional `?` or numbered `?NNN` | `SELECT * FROM users WHERE id = ? AND age > ?` |
| **MySQL** | Positional `?` | Same as SQLite |

**Repository constraint**: `arsy/engine/query/compiler` (issue #132) uses `$n` style explicitly; future PostgreSQL adapter can reuse; SQLite adapter must translate or dbx handles.

**Recommendation**: Preserve dbx placeholder abstraction. Do NOT add a secondary translation layer in `arsy/engine/query/compiler`.

---

## 3. Transaction Handling

**Source**: `core/db_tx.go:32` — `dbx.DB.Transactional(func(*dbx.Tx) error) error`  
**Mechanism**: Wraps `database/sql` transactions; driver-agnostic at dbx level.

### Dialect Differences (Isolation/Locking)
| Feature | SQLite | PostgreSQL |
|---|---|---|
| Default isolation | `SERIALIZABLE` (effectively) | `READ COMMITTED` |
| Row-level locking | ❌ No (`BEGIN IMMEDIATE/EXCLUSIVE` for write) | ✅ Yes (`SELECT ... FOR UPDATE`) |
| Concurrent reads | ✅ Yes (WAL mode) | ✅ Yes |
| `SAVEPOINT` support | ✅ Yes | ✅ Yes |

**Repository usage**: `core/base.go` uses `PRAGMA journal_mode(WAL)` for SQLite. PostgreSQL adapter must use connection-pool config, NOT pragma.

**Incompatibility**: SQLite `BEGIN IMMEDIATE` for write-ahead intent has no PostgreSQL equivalent; PostgreSQL uses row locks instead. Do NOT hide this in an abstraction.

---

## 4. JSON Operations

**SQLite**: `json_extract(column, '$.path')`, `json_array_length()`, `json_type()`  
**PostgreSQL**: `column->>'path'`, `jsonb_array_length()`, `jsonb_typeof()`

**Repository usage**: Not heavily used in PocketBase core for filter queries (uses Go unmarshaling). If `arsy/engine` adds JSON path filtering, dialect-specific SQL generation required.

**Recommendation**: Keep JSON query generation out of `arsy/engine/query/compiler` (issue #132) until explicit requirement with dialect detection.

---

## 5. Migrations (`core/migrations_runner.go`)

**Current**: PocketBase uses `dbx.Query` for DDL; SQLite-specific `CREATE TABLE` syntax.  
**Example**: `core/db_builder.go:75` — `CreateTable(table, cols map[string]string, options)`

### Dialect Differences (DDL)
| Feature | SQLite | PostgreSQL |
|---|---|---|
| `ALTER TABLE DROP COLUMN` | ❌ No (requires table rebuild) | ✅ Yes |
| `CREATE INDEX CONCURRENTLY` | ❌ No | ✅ Yes |
| `SERIAL` / `AUTOINCREMENT` | `INTEGER PRIMARY KEY AUTOINCREMENT` | `SERIAL` or `GENERATED ALWAYS AS IDENTITY` |
| `IF NOT EXISTS` | ✅ Yes | ✅ Yes |

**Repository constraint (issue #133)**: Migration executor must NOT auto-translate SQLite DDL to PostgreSQL. Require explicit per-dialect migration files or feature detection.

---
## 6. Pagination and Sorting

**dbx abstraction**: `SelectQuery.Limit(n).Offset(m).OrderBy(col)`  
**Dialect-independent**: Both SQLite and PostgreSQL support `LIMIT`/`OFFSET`.

**Performance difference**:
- PostgreSQL: efficient with indexes
- SQLite: full table scan for large offsets without covering index

**Recommendation**: No abstraction needed; document performance trade-offs in application layer.

---

## 7. Error Handling

**dbx**: Returns `database/sql` driver errors wrapped.  
**Dialect-specific error codes**:
- SQLite: `SQLITE_CONSTRAINT` (19), `SQLITE_BUSY` (5)
- PostgreSQL: `23505` (unique violation), `40001` (serialization failure)

**Repository usage**: `core/db_retry.go:11` — retry on lock/busy errors (SQLite-specific `SQLITE_BUSY`).

**Incompatibility**: PostgreSQL serialization failures need `40001`/`40P01` detection. Do NOT unify error codes; provide dialect-aware retry logic.

---

## 8. PocketBase-Specific APIs (Not in dbx)

**Source**: `core/record_query.go`, `core/collection.go`  
**PocketBase abstractions**:
- `Collection` (schema metadata)
- `Record` (dynamic field access via `Get()`/`Set()`)
- `RecordQuery()` with expand/filter DSL

**Recommendation**: These are application-level abstractions above dbx. Future PostgreSQL adapter reuses `Collection`/`Record` unchanged; only `RecordQuery` SQL generation needs dialect fork.

---

## 9. Recommended Abstraction Boundaries (No Library Replacement)

### ✅ Keep Using dbx For
1. **Placeholder generation** (`GeneratePlaceholder`)
2. **Identifier quoting** (`Quote`, `QuoteSimpleTableName`)
3. **Query builders** (`Select`, `Insert`, `Update`, `Delete`)
4. **Transaction wrapping** (`Transactional`)
5. **Model CRUD** (`Model().Insert()`, etc.)

### ⚠️ Dialect-Specific Extensions (Do NOT Hide)
1. **JSON path queries** — separate SQLite/PostgreSQL implementations
2. **DDL migrations** — per-dialect SQL files or feature detection
3. **Locking semantics** — document SQLite `BEGIN IMMEDIATE` vs PostgreSQL `FOR UPDATE`
4. **Error retry logic** — dialect-aware error code detection
5. **`RETURNING` clause** — feature-detect or require PostgreSQL 9.5+/SQLite 3.35+

### ❌ Do NOT Add
1. **Custom query builder replacing dbx** — unnecessary; dbx supports both dialects
2. **Unified error abstraction hiding codes** — breaks diagnostic clarity
3. **Automatic DDL translation** — unsafe; require explicit per-dialect migrations

---

## 10. Future PostgreSQL Adapter Checklist (Issue #96 Context)

When implementing optional PostgreSQL support:

- [ ] Use `github.com/lib/pq` or `pgx` driver; register with `database/sql`
- [ ] Provide PostgreSQL-specific `DefaultDBConnect` (separate build tag or config)
- [ ] Keep dbx as query builder; verify `GeneratePlaceholder` returns `$n`
- [ ] Add PostgreSQL-specific retry logic for serialization failures (`40001`, `40P01`)
- [ ] Document SQLite→PostgreSQL migration path (schema export/import)
- [ ] Preserve PocketBase `Collection`/`Record` APIs unchanged
- [ ] Test `arsy/engine/query/compiler` with PostgreSQL (already uses `$n` placeholders)
- [ ] Do NOT auto-migrate SQLite DDL; require manual PostgreSQL schema definition

---

## 11. Exact Source Paths (Repository Evidence)

| Component | File | Line Range | Notes |
|---|---|---|---|
| dbx.Builder wrapper | `core/db_builder.go` | 14–180 | `dualDBBuilder` routes reads/writes |
| Placeholder delegation | `core/db_builder.go` | 30–31 | Calls `concurrentDB.GeneratePlaceholder(i)` |
| Quote delegation | `core/db_builder.go` | 35–46 | `Quote`, `QuoteSimpleTableName`, `QuoteSimpleColumnName` |
| SQLite connection | `core/db_connect.go` | 6–21 | `DefaultDBConnect` with WAL pragma |
| Transaction wrapper | `core/db_tx.go` | 18–45 | `RunInTransaction`, `AuxRunInTransaction` |
| ModelQuery usage | `core/db.go` | 81–88 | `ModelQuery(m Model) *dbx.SelectQuery` |
| Retry logic (SQLite) | `core/db_retry.go` | 11–end | `baseLockRetry`, `execLockRetry` for `SQLITE_BUSY` |
| DDL builders | `core/db_builder.go` | 75–145 | `CreateTable`, `AddColumn`, `CreateIndex`, etc. |

---

## 12. Constraints Summary (Per Issue #101)

✅ **Deliverable met**: Exact source paths, API names, dialect-dependent vs independent behavior, SQLite/PostgreSQL incompatibilities documented.  
✅ **No production code changes**: This is evidence-only documentation.  
✅ **No library replacement recommended**: dbx is sufficient; add dialect detection/config only.  
✅ **Incompatibilities NOT hidden**: Transaction isolation, JSON ops, DDL, error codes require explicit handling.

---

**End of dbx-dialect-seams.md**
