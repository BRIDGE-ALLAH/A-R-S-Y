# AENG-001: Inventory PocketBase Record and Collection Capabilities

**Status:** Draft
**Owner:** A-R-S-Y Team
**Date:** 2026-11-09

## Purpose
Comprehensive inventory of PocketBase's built-in record and collection capabilities to inform AENG module design decisions. This document captures what PocketBase provides natively vs what AENG must add.

---

## 1. Collection Types

| Type | Description | System/Created By | Notes |
|------|-------------|-------------------|-------|
| **Auth** | User accounts with built-in auth (email/password, OAuth2, JWT) | User | `core.CollectionTypeAuth` |
| **Base** | Standard data collections | User | `core.CollectionTypeBase` |
| **View** | Read-only SQL views | User | `core.CollectionTypeView` |
| **Superusers** | Admin accounts (separate from regular auth) | System | `core.CollectionNameSuperusers` |

---

## 2. Field Types (Complete List)

### Core Field Types
| Field | Go Type | PocketBase Type | Constraints | Notes |
|-------|---------|-----------------|-------------|-------|
| **Text** | `string` | `core.TextField` | `Min`, `Max`, `Pattern` | Single-line text |
| **Editor** | `string` | `core.EditorField` | `Min`, `Max` | Rich text / long text |
| **Number** | `float64` / `int64` | `core.NumberField` | `Min`, `Max`, `OnlyInt` | Integers or floats |
| **Bool** | `bool` | `core.BoolField` | — | True/false |
| **Email** | `string` | `core.EmailField` | — | Validates email format |
| **URL** | `string` | `core.URLField` | — | Validates URL format |
| **File** | `[]string` | `core.FileField` | `MaxSelect`, `MaxSize`, `MimeTypes` | File uploads |
| **Select** | `string` / `[]string` | `core.SelectField` | `Values`, `MaxSelect` | Dropdown/multi-select |
| **Relation** | `string` / `[]string` | `core.RelationField` | `CollectionId`, `MaxSelect`, `CascadeDelete` | Foreign key to another collection |
| **Date** | `string` (RFC3339) | `core.DateField` | `Min`, `Max`, `ExceptDates` | Date only (no time) |
| **DateTime** | `string` (RFC3339) | `core.AutodateField` / `core.DateField` | `Min`, `Max` | Date + time |
| **Autodate** | `string` (RFC3339) | `core.AutodateField` | `OnCreate`, `OnUpdate` | Auto-set on create/update |
| **JSON** | `map[string]any` | `core.JSONField` | `MaxSize` | Arbitrary JSON |

### Special/System Fields
| Field | Purpose |
|-------|---------|
| `id` | Auto-generated unique record ID (15-21 chars) |
| `created` | Creation timestamp (Autodate OnCreate) |
| `updated` | Last update timestamp (Autodate OnUpdate) |
| `collectionId` | Reference to owning collection |
| `collectionName` | Name of owning collection |

---

## 3. Index Types

| Index Type | SQL Equivalent | Use Case |
|------------|----------------|----------|
| Single-field | `CREATE INDEX idx_x ON t (x)` | Simple lookups |
| Composite | `CREATE INDEX idx_xy ON t (x, y)` | Multi-column queries |
| Unique | `CREATE UNIQUE INDEX idx_x ON t (x)` | Enforce uniqueness |
| Partial | `CREATE INDEX idx_x ON t (x) WHERE y = 'z'` | Conditional indexing |

**Limitation:** No full-text search indexes, no spatial/geospatial indexes, no expression indexes.

---

## 4. Query Capabilities

### Supported (via `dbx` / `core.RecordQuery`)
- **Filter:** `=`, `!=`, `>`, `>=`, `<`, `<=`, `LIKE`, `NOT LIKE`, `IN`, `NOT IN`, `~` (regex), `!~`
- **Logical:** `&&` (AND), `||` (OR), `!` (NOT)
- **Grouping:** Parentheses for precedence
- **Sorting:** `field`, `-field` (desc)
- **Pagination:** `limit`, `offset` / `page`, `perPage`
- **Expansion:** `expand=relationField` (eager load related records)

