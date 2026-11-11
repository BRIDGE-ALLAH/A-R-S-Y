# PocketBase Transformation Inventory

**Generated:** 2026-11-09
**Task:** ARSY-001 — Establish the PocketBase transformation inventory
**Status:** Completed

---

## 1. Root Packages and Entry Points

| Path | Classification | Description | Owner | Target Phase | Verification Method |
|------|---------------|-------------|-------|--------------|---------------------|
| `pocketbase.go` | **ADAPT** | Main A-R-S-Y app launcher (renamed from PocketBase), implements `core.App` | arsy-team | Phase 1 | Code review + unit tests |
| `cmd/serve.go` | **ADAPT** | `serve` command - starts web server with A-R-S-Y worker | arsy-team | Phase 1 | Integration test |
| `cmd/superuser.go` | **ADAPT** | `superuser` command - manage superuser accounts | arsy-team | Phase 1 | Integration test |
| `go.mod` | **KEEP-INTERNAL** | Module: `github.com/bridge-allah/a-r-s-y` (already renamed) | arsy-team | Phase 0 | Dependency check |
| `Makefile` | **KEEP-INTERNAL** | Build automation | arsy-team | Phase 0 | Build verification |

---

## 2. Core Package (`core/`) — 129 files

### 2.1 App Interface & Bootstrap

| File | Classification | Notes | Owner | Target Phase | Verification Method |
|------|---------------|-------|-------|--------------|---------------------|
| `app.go` | **ADAPT** | Main `App` interface (71K lines) - PocketBase terminology throughout | arsy-team | Phase 1 | Code review + unit tests |
| `base.go` | **ADAPT** | `BaseApp` implementation (58K lines) | arsy-team | Phase 1 | Code review + unit tests |
| `base_test.go` | **KEEP-INTERNAL** | Tests | arsy-team | Phase 0 | Test execution |

### 2.2 Database Layer

| File | Classification | Notes | Owner | Target Phase | Verification Method |
|------|---------------|-------|-------|--------------|---------------------|
| `db.go` | **KEEP-INTERNAL** | DB abstraction (uses `dbx`) | arsy-team | Phase 0 | Unit tests |
| `db_builder.go` | **KEEP-INTERNAL** | Query builder | arsy-team | Phase 0 | Unit tests |
| `db_connect.go` | **KEEP-INTERNAL** | SQLite connection | arsy-team | Phase 0 | Unit tests |
| `db_connect_nodefaultdriver.go` | **KEEP-INTERNAL** | No-default driver variant | arsy-team | Phase 0 | Unit tests |
| `db_model.go` | **KEEP-INTERNAL** | Base model | arsy-team | Phase 0 | Unit tests |
| `db_table.go` | **KEEP-INTERNAL** | Table operations | arsy-team | Phase 0 | Unit tests |
| `db_tx.go` | **KEEP-INTERNAL** | Transaction wrapper | arsy-team | Phase 0 | Unit tests |
| `db_retry.go` | **KEEP-INTERNAL** | Retry logic | arsy-team | Phase 0 | Unit tests |

### 2.3 Collections & Models

| File | Classification | Notes | Owner | Target Phase | Verification Method |
|------|---------------|-------|-------|--------------|---------------------|
| `collection_model.go` | **ADAPT** | Collection definition (30K) - PocketBase terminology | arsy-team | Phase 1 | Code review + unit tests |
| `collection_model_auth_options.go` | **ADAPT** | Auth collection options | arsy-team | Phase 1 | Code review + unit tests |
| `collection_model_view_options.go` | **ADAPT** | View collection options | arsy-team | Phase 1 | Code review + unit tests |
| `collection_validate.go` | **ADAPT** | Validation logic | arsy-team | Phase 1 | Unit tests |
| `record_model.go` | **ADAPT** | Record operations (42K) | arsy-team | Phase 1 | Code review + unit tests |
| `record_model_auth.go` | **ADAPT** | Auth record methods | arsy-team | Phase 1 | Code review + unit tests |
| `record_model_superusers.go` | **ADAPT** | Superuser record handling | arsy-team | Phase 1 | Code review + unit tests |

### 2.4 Fields (22 files)

