# PocketBase Transformation Inventory

**Generated:** 2026-11-09
**Task:** ARSY-001 — Establish the PocketBase transformation inventory
**Status:** In Progress

---

## 1. Root Packages and Entry Points

| Path | Classification | Description |
|------|---------------|-------------|
| `pocketbase.go` | **ADAPT** | Main A-R-S-Y app launcher (renamed from PocketBase), implements `core.App` |
| `cmd/serve.go` | **ADAPT** | `serve` command - starts web server with A-R-S-Y worker |
| `cmd/superuser.go` | **ADAPT** | `superuser` command - manage superuser accounts |
| `go.mod` | **KEEP-INTERNAL** | Module: `github.com/bridge-allah/a-r-s-y` (already renamed) |
| `Makefile` | **KEEP-INTERNAL** | Build automation |

---

## 2. Core Package (`core/`) — 129 files

### 2.1 App Interface & Bootstrap
| File | Classification | Notes |
|------|---------------|-------|
| `app.go` | **ADAPT** | Main `App` interface (71K lines) - PocketBase terminology throughout |
| `base.go` | **ADAPT** | `BaseApp` implementation (58K lines) |
| `base_test.go` | **KEEP-INTERNAL** | Tests |

### 2.2 Database Layer
| File | Classification | Notes |
|------|---------------|-------|
| `db.go` | **KEEP-INTERNAL** | DB abstraction (uses `dbx`) |
| `db_builder.go` | **KEEP-INTERNAL** | Query builder |
| `db_connect.go` | **KEEP-INTERNAL** | SQLite connection |
| `db_connect_nodefaultdriver.go` | **KEEP-INTERNAL** | No-default driver variant |
| `db_model.go` | **KEEP-INTERNAL** | Base model |
| `db_table.go` | **KEEP-INTERNAL** | Table operations |
| `db_tx.go` | **KEEP-INTERNAL** | Transaction wrapper |
| `db_retry.go` | **KEEP-INTERNAL** | Retry logic |

### 2.3 Collections & Models
| File | Classification | Notes |
|------|---------------|-------|
| `collection_model.go` | **ADAPT** | Collection definition (30K) - PocketBase terminology |
| `collection_model_auth_options.go` | **ADAPT** | Auth collection options |
| `collection_model_view_options.go` | **ADAPT** | View collection options |
| `collection_validate.go` | **ADAPT** | Validation logic |
| `record_model.go` | **ADAPT** | Record operations (42K) |
| `record_model_auth.go` | **ADAPT** | Auth record methods |
| `record_model_superusers.go` | **ADAPT** | Superuser record handling |

### 2.4 Fields (22 files)
| File | Classification |
|------|---------------|
| `field.go` | **KEEP-INTERNAL** |
| `field_autodate.go` | **KEEP-INTERNAL** |
| `field_bool.go` | **KEEP-INTERNAL** |
| `field_date.go` | **KEEP-INTERNAL** |
| `field_email.go` | **KEEP-INTERNAL** |
| `field_file.go` | **KEEP-INTERNAL** |
| `field_geo_point.go` | **KEEP-INTERNAL** |
| `field_json.go` | **KEEP-INTERNAL** |
| `field_number.go` | **KEEP-INTERNAL** |
| `field_password.go` | **KEEP-INTERNAL** |
| `field_relation.go` | **KEEP-INTERNAL** |
| `field_select.go` | **KEEP-INTERNAL** |
| `field_text.go` | **KEEP-INTERNAL** |
| `field_url.go` | **KEEP-INTERNAL** |

### 2.5 Auth & Security
| File | Classification |
|------|---------------|
| `auth_origin_model.go` | **ADAPT** |
| `auth_origin_query.go` | **ADAPT** |
| `external_auth_model.go` | **ADAPT** |
| `external_auth_query.go` | **ADAPT** |
| `mfa_model.go` | **KEEP-INTERNAL** |
| `otp_model.go` | **KEEP-INTERNAL** |
| `record_tokens.go` | **KEEP-INTERNAL** |