### NOT Supported Natively
- Joins across collections (must use expand or manual queries)
- Subqueries
- Aggregations (COUNT, SUM, AVG, etc.) - must use raw SQL
- Window functions
- CTEs (Common Table Expressions) in filter syntax
- Full-text search (no `tsvector`/`tsquery`)

---

## 5. Record Operations

| Operation | Method | Hooks Fired |
|-----------|--------|-------------|
| Create | `app.Save(record)` | `OnRecordCreate`, `OnRecordAfterCreateSuccess` |
| Update | `app.Save(record)` | `OnRecordUpdate`, `OnRecordAfterUpdateSuccess` |
| Delete | `app.Delete(record)` | `OnRecordDelete`, `OnRecordAfterDeleteSuccess` |
| Bulk | `app.DeleteRecords(collection, filter)` | Per-record hooks |

**Validation:** Runs on `Save()` - field constraints, required fields, pattern matching, unique indexes.

---

## 6. Authentication & Authorization

### Built-in Auth
- **Email/password** with bcrypt
- **OAuth2 providers:** Google, GitHub, GitLab, Microsoft, Apple, Discord, etc.
- **JWT tokens** with configurable expiry
- **Password reset** via email
- **Email verification** flow
- **2FA (TOTP)** support

### Authorization Model
| Level | Mechanism |
|-------|-----------|
| **Superuser** | Full admin access, bypasses all rules |
| **Collection-level** | View/Manage permissions per collection per role |
| **Record-level** | `@request.auth.id = owner` style filters in API rules |
| **API Rules** | `listRule`, `viewRule`, `createRule`, `updateRule`, `deleteRule` per collection |

**Limitation:** No built-in RBAC roles (admin/developer/viewer) - must implement via custom fields + API rules.

---

## 7. File Storage

| Feature | Support |
|---------|---------|
| Local filesystem | ✅ Default |
| S3-compatible | ✅ Via `filesystem.S3` |
| Max file size | Configurable per field |
| MIME type restriction | Per field |
| File versioning | ❌ Not built-in |
| CDN integration | Manual (serve via reverse proxy) |

---

## 8. Realtime / Subscriptions

| Feature | Support |
|---------|---------|
| Realtime subscribe | ✅ `pb.realtime.subscribe(collection, callback)` |
| Presence | ✅ `pb.realtime.presence` |
| Filter subscriptions | ✅ Topic-based with filter syntax |
| Server-sent events | ✅ Native |

---

## 9. Extensibility Points

| Extension Point | Description |
|-----------------|-------------|
| **Hooks** | Go callbacks on record lifecycle, API requests, auth events |
| **Custom Routes** | `app.OnServe().BindFunc(e => e.Router.GET(...))` |
| **Migrations** | Versioned schema changes via `core.SystemMigrations` |
| **Settings** | Key-value store via `app.Settings()` |
| **Mailer** | Custom SMTP / SendGrid / custom transport |
| **FileSystem** | Custom storage backends |

---

## 10. Gaps for AENG (What We Must Build)

| Capability | PocketBase | AENG Must Provide |
|------------|------------|-------------------|
| Multi-tenant isolation | ❌ | Tenant-scoped collections, routing |
| Schema versioning per module | ❌ | Migration executor per add-on |
| Model metadata (relations, views, actions) | ❌ | Registry + manifest |
| Seed data management | ❌ | Idempotent seed runner |
| Security metadata enforcement | ❌ | Runtime policy engine |
| View/Action manifest | ❌ | Metadata-driven UI contracts |
| Static Go model mapping | ❌ | Codegen / reflection mapper |

---

## 11. References

- [PocketBase Go SDK](https://github.com/pocketbase/pocketbase)
- [Core Package](https://pkg.go.dev/github.com/pocketbase/pocketbase/core)
- [Tools Router](https://pkg.go.dev/github.com/pocketbase/pocketbase/tools/router)
- [DBX Query Builder](https://github.com/go-dbx/dbx)