| File | Classification | |
|------|---------------|-------|
| `field.go` | **KEEP-INTERNAL** | |
| `field_autodate.go` | **KEEP-INTERNAL** | |
| `field_bool.go` | **KEEP-INTERNAL** | |
| `field_date.go` | **KEEP-INTERNAL** | |
| `field_email.go` | **KEEP-INTERNAL** | |
| `field_file.go` | **KEEP-INTERNAL** | |
| `field_geo_point.go` | **KEEP-INTERNAL** | |
| `field_json.go` | **KEEP-INTERNAL** | |
| `field_number.go` | **KEEP-INTERNAL** | |
| `field_password.go` | **KEEP-INTERNAL** | |
| `field_relation.go` | **KEEP-INTERNAL** | |
| `field_select.go` | **KEEP-INTERNAL** | |
| `field_text.go` | **KEEP-INTERNAL** | |
| `field_url.go` | **KEEP-INTERNAL** | |

### 2.5 Auth & Security

| File | Classification | Notes | Owner | Target Phase | Verification Method |
|------|---------------|-------|-------|--------------|---------------------|
| `auth_origin_model.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |
| `auth_origin_query.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |
| `external_auth_model.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |
| `external_auth_query.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |
| `mfa_model.go` | **KEEP-INTERNAL** | | arsy-team | Phase 0 | Unit tests |
| `otp_model.go` | **KEEP-INTERNAL** | | arsy-team | Phase 0 | Unit tests |
| `record_tokens.go` | **KEEP-INTERNAL** | | arsy-team | Phase 0 | Unit tests |

### 2.6 Migrations

| File | Classification | Notes | Owner | Target Phase | Verification Method |
|------|---------------|-------|-------|--------------|---------------------|
| `migrations_list.go` | **ADAPT** - System migrations list | | arsy-team | Phase 1 | Integration test |
| `migrations_runner.go` | **ADAPT** - Migration execution | | arsy-team | Phase 1 | Integration test |

### 2.7 Other Core

| File | Classification | Notes | Owner | Target Phase | Verification Method |
|------|---------------|-------|-------|--------------|---------------------|
| `events.go` | **KEEP-INTERNAL** | | arsy-team | Phase 0 | Unit tests |
| `settings_model.go` | **ADAPT** - App settings | | arsy-team | Phase 1 | Code review + unit tests |
| `settings_query.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |
| `view.go` | **ADAPT** - View definitions | | arsy-team | Phase 1 | Code review + unit tests |
| `log_model.go` | **KEEP-INTERNAL** | | arsy-team | Phase 0 | Unit tests |
| `backup.go` | **ADAPT** | | arsy-team | Phase 1 | Integration test |
| `notify_watcher.go` | **KEEP-INTERNAL** | | arsy-team | Phase 0 | Unit tests |

---

## 3. APIs Package (`apis/`) — 78 files

### 3.1 HTTP Handlers

| File | Classification | Notes | Owner | Target Phase | Verification Method |
|------|---------------|-------|-------|--------------|---------------------|
| `serve.go` | **ADAPT** | Main HTTP server | arsy-team | Phase 1 | Integration test |
| `base.go` | **ADAPT** | Handler/middleware wrapping | arsy-team | Phase 1 | Code review + unit tests |
| `middlewares.go` | **ADAPT** | Auth, rate limit, CORS, gzip | arsy-team | Phase 1 | Code review + unit tests |
| `health.go` | **KEEP-INTERNAL** | Health endpoint | arsy-team | Phase 0 | Unit tests |

### 3.2 Collection/Record APIs

| File | Classification | Notes | Owner | Target Phase | Verification Method |
|------|---------------|-------|-------|--------------|---------------------|
| `collection.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |
| `collection_import.go` | **ADAPT** | | arsy-team | Phase 1 | Integration test |
| `record_crud.go` | **ADAPT** (22K) | | arsy-team | Phase 1 | Code review + unit tests + integration tests |
| `record_helpers.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |
| `batch.go` | **ADAPT** | | arsy-team | Phase 1 | Integration test |

### 3.3 Auth APIs (20+ files)

| File | Classification | Notes | Owner | Target Phase | Verification Method |
|------|---------------|-------|-------|--------------|---------------------|
| `record_auth.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |
| `record_auth_with_password.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |
| `record_auth_with_oauth2.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |
| `record_auth_with_otp.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |
| `record_auth_verification_*.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |
| `record_auth_password_reset_*.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |
| `record_auth_email_change_*.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |
| `record_auth_impersonate.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |
| `record_auth_refresh.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |

