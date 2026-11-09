# AENG-003: Document PocketBase Schema and Field Type Limitations

**Status:** Draft
**Owner:** A-R-S-Y Team
**Date:** 2026-11-09

## Purpose
Document known limitations of PocketBase's schema system and field types to inform AENG design decisions and workarounds.

---

## 1. Field Type Limitations

### 1.1 TextField
| Limitation | Impact | Workaround |
|------------|--------|------------|
| Max 65535 chars (SQLite TEXT) | Large content needs EditorField | Use EditorField for long text |
| Pattern validation only regex | Complex validation not possible | Hook validation |
| No full-text search | Cannot search within text | External search (Meilisearch, Typesense) |

### 1.2 EditorField
| Limitation | Impact | Workaround |
|------------|--------|------------|
| Stored as TEXT | Same size limit as TextField | Acceptable for most content |
| No built-in sanitization | XSS risk if rendered raw | Sanitize on output |

### 1.3 NumberField
| Limitation | Impact | Workaround |
|------------|--------|------------|
| Float64 internally | Precision issues for financial | Use `OnlyInt: true` for integers, or store as string |
| No decimal precision control | Cannot enforce 2 decimal places | Application-level validation |

### 1.4 BoolField
| Limitation | Impact | Workaround |
|------------|--------|------------|
| Only true/false | No nullable/three-state | Use SelectField with "unknown" option |

### 1.5 SelectField
| Limitation | Impact | Workaround |
|------------|--------|------------|
| Values fixed at schema time | Dynamic options not supported | RelationField to options collection |
| MaxSelect only for multi | Single select enforced by UI | Application logic |

### 1.6 RelationField
| Limitation | Impact | Workaround |
|------------|--------|------------|
| Single target collection | Polymorphic relations not supported | Separate relation fields per type |
| CascadeDelete all-or-nothing | Cannot cascade some, restrict others | Application-level cascade logic |
| No self-referential depth limit | Recursive relations can cause issues | Application validation |

### 1.7 DateField / DateTimeField
| Limitation | Impact | Workaround |
|------------|--------|------------|
| RFC3339 string storage | No native date arithmetic in filters | Application-level date math |
| Timezone handling | Stored as UTC, displayed per client | Store timezone separately if needed |

### 1.8 JSONField
| Limitation | Impact | Workaround |
|------------|--------|------------|
| No schema validation | Arbitrary JSON structure | Application-level JSON Schema validation |
| No partial updates | Full replacement only | Application merges |
| No indexing within JSON | Cannot query JSON keys efficiently | Extract to separate fields |

### 1.9 FileField
| Limitation | Impact | Workaround |
|------------|--------|------------|
| No versioning | Overwrites on update | Manual versioning with separate records |
| No thumbnail generation | Must generate externally | Hook on upload |
| Max file size per field | Cannot vary per record | Set high limit, validate in hook |

---

## 2. Schema Limitations

### 2.1 Collection-Level
| Limitation | Impact | Workaround |
|------------|--------|------------|
| No table partitioning | Large collections slow | Archive old records to separate collection |
| No materialized views | Computed aggregates not stored | Cron job to update summary collection |
| No foreign key enforcement | SQLite FK not enforced by default | Enable `PRAGMA foreign_keys = ON` + hooks |

### 2.2 Index Limitations
| Limitation | Impact | Workaround |
|------------|--------|------------|
| No partial indexes in UI | Must use raw SQL in migration | Create in migration `up` function |
| No expression indexes | Cannot index computed values | Computed field + trigger |
| No full-text indexes | No native search | External search engine |

### 2.3 Migration Limitations
| Limitation | Impact | Workaround |
|------------|--------|------------|
| No data migration helpers | Manual SQL for data transforms | Write custom migration SQL |
| No rollback for data changes | Down migration only drops schema | Backup before migration |
| Single transaction per migration | Large migrations may lock DB | Batch in multiple migrations |

---

## 3. Query/Filter Limitations

| Limitation | Impact | Workaround |
|------------|--------|------------|
| No JOIN in filter syntax | Cannot filter by related fields directly | Use `expand` + post-filter, or raw SQL |
| No subqueries | Complex filtering limited | Multiple queries + application logic |
| No aggregations | COUNT, SUM, AVG not in filter | Raw SQL or separate analytics DB |
| No window functions | Ranking, running totals not possible | Application computation |
| Regex (`~`) slow on large data | Performance issues | Limit dataset first, or use external search |

---

## 4. Authorization Limitations

| Limitation | Impact | Workaround |
|------------|--------|------------|
| No built-in RBAC | Roles must be custom | Membership collection + API rules |
| API rules are string filters | Complex logic hard to express | Hook-based authorization |
| No field-level permissions | All-or-nothing per record | Separate collections per sensitivity |
| No column-level encryption | Sensitive fields visible to admins | Application-level encryption |

---

## 5. Realtime Limitations

| Limitation | Impact | Workaround |
|------------|--------|------------|
| No presence persistence | Presence lost on reconnect | Store presence in collection |
| No message history | New subscribers miss past events | Event sourcing pattern |
| Filter syntax same as API | Same query limitations | Pre-compute subscriptions |

---

## 5. File Storage Limitations

| Limitation | Impact | Workaround |
|------------|--------|------------|
| Single filesystem per app | Cannot separate by tenant | Prefix paths, or custom FS per tenant |
| No CDN integration | Static serving only | Reverse proxy with cache headers |
| No image processing | No resize/convert on upload | Hook + external processor |

---

## 6. AENG Design Implications

Based on these limitations, AENG must provide:

| AENG Feature | Addresses Limitation |
|--------------|---------------------|
| **Module Manifest** | Schema versioning, declarative fields |
| **Migration Executor** | Data migration helpers, rollback safety |
| **Tenant Routing** | Multi-database isolation (not single SQLite) |
| **Security Metadata** | Field-level permissions, RBAC |
| **View/Action Manifest** | Computed fields, derived data |
| **Seed Data Runner** | Idempotent data initialization |
| **Static Model Mapper** | Type-safe Go models from schema |

---

## 7. References

- [PocketBase Schema](https://pocketbase.io/docs/schema/)
- [Field Types Reference](https://pocketbase.io/docs/fields-overview/)
- [Filter Syntax](https://pocketbase.io/docs/filtering/)
- [Migrations](https://pocketbase.io/docs/go-migrations/)