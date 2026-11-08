# AENG-001: PocketBase Capability Inventory (pinned dependency)

Source baseline: `go.mod` pins `github.com/pocketbase/pocketbase` (module identity in this fork: `github.com/bridge-allah/a-r-s-y`), Go 1.24, dbx via `ganigeorgiev.com/pocketbase/dbx`.

## Capability map

| Area | PB API / source | Status | Notes for AENG |
|---|---|---|---|
| Collections (schema) | `core.Collection`, `core.App.FindCollectionByNameOrId`, `apis` collection routes; migrations under `migrations/` + `arsy/migrations/` | Reusable | Control-plane metadata; dynamic collection creation works but not tenant-isolated by default. |
| Records CRUD | `core.Record`, `core.App.Save`, `FindRecordById`, `FindRecordsByFilter`; `apis` record handlers | Reusable | Base CRUD; AENG must not replace PB Record API. Tenant isolation requires separate DB or strict scoping. |
| Fields | `core.TextField`, `NumberField`, `BoolField`, `DateField`, `JSONField`, `FileField`, `RelationField`, `SelectField`, `EmailField`, `URLField`, etc. (`core/*field*.go`) | Extendable | Field types are PB-native; AENG engine metadata (`arsy/engine/fields`) adds validation/defaults layer. |
| Query/filter | `dbx` builder, `FindRecordsByFilter(expr, sort, limit, offset, params)`; filter parser in `tools/search`/`core` | Extendable | PB filter syntax is string-based; AENG query AST (`arsy/engine/query`) provides engine-neutral, validated compilation. |
| Relations | `RelationField`, `expand` support in record APIs | Reusable with limits | Many-to-one native; many-to-many/join ownership needs AENG metadata (`arsy/engine/relations`). |
| Validation | Field validators, `core.Record` validation hooks, collection rules | Reusable | AENG adds pipeline (`arsy/engine/validation`) and typed field validation; must compose, not bypass. |
| Hooks/lifecycle | `core.App.On*` hooks (bootstrap, serve, model/record hooks), `arsy/engine/hooks` | Reusable | AENG hook registry wraps transaction-aware semantics; PB hooks remain authoritative for PB collections. |
| Migrations | `core.SystemMigrations.Register`, `migratecmd` plugin, `arsy/migrations/registry.go` | Reusable | Control-plane migrations only; tenant-local schema needs separate runner (BAAS-011/AENG-032). |
| Auth/rules | `_superusers`/auth collections, collection API rules, `apis.Require*` middlewares | Reusable | AENG security metadata (#116/#136) must enforce server-side; tenant auth via membership registry. |
| Multi-DB/tenancy | PB app owns one `*dbx.DB` per app instance (`core/base.go`/app internals) | Unsuitable as-is | Dedicated per-tenant SQLite achieved via separate `sql.DB`/dbx handles (`arsy/tenancy/sqlite_provider.go`), not multiple full PB app instances sharing process state unsafely. |

## Verified examples (source-linked)

1. **Record CRUD**: `core.Record` + `App.Save/FindRecordById/Delete` used throughout `core/*_test.go` and A-R-S-Y HTTP handlers (`arsy/httpapi/*`). Pattern: find collection -> `core.NewRecord(col)` -> `Set` fields -> `app.Save(record)`.
2. **Relation/query**: `App.FindRecordsByFilter(collection, filter, sort, limit, offset, params)` used in `arsy/httpapi/tenants_admin.go`, `baas_memberships` lookups; relation expansion via PB record APIs where configured.
3. **Migrations**: `core.SystemMigrations.Register(up, down, "timestamp_name.go")` in `arsy/migrations/baas_tenants.go`, `baas_memberships.go`.

## Gaps for a general-purpose application engine

- No built-in per-tenant database isolation (shared app DB by default).
- String filter syntax lacks typed AST/field allowlisting guarantees needed for safe dynamic queries.
- No first-class static Go model ORM mapping to PB/dbx with typed field-addressable errors (AENG-030 addresses partially).
- No module/add-on lifecycle (install/enable/disable/upgrade) with dependency resolution at runtime.
- No declarative view/action metadata validated against runtime registry.
- Schema migrations are global/system-oriented; tenant-local, versioned, checksum-tracked migrations need dedicated executor.
- Record rules and multi-company scope are distinct concerns from tenant isolation; require explicit enforcement layer.

## Constraints verified
- Documentation-only; no production code changes.
- Facts derived from pinned module usage and repository sources (`arsy/*`, `core/*`, `migrations/*`). Unverified upstream internals are not asserted.