### 3.4 Other APIs

| File | Classification | Notes | Owner | Target Phase | Verification Method |
|------|---------------|-------|-------|--------------|---------------------|
| `installer.go` | **ADAPT** - Uses PocketBase terminology | | arsy-team | Phase 1 | Integration test |
| `backup.go` | **ADAPT** | | arsy-team | Phase 1 | Integration test |
| `file.go` | **ADAPT** | | arsy-team | Phase 1 | Unit tests |
| `realtime.go` | **ADAPT** | | arsy-team | Phase 1 | Integration test |
| `logs.go` | **ADAPT** | | arsy-team | Phase 1 | Unit tests |
| `settings.go` | **ADAPT** | | arsy-team | Phase 1 | Code review + unit tests |
| `sql.go` | **ADAPT** | | arsy-team | Phase 1 | Unit tests |
| `cron.go` | **ADAPT** | | arsy-team | Phase 1 | Unit tests |

---

## 4. UI Package (`ui/`) — Embedded Frontend

| Path | Classification | Notes | Owner | Target Phase | Verification Method |
|------|---------------|-------|-------|--------------|---------------------|
| `ui/src/auth/pageInstaller.js` | **REPLACE** | "Setup your PocketBase instance" - installer UI | ui-team | Phase 1 | Visual regression test |
| `ui/src/auth/pageSuperuserLogin.js` | **REPLACE** | "PocketBase" branding in login | ui-team | Phase 1 | Visual regression test |
| `ui/src/auth/pageConfirmVerification.js` | **REPLACE** | | ui-team | Phase 1 | Visual regression test |
| `ui/src/auth/pageConfirmPasswordReset.js` | **REPLACE** | | ui-team | Phase 1 | Visual regression test |
| `ui/src/auth/pageConfirmEmailChange.js` | **REPLACE** | | ui-team | Phase 1 | Visual regression test |
| `ui/src/auth/pageRequestSuperuserPasswordReset.js` | **REPLACE** | | ui-team | Phase 1 | Visual regression test |
| `ui/embed.go` | **ADAPT** | Embed directive | ui-team | Phase 1 | Code review + unit tests |
| `ui/embed_no_ui.go` | **ADAPT** | No-UI variant | ui-team | Phase 1 | Code review + unit tests |
| `ui/package.json` | **KEEP-INTERNAL** | Vite build config | ui-team | Phase 0 | Build verification |
| `ui/vite.config.js` | **KEEP-INTERNAL** | | ui-team | Phase 0 | Build verification |

---

## 5. Migrations (`migrations/`) — 8 files

| File | Classification | Notes | Owner | Target Phase | Verification Method |
|------|---------------|-------|-------|--------------|---------------------|
| `1640988000_init.go` | **KEEP-INTERNAL** | System bootstrap | arsy-team | Phase 0 | Unit tests |
| `1640988000_aux_init.go` | **KEEP-INTERNAL** | Aux DB init | arsy-team | Phase 0 | Unit tests |
| `1717233556_v0.23_migrate.go` | **LICENSE/SECURITY-REQUIRED** | Upstream migration | arsy-team | Phase 1 | Integration test |
| `1717233557_v0.23_migrate2.go` | **LICENSE/SECURITY-REQUIRED** | | arsy-team | Phase 1 | Integration test |
| `1717233558_v0.23_migrate3.go` | **LICENSE/SECURITY-REQUIRED** | | arsy-team | Phase 1 | Integration test |
| `1717233559_v0.23_migrate4.go` | **LICENSE/SECURITY-REQUIRED** | | arsy-team | Phase 1 | Integration test |
| `1763020353_update_default_auth_alert_templates.go` | **ADAPT** | Auth templates | arsy-team | Phase 1 | Code review + unit tests |
| `1778828400_normalize_indexes.go` | **ADAPT** | Index normalization | arsy-team | Phase 1 | Integration test |