### 2.6 Migrations
| File | Classification |
|------|---------------|
| `migrations_list.go` | **ADAPT** - System migrations list |
| `migrations_runner.go` | **ADAPT** - Migration execution |

### 2.7 Other Core
| File | Classification |
|------|---------------|
| `events.go` | **KEEP-INTERNAL** |
| `settings_model.go` | **ADAPT** - App settings |
| `settings_query.go` | **ADAPT** |
| `view.go` | **ADAPT** - View definitions |
| `log_model.go` | **KEEP-INTERNAL** |
| `backup.go` | **ADAPT** |
| `notify_watcher.go` | **KEEP-INTERNAL** |

---

## 3. APIs Package (`apis/`) — 78 files

### 3.1 HTTP Handlers
| File | Classification | Notes |
|------|---------------|-------|
| `serve.go` | **ADAPT** | Main HTTP server |
| `base.go` | **ADAPT** | Handler/middleware wrapping |
| `middlewares.go` | **ADAPT** | Auth, rate limit, CORS, gzip |
| `health.go` | **KEEP-INTERNAL** | Health endpoint |

### 3.2 Collection/Record APIs
| File | Classification |
|------|---------------|
| `collection.go` | **ADAPT** |
| `collection_import.go` | **ADAPT** |
| `record_crud.go` | **ADAPT** (22K) |
| `record_helpers.go` | **ADAPT** |
| `batch.go` | **ADAPT** |

### 3.3 Auth APIs (20+ files)
| File | Classification |
|------|---------------|
| `record_auth.go` | **ADAPT** |
| `record_auth_with_password.go` | **ADAPT** |
| `record_auth_with_oauth2.go` | **ADAPT** |
| `record_auth_with_otp.go` | **ADAPT** |
| `record_auth_verification_*.go` | **ADAPT** |
| `record_auth_password_reset_*.go` | **ADAPT** |
| `record_auth_email_change_*.go` | **ADAPT** |
| `record_auth_impersonate.go` | **ADAPT** |
| `record_auth_refresh.go` | **ADAPT** |

### 3.4 Other APIs
| File | Classification |
|------|---------------|
| `installer.go` | **ADAPT** - Uses PocketBase terminology |
| `backup.go` | **ADAPT** |
| `file.go` | **ADAPT** |
| `realtime.go` | **ADAPT** |
| `logs.go` | **ADAPT** |
| `settings.go` | **ADAPT** |
| `sql.go` | **ADAPT** |
| `cron.go` | **ADAPT** |

---

## 4. UI Package (`ui/`) — Embedded Frontend

| Path | Classification | Notes |
|------|---------------|-------|
| `ui/src/auth/pageInstaller.js` | **REPLACE** | "Setup your PocketBase instance" - installer UI |
| `ui/src/auth/pageSuperuserLogin.js` | **REPLACE** | "PocketBase" branding in login |
| `ui/src/auth/pageConfirmVerification.js` | **REPLACE** | |
| `ui/src/auth/pageConfirmPasswordReset.js` | **REPLACE** | |
| `ui/src/auth/pageConfirmEmailChange.js` | **REPLACE** | |
| `ui/src/auth/pageRequestSuperuserPasswordReset.js` | **REPLACE** | |
| `ui/src/auth/pageOAuth2Redirect*.js` | **REPLACE** | |
| `ui/embed.go` | **ADAPT** | Embed directive |
| `ui/embed_no_ui.go` | **ADAPT** | No-UI variant |
| `ui/package.json` | **KEEP-INTERNAL** | Vite build config |
| `ui/vite.config.js` | **KEEP-INTERNAL** | |

---

## 5. Migrations (`migrations/`) — 8 files

