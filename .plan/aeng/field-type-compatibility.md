# AENG-003: PocketBase Field Type Compatibility Matrix

Baseline: pinned PocketBase field implementations (`core/*field*.go`) + SQLite storage semantics.

| AENG/business type | PocketBase native field | Native support | AENG/SQLite notes |
|---|---|---|---|
| char / short text | `TextField` | Yes | `Max`, `Pattern`, required supported. SQLite stores TEXT; length/pattern enforced by PB validation. |
| long text | `TextField` / `EditorField` | Yes | Editor stores HTML; sanitize/output-encode required at render boundaries (XSS). |
| integer | `NumberField` (no decimals) / app-level int | Partial | PB `NumberField` is numeric; strict integer/size constraints need AENG validation. SQLite INTEGER affinity; overflow/precision rules must be explicit. |
| number / decimal | `NumberField` | Partial | Floating/decimal precision limits in SQLite/JSON; money should use integer minor units or explicit decimal strategy in AENG. |
| bool | `BoolField` | Yes | Stored/serialized consistently by PB; SQLite uses integer affinity under the hood. |
| date/time | `DateField` / autodate fields | Yes | ISO-ish strings in PB APIs; timezone normalization is application responsibility. |
| JSON / structured | `JSONField` | Yes | SQLite JSON1 features vary by build; complex querying should use validated AENG query paths, not raw SQL from clients. |
| file / blob ref | `FileField` | Yes | Files stored in PB filesystem/S3 backends, not in SQLite rows; tenant file isolation is separate from DB isolation. |
| email | `EmailField` | Yes | Format validation native; uniqueness/indexing via collection indexes where needed. |
| URL | `URLField` | Yes | Format validation native. |
| select (single/multi) | `SelectField` | Yes | Values constrained to configured options; max selections for multi. |
| relation (many-to-one) | `RelationField` | Yes | Single/multi relation IDs; cascade behavior is **not** implicit business delete—AENG must define explicit delete policy (#106). |
| one-to-many / many-to-many | Relation patterns / join collections | Requires AENG code | Inverse/through-table ownership and naming must be declared in AENG relation metadata. |

## SQLite-specific limitations affecting correctness
- Dynamic typing/affinity can surprise comparisons; bind types and normalize in mappers.
- Concurrent writes limited; tenant DBs use `MaxOpenConns(1)` pragmatically in `SQLiteProvider`.
- `ALTER TABLE` capabilities are limited vs PostgreSQL; migrations must be planned (AENG-014/032).
- Foreign key enforcement depends on pragma/settings and schema; do not assume cascades.
- Large JSON/file workloads: keep files out of DB; index/query JSON cautiously.

## Focused references/examples
1. AENG field metadata/validation: `arsy/engine/fields/types.go`, `validate.go`, `validate_test.go`.
2. PB field usage in migrations/collections: `arsy/migrations/baas_tenants.go`, `mvp1.go`.
3. Tenant SQLite handle pragmas: `arsy/tenancy/sqlite_provider.go` (`SetMaxOpenConns(1)`, Ping).
4. Query safety boundary: `arsy/engine/query/query.go` + compiler vs raw SQL prohibition in API paths.
5. Relation metadata contract: `arsy/engine/relations/*` (explicit cardinality/delete intent; no implicit cascade).

No schema or runtime changes.