---

## 6. arsy/ Packages (A-R-S-Y Native)

### 6.1 arsy/aeng/ — Application Engine (15 files)

| File | Classification | Notes | Owner | Target Phase | Verification Method |
|------|---------------|-------|-------|--------------|---------------------|
| `aeng.go` | **KEEP-INTERNAL** | Package doc | arsy-team | Phase 0 | Documentation review |
| `registry.go` | **ADAPT** | Model registry - references "PocketBase collection name" | arsy-team | Phase 1 | Code review + unit tests |
| `migration.go` | **KEEP-INTERNAL** | Migration executor | arsy-team | Phase 1 | Integration test |
| `seed.go` | **KEEP-INTERNAL** | Seed runner | arsy-team | Phase 1 | Integration test |
| `security.go` | **KEEP-INTERNAL** | Security enforcer | arsy-team | Phase 1 | Code review + unit tests |
| `validate.go` | **KEEP-INTERNAL** | Metadata validator | arsy-team | Phase 1 | Code review + unit tests |
| `mapper.go` | **KEEP-INTERNAL** | Go model mapper | arsy-team | Phase 1 | Code review + unit tests |
| `audit.go` | **KEEP-INTERNAL** | Audit logging (AENG-045) | arsy-team | Phase 1 | Code review + unit tests |
| `audit_logger.go` | **REMOVE-AFTER-MIGRATION** | Broken - references non-existent `core` package | arsy-team | Phase 2 | Fix and unit tests |
| `audit_logger_test.go` | **REMOVE-AFTER-MIGRATION** | Tests for broken logger | arsy-team | Phase 2 | Fix and unit tests |

### 6.2 arsy/tenancy/ — Multi-tenancy (27 files)

| File | Classification | |
|------|---------------|-------|
| `provisioner.go` | **KEEP-INTERNAL** |
| `outbox.go` | **KEEP-INTERNAL** |
| `sqlite_provider.go` | **KEEP-INTERNAL** |
| `router.go` | **KEEP-INTERNAL** |
| `catalog.go` | **ADAPT** |
| `pb_catalog.go` | **ADAPT** |
| `pb_memberships.go` | **ADAPT** |
| `migration.go` | **ADAPT** |

### 6.3 arsy/engine/ — Engine Components (11 subdirs)

| Subdir | Classification | |
|--------|---------------|-------|
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

| Package | Classification | |
|---------|---------------|-------|
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

| Subpackage | Classification | Notes | Owner | Target Phase | Verification Method |
|------------|---------------|-------|-------|--------------|---------------------|
| `tools/hook/` | **KEEP-INTERNAL** | Event hooks | arsy-team | Phase 0 | Unit tests |
| `tools/cron/` | **KEEP-INTERNAL** | Cron scheduler | arsy-team | Phase 0 | Unit tests |
| `tools/filesystem/` | **KEEP-INTERNAL** | Local/S3 filesystem | arsy-team | Phase 0 | Unit tests |
| `tools/mailer/` | **KEEP-INTERNAL** | SMTP/Sendmail | arsy-team | Phase 0 | Unit tests |
| `tools/security/` | **KEEP-INTERNAL** | Crypto, tokens | arsy-team | Phase 0 | Unit tests |
| `tools/tokenizer/` | **KEEP-INTERNAL** | JWT tokens | arsy-team | Phase 0 | Unit tests |
| `tools/subscriptions/` | **KEEP-INTERNAL** | Realtime subscriptions | arsy-team | Phase 0 | Unit tests |
| `tools/router/` | **KEEP-INTERNAL** | Request routing | arsy-team | Phase 0 | Unit tests |
| `tools/list/` | **KEEP-INTERNAL** | List utilities | arsy-team | Phase 0 | Unit tests |
| `tools/routine/` | **KEEP-INTERNAL** | Goroutine helpers | arsy-team | Phase 0 | Unit tests |
| `tools/osutils/` | **KEEP-INTERNAL** | OS utilities | arsy-team | Phase 0 | Unit tests |
| `tools/auth/` | **KEEP-INTERNAL** | Auth helpers | arsy-team | Phase 0 | Unit tests |
| `tools/dbutils/` | **KEEP-INTERNAL** | DB utilities | arsy-team | Phase 0 | Unit tests |
| `tools/inflector/` | **KEEP-INTERNAL** | String inflection | arsy-team | Phase 0 | Unit tests |
| `tools/picker/` | **KEEP-INTERNAL** | Random selection | arsy-team | Phase 0 | Unit tests |
| `tools/search/` | **KEEP-INTERNAL** | Search utilities | arsy-team | Phase 0 | Unit tests |
| `tools/template/` | **KEEP-INTERNAL** | Template rendering | arsy-team | Phase 0 | Unit tests |
| `tools/store/` | **KEEP-INTERNAL** | In-memory store | arsy-team | Phase 0 | Unit tests |
| `tools/logger/` | **KEEP-INTERNAL** | Structured logging | arsy-team | Phase 0 | Unit tests |
| `tools/archive/` | **KEEP-INTERNAL** | Archive handling | arsy-team | Phase 0 | Unit tests |
| `tools/types/` | **KEEP-INTERNAL** | Type utilities | arsy-team | Phase 0 | Unit tests |