| File | Classification | Notes |
|------|---------------|-------|
| `1640988000_init.go` | **KEEP-INTERNAL** | System bootstrap |
| `1640988000_aux_init.go` | **KEEP-INTERNAL** | Aux DB init |
| `1717233556_v0.23_migrate.go` | **LICENSE/SECURITY-REQUIRED** | Upstream migration |
| `1717233557_v0.23_migrate2.go` | **LICENSE/SECURITY-REQUIRED** | |
| `1717233558_v0.23_migrate3.go` | **LICENSE/SECURITY-REQUIRED** | |
| `1717233559_v0.23_migrate4.go` | **LICENSE/SECURITY-REQUIRED** | |
| `1763020353_update_default_auth_alert_templates.go` | **ADAPT** | Auth templates |
| `1778828400_normalize_indexes.go` | **ADAPT** | Index normalization |

---

## 6. arsy/ Packages (A-R-S-Y Native)

### 6.1 arsy/aeng/ — Application Engine (15 files)
| File | Classification | Notes |
|------|---------------|-------|
| `aeng.go` | **KEEP-INTERNAL** | Package doc |
| `registry.go` | **ADAPT** | Model registry - references "PocketBase collection name" |
| `migration.go` | **KEEP-INTERNAL** | Migration executor |
| `seed.go` | **KEEP-INTERNAL** | Seed runner |
| `security.go` | **KEEP-INTERNAL** | Security enforcer |
| `validate.go` | **KEEP-INTERNAL** | Metadata validator |
| `mapper.go` | **KEEP-INTERNAL** | Go model mapper |
| `audit.go` | **KEEP-INTERNAL** | Audit logging (AENG-045) |
| `audit_logger.go` | **REMOVE-AFTER-MIGRATION** | Broken - references non-existent `core` package |
| `audit_logger_test.go` | **REMOVE-AFTER-MIGRATION** | Tests for broken logger |

### 6.2 arsy/tenancy/ — Multi-tenancy (27 files)
| File | Classification |
|------|---------------|
| `provisioner.go` | **KEEP-INTERNAL** |
| `outbox.go` | **KEEP-INTERNAL** |
| `sqlite_provider.go` | **KEEP-INTERNAL** |
| `router.go` | **KEEP-INTERNAL** |
| `catalog.go` | **ADAPT** |
| `pb_catalog.go` | **ADAPT** |
| `pb_memberships.go` | **ADAPT** |
| `migration.go` | **ADAPT** |

### 6.3 arsy/engine/ — Engine Components (11 subdirs)
| Subdir | Classification |
|--------|---------------|
| `bootstrap/` | **KEEP-INTERNAL** |
| `manifest/` | **KEEP-INTERNAL** |
| `registry/` | **KEEP-INTERNAL** |
| `persist/` | **KEEP-INTERNAL** |
| `query/` | **KEEP-INTERNAL** |
| `relations/` | **KEEP-INTERNAL** |
| `uow/` | **KEEP-INTERNAL** |
| `validation/` | **KEEP-INTERNAL** |
| `hooks/` | **KEEP-INTERNAL** |
| `fields/` | **KEEP-INTERNAL** |
| `fixtures/` | **KEEP-INTERNAL** |

### 6.4 Other arsy/
| Package | Classification |
|---------|---------------|
| `arsy/config/` | **KEEP-INTERNAL** |
| `arsy/console/` | **KEEP-INTERNAL** |
| `arsy/controlplane/` | **KEEP-INTERNAL** |
| `arsy/deployments/` | **KEEP-INTERNAL** |
| `arsy/httpapi/` | **KEEP-INTERNAL** |
| `arsy/migrations/` | **KEEP-INTERNAL** |
| `arsy/security/` | **KEEP-INTERNAL** |
| `arsy/servers/` | **KEEP-INTERNAL** |

---

## 7. Tools Package (`tools/`) — 21 subpackages

