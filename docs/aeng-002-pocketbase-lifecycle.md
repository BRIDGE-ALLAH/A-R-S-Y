# AENG-002: Map PocketBase App Lifecycle and Database Ownership

**Status:** Draft
**Owner:** A-R-S-Y Team
**Date:** 2026-11-09

## Purpose
Document the complete lifecycle of a PocketBase application and database ownership model to define integration seams for AENG modules.

---

## 1. Application Lifecycle

### 1.1 Startup Sequence
```
1. main() → pocketbase.New()
2. Load settings from pb_data/data.db (settings table)
3. Initialize:
   - FileSystem (local/S3)
   - Mailer (SMTP/custom)
   - Token/Signing keys
   - Admin UI assets (embedded)
4. Run migrations (core.SystemMigrations)
5. Bootstrap superuser if none exists
6. Start HTTP server (app.OnServe())
7. Register custom routes/hooks
8. Serve requests
```

### 1.2 Shutdown Sequence
```
1. Signal received (SIGTERM/SIGINT)
2. app.OnTerminate() hooks fire
3. Close database connections
4. Flush pending writes
5. Exit
```

### 1.3 Key Lifecycle Events (Hooks)
| Event | Timing | Use Case |
|-------|--------|----------|
| `OnBootstrap()` | Before first superuser | Initialize default data |
| `OnServe()` | HTTP server starting | Register routes, middleware |
| `OnTerminate()` | Graceful shutdown | Cleanup, close connections |
| `OnRecordCreate/Update/Delete` | Per record operation | Business logic, validation |
| `OnRecordAfter*Success` | After commit | Side effects (emails, webhooks) |
| `OnRecordsImport` | Bulk import | Data transformation |

---

## 2. Database Ownership Model

### 2.1 Single Database File
- PocketBase uses **one SQLite database** (`pb_data/data.db`) for everything:
  - System collections (superusers, settings, migrations)
  - User collections (base, auth, view)
  - Record data
  - File metadata (file records point to filesystem)

### 2.2 Schema Storage
- Collection definitions stored in `collections` table (JSON blobs)
- Migrations tracked in `migrations` table
- Settings in `settings` table (key-value)

### 2.3 No Multi-Database Support
- PocketBase **does not** support multiple database files per app instance
- All collections share the same SQLite connection pool
- AENG multi-tenancy must be implemented at **application layer**, not DB layer

---

## 3. Collection Lifecycle

### 3.1 Creation
```
User: app.Save(collection) OR SystemMigrations
  → Validates fields, indexes, rules
  → Creates table in SQLite
  → Registers in collections table
  → Fires OnCollectionCreate hooks
```

### 3.2 Modification
```
User: app.Save(collection) with changes
  → ALTER TABLE for new fields
  → DROP/CREATE for modified fields (data loss risk!)
  → Rebuild indexes
  → Updates collections table
  → Fires OnCollectionUpdate hooks
```

### 3.3 Deletion
```
User: app.Delete(collection)
  → DROP TABLE
  → Removes from collections table
  → Fires OnCollectionDelete hooks
```

**Critical:** Field type changes may require data migration. PocketBase does not auto-migrate data on type changes.

---

## 4. Record Lifecycle

### 4.1 Create
```
app.Save(newRecord)
  → Validate required fields, constraints, patterns
  → Generate ID (if empty)
  → Set created/updated timestamps
  → Fire OnRecordBeforeCreate hooks
  → INSERT INTO table
  → Fire OnRecordAfterCreateSuccess hooks
```

### 4.2 Update
```
app.Save(existingRecord)
  → Validate changed fields
  → Update updated timestamp
  → Fire OnRecordBeforeUpdate hooks
  → UPDATE table
  → Fire OnRecordAfterUpdateSuccess hooks
```

### 4.3 Delete
```
app.Delete(record)
  → Fire OnRecordBeforeDelete hooks
  → DELETE FROM table
  → Fire OnRecordAfterDeleteSuccess hooks
  → Cascade delete related records (if RelationField.CascadeDelete)
```

---

## 5. Migration System

### 5.1 Core Migrations (SystemMigrations)
```go
core.SystemMigrations.Register(upFn, downFn, "version_description.go")
```
- Runs at startup before serving
- Tracks applied migrations in `migrations` table
- Version = timestamp + description (e.g., `20261107010000_baas_tenants.go`)

### 5.2 AENG Module Migrations
Each add-on module should register its own migrations:
- Separate version namespace per module
- Up/Down functions for schema changes
- Idempotent (safe to re-run)

---

## 6. Integration Seams for AENG

| Seam | Mechanism | AENG Usage |
|------|-----------|------------|
| **App Bootstrap** | `OnBootstrap()` | Register module collections, seed data |
| **HTTP Server** | `OnServe()` | Register module routes, middleware |
| **Record Ops** | `OnRecord*` hooks | Enforce security, validate relations |
| **Migrations** | `SystemMigrations` | Module schema versioning |
| **Settings** | `app.Settings()` | Module configuration |
| **FileSystem** | Custom FS | Module asset storage |

---

## 7. Database Connection Management

| Aspect | PocketBase | AENG Consideration |
|--------|------------|-------------------|
| Pool size | Single connection (SQLite) | Tenant DBs need separate pools |
| WAL mode | Enabled by default | Required for tenant DBs |
| Max connections | 1 (SQLite limitation) | Per-tenant connection pools |
| Transactions | Manual `app.RunInTransaction()` | Module operations need tx support |

---

## 8. References
- [PocketBase App Lifecycle](https://pocketbase.io/docs/go-app-lifecycle/)
- [Database Migrations](https://pocketbase.io/docs/go-migrations/)
- [Record Hooks](https://pocketbase.io/docs/go-record-hooks/)