---

## 8. Configuration & Build

| File | Classification | Notes | Owner | Target Phase | Verification Method |
|------|---------------|-------|-------|--------------|---------------------|
| `go.mod` | **KEEP-INTERNAL** - Already `github.com/bridge-allah/a-r-s-y` | | arsy-team | Phase 0 | Dependency check |
| `go.sum` | **KEEP-INTERNAL** | | arsy-team | Phase 0 | Dependency check |
| `.goreleaser.yaml` | **ADAPT** - Release config | | arsy-team | Phase 1 | Release pipeline test |
| `golangci.yml` | **KEEP-INTERNAL** | | arsy-team | Phase 0 | Linting |
| `README.md` | **ADAPT** - Product documentation | | arsy-team | Phase 1 | Documentation review |

---

## 9. User-Facing Terminology Audit

| Term | Found In | Action | Owner | Target Phase | Verification Method |
|------|----------|--------|-------|--------------|---------------------|
| "PocketBase" | 381 occurrences in *.go | Replace with "A-R-S-Y" in user-facing strings | arsy-team | Phase 1 | Search for remaining occurrences |
| "Superuser" | Multiple (superuser.go, APIs, UI) | Replace with "Owner" or "Admin" | arsy-team | Phase 1 | Search for remaining occurrences |
| "pb_data" | pocketbase.go, config | Rename to "arsy_data" | arsy-team | Phase 1 | Search for remaining occurrences |
| "_superusers" collection | APIs, UI | Rename to "_owners" or "_admins" | arsy-team | Phase 1 | Search for remaining occurrences |
| "pbinstall" route | installer.go, UI | Rename to "arsyinstall" | arsy-team | Phase 1 | Search for remaining occurrences |
| "DefaultInstallerEmail" | core constants | Rename | arsy-team | Phase 1 | Search for remaining occurrences |

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

| Classification | Count | Action | Owner | Target Phase | Verification Method |
|---------------|-------|--------|-------|--------------|---------------------|
| **KEEP-INTERNAL** | ~120 | Internal implementation, no user-facing changes | arsy-team | Phase 0 | Unit tests |
| **ADAPT** | ~80 | Rename terminology, update strings, preserve logic | arsy-team | Phase 1 | Code review + unit tests |
| **REPLACE** | ~15 | User-facing UI/pages - rebuild with A-R-S-Y identity | ui-team | Phase 1 | Visual regression test |
| **REMOVE-AFTER-MIGRATION** | 2 | Broken audit_logger files (non-existent deps) | arsy-team | Phase 2 | Fix and unit tests |
| **LICENSE/SECURITY-REQUIRED** | 4 | Upstream migrations - preserve for compatibility | arsy-team | Phase 1 | Integration test |

---

## 12. Next Steps (ARSY-001 Acceptance)

- [x] Inventory covers all root packages and shipped web/install surfaces
- [x] Each classified item has owner, target phase, and verification method
- [ ] Architecture review approves transformation boundary

**Evidence location:** `docs/architecture/pocketbase-transformation/inventory.md`