| Subpackage | Classification | Notes |
|------------|---------------|-------|
| `tools/hook/` | **KEEP-INTERNAL** | Event hooks |
| `tools/cron/` | **KEEP-INTERNAL** | Cron scheduler |
| `tools/filesystem/` | **KEEP-INTERNAL** | Local/S3 filesystem |
| `tools/mailer/` | **KEEP-INTERNAL** | SMTP/Sendmail |
| `tools/security/` | **KEEP-INTERNAL** | Crypto, tokens |
| `tools/tokenizer/` | **KEEP-INTERNAL** | JWT tokens |
| `tools/subscriptions/` | **KEEP-INTERNAL** | Realtime subscriptions |
| `tools/router/` | **KEEP-INTERNAL** | Request routing |
| `tools/list/` | **KEEP-INTERNAL** | List utilities |
| `tools/routine/` | **KEEP-INTERNAL** | Goroutine helpers |
| `tools/osutils/` | **KEEP-INTERNAL** | OS utilities |
| `tools/auth/` | **KEEP-INTERNAL** | Auth helpers |
| `tools/dbutils/` | **KEEP-INTERNAL** | DB utilities |
| `tools/inflector/` | **KEEP-INTERNAL** | String inflection |
| `tools/picker/` | **KEEP-INTERNAL** | Random selection |
| `tools/search/` | **KEEP-INTERNAL** | Search utilities |
| `tools/template/` | **KEEP-INTERNAL** | Template rendering |
| `tools/store/` | **KEEP-INTERNAL** | In-memory store |
| `tools/logger/` | **KEEP-INTERNAL** | Structured logging |
| `tools/archive/` | **KEEP-INTERNAL** | Archive handling |
| `tools/types/` | **KEEP-INTERNAL** | Type utilities |

---

## 8. Configuration & Build

| File | Classification |
|------|---------------|
| `go.mod` | **KEEP-INTERNAL** - Already `github.com/bridge-allah/a-r-s-y` |
| `go.sum` | **KEEP-INTERNAL** |
| `.goreleaser.yaml` | **ADAPT** - Release config |
| `golangci.yml` | **KEEP-INTERNAL** |
| `README.md` | **ADAPT** - Product documentation |

---

## 9. User-Facing Terminology Audit

| Term | Found In | Action |
|------|----------|--------|
| "PocketBase" | 381 occurrences in *.go | Replace with "A-R-S-Y" in user-facing strings |
| "Superuser" | Multiple (superuser.go, APIs, UI) | Replace with "Owner" or "Admin" |
| "pb_data" | pocketbase.go, config | Rename to "arsy_data" |
| "_superusers" collection | APIs, UI | Rename to "_owners" or "_admins" |
| "pbinstall" route | installer.go, UI | Rename to "arsyinstall" |
| "DefaultInstallerEmail" | core constants | Rename |

---

## 10. Dependency Graph Summary

```
pocketbase.go (entry)
  -> core.BaseApp (core/app.go, core/base.go)
      -> core.DB (dbx)
      -> core.Settings (core/settings_model.go)
      -> core.Migrations (core/migrations_runner.go)
      -> apis.Serve (apis/serve.go)
          -> apis middlewares
          -> apis record/collection handlers
          -> apis installer (apis/installer.go)
      -> ui.embed (ui/embed.go) -> Vite build -> ui/src/
      -> arsy.* packages (engine, tenancy, aeng, etc.)
      -> tools.* (hooks, cron, filesystem, etc.)
```

---

## 11. Classification Summary

| Classification | Count | Action |
|---------------|-------|--------|
| **KEEP-INTERNAL** | ~120 | Internal implementation, no user-facing changes |
| **ADAPT** | ~80 | Rename terminology, update strings, preserve logic |
| **REPLACE** | ~15 | User-facing UI/pages - rebuild with A-R-S-Y identity |
| **REMOVE-AFTER-MIGRATION** | 2 | Broken audit_logger files (non-existent deps) |
| **LICENSE/SECURITY-REQUIRED** | 4 | Upstream migrations - preserve for compatibility |

---

## 12. Next Steps (ARSY-001 Acceptance)

- [x] Inventory covers all root packages and shipped web/install surfaces
- [ ] Each classified item has owner, target phase, verification method
- [ ] Architecture review approves transformation boundary

**Evidence location:** `docs/architecture/pocketbase-transformation/inventory.md`