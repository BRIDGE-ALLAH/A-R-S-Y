# A-R-S-Y Studio Verification Report

**Task:** ARSY-010 — Verify A-R-S-Y Studio is the sole administration experience  
**Status:** In Progress  
**Depends on:** ARSY-001, ARSY-002 (completed)  
**Generated:** 2026-11-10

---

## 1. Shipped Studio Routes Inventory

### 1.1 Guest-Only Routes (Unauthenticated)
| Route | Handler | Purpose | Legacy Terminology |
|-------|---------|---------|-------------------|
| `#/pbinstall/{token}` | `pageInstaller` | First admin setup | **pbinstall** (should be `arsyinstall`) |
| `#/login` | `pageSuperuserLogin` | Admin login | **Superuser** (should be Owner/Admin) |
| `#/request-password-reset` | `pageRequestSuperuserPasswordReset` | Password reset | **Superuser** |
| `#/auth/confirm-password-reset/{token}` | `pageConfirmPasswordReset` | Confirm reset | PocketBase SDK import |
| `#/auth/confirm-verification/{token}` | `pageConfirmVerification` | Email verification | PocketBase SDK import |
| `#/auth/confirm-email-change/{token}` | `pageConfirmEmailChange` | Email change | PocketBase SDK import |
| `#/auth/oauth2-redirect-success` | `pageOAuth2RedirectSuccess` | OAuth2 success | — |
| `#/auth/oauth2-redirect-failure` | `pageOAuth2RedirectFailure` | OAuth2 failure | — |

### 1.2 Superuser-Only Routes (Authenticated)
| Route | Handler | Purpose | Legacy Terminology |
|-------|---------|---------|-------------------|
| `#/collections` | `pageCollections` | Model management | — |
| `#/logs` | `pageLogs` | Request logs | **Superuser** references in filter presets |
| `#/settings` | `pageApplicationSettings` | App settings | **Superuser** IPs, trusted proxy |
| `#/settings/mail` | `pageMailSettings` | Mail config | "PocketBase uses sendmail" |
| `#/settings/storage` | `pageStorageSettings` | File storage | "PocketBase uses local FS" |
| `#/settings/backups` | `pageBackupsSettings` | Backup/restore | "Backup your PocketBase data" |
| `#/settings/crons` | `pageCronsSettings` | Cron jobs | — |
| `#/settings/export-collections` | `pageExportCollections` | Export schema | "Import in another PocketBase" |
| `#/settings/import-collections` | `pageImportCollections` | Import schema | — |
| `#/settings/sql` | `pageSQLConsole` | SQL console | — |

### 1.3 Missing A-R-S-Y Native Routes (Gaps)
| Expected Route | Purpose | Status |
|----------------|---------|--------|
| `#/tenants` | Tenant management | **MISSING** |
| `#/modules` | Module marketplace/management | **MISSING** |
| `#/models` | Model designer (A-R-S-Y native) | **MISSING** (uses `/collections`) |
| `#/migrations` | Migration history/status | **MISSING** |
| `#/health` | Runtime health dashboard | **MISSING** |
| `#/api-keys` | API key management | **MISSING** |
| `#/audit` | Audit log viewer | **MISSING** |

---

## 2. UI Screens & Terminology Audit

### 2.1 Files Requiring Terminology Replacement (PRIORITY: REPLACE)

| File | Lines | Legacy Terms | A-R-S-Y Replacement |
|------|-------|--------------|---------------------|
| `ui/src/auth/pageInstaller.js` | 13, 36, 45 | "Setup your PocketBase instance", `_superusers` | "Setup your A-R-S-Y instance", `_admins` |
| `ui/src/auth/pageSuperuserLogin.js` | 1, 6, 16, 31, 38, 52 | `pageSuperuserLogin`, "Superuser login", `_superusers` | `pageOwnerLogin`, "Owner login", `_admins` |
| `ui/src/auth/pageRequestSuperuserPasswordReset.js` | 1, 6, 11 | `pageRequestSuperuserPasswordReset`, `_superusers` | `pageRequestOwnerPasswordReset`, `_admins` |
| `ui/src/router.js` | 1, 129, 133, 136, 165 | `pageSuperuserLogin`, `pbinstall`, `_superusers` | `pageOwnerLogin`, `arsyinstall`, `_admins` |
| `ui/src/base/appHeader.js` | — | `_superusers` in nav | `_admins` |
| `ui/src/base/fieldSettings.js` | — | "Superuser UI relation listings" | "Admin UI relation listings" |
| `ui/src/base/ruleField.js` | — | "(Superusers only)" | "(Admins only)" |
| `ui/src/collections/collectionsOverviewModal.js` | — | "Superusers only" label | "Admins only" |
| `ui/src/collections/mfaAccordion.js` | — | `isSuperusers`, "Superusers" | `isAdmins`, "Admins" |
| `ui/src/collections/otpAccordion.js` | — | `isSuperusers`, "Superusers" | `isAdmins`, "Admins" |
| `ui/src/collections/tokenOptionsAccordion.js` | — | `_superusers` | `_admins` |
| `ui/src/fields/text/view.js` | — | `_superusers` | `_admins` |
| `ui/src/logs/pageLogs.js` | Multiple | `SUPERUSER_REQUESTS`, `superuser` | `ADMIN_REQUESTS`, `admin` |
| `ui/src/pb.js` | Multiple | `PocketBase`, `__pb_superusers__`, `_superusers` | `ARSY`, `__arsy_admins__`, `_admins` |
| `ui/src/store.js` | — | `PocketBase ${version}` | `A-R-S-Y ${version}` |
| `ui/src/settings/application/pageApplicationSettings.js` | — | `superuserIPs`, `Superuser IPs` | `adminIPs`, `Admin IPs` |
| `ui/src/settings/application/superuserAccordion.js` | — | "Superuser IPs", "Superuser IPs and subnets" | "Admin IPs" |
| `ui/src/settings/application/rateLimitAccordion.js` | — | "Superusers are always excluded" | "Admins are always excluded" |
| `ui/src/settings/application/trustedProxyAccordion.js` | Multiple | "PocketBase" (5 occurrences) | "A-R-S-Y" |
| `ui/src/settings/backups/pageBackupsSettings.js` | — | "Backup and restore your PocketBase data" | "Backup and restore your A-R-S-Y data" |
| `ui/src/settings/mail/pageMailSettings.js` | — | "PocketBase uses sendmail" | "A-R-S-Y uses sendmail" |
| `ui/src/settings/storage/pageStorageSettings.js` | — | "PocketBase uses local file system" | "A-R-S-Y uses local file system" |
| `ui/src/settings/sync/pageExportCollections.js` | — | "another PocketBase environment" | "another A-R-S-Y environment" |
| `ui/src/records/recordsList.js` | — | `isSuperusersCollection`, `_superusers` | `isAdminsCollection`, `_admins` |
| `ui/src/records/recordUpsertModal.js` | — | `isSuperusersCollection`, `_superusers` | `isAdminsCollection`, `_admins` |
| `ui/src/records/recordImpersonateModal.js` | — | `PocketBase` SDK import | `ARSY` SDK import |
| `ui/src/apiPreview/*.js` | 20+ files | `PocketBase` SDK, `_superusers`, "Superusers" | `ARSY` SDK, `_admins`, "Admins" |

### 2.2 PocketBase SDK Imports (27 files in apiPreview)
All `apiPreview/docs*.js` files import `PocketBase` from `'pocketbase'` for code generation examples. These are **developer-facing documentation** and should show A-R-S-Y SDK usage.

---

## 3. API Routes Audit (apis/)

### 3.1 Routes Requiring Superuser Auth (Use `RequireSuperuserAuth()` middleware)
| Endpoint | Method | Middleware | Legacy Name |
|----------|--------|------------|-------------|
| `/api/collections` | GET/POST | `RequireSuperuserAuth()` | — |
| `/api/collections/*` | GET/PATCH/DELETE | `RequireSuperuserAuth()` | — |
| `/api/backups` | GET/POST/DELETE | `RequireSuperuserAuth()` | — |
| `/api/backups/upload` | POST | `RequireSuperuserAuth()` | — |
| `/api/backups/{key}/restore` | POST | `RequireSuperuserAuth()` | — |
| `/api/crons` | GET/POST | `RequireSuperuserAuth()` | — |
| `/api/logs` | GET | `RequireSuperuserAuth()` | — |

### 3.2 Terminology in Middleware & Code
| File | Legacy Term | Replacement |
|------|-------------|-------------|
| `apis/middlewares.go` | `pbSuperuserIPsWhitelist` | `arsyAdminIPsWhitelist` |
| `apis/middlewares.go` | `pbRequireSuperuserAuth` | `arsyRequireAdminAuth` |
| `apis/middlewares.go` | `pbRequireSuperuserOrOwnerAuth` | `arsyRequireAdminOrOwnerAuth` |
| `apis/installer.go` | `pbinstall` route | `arsyinstall` |
| `apis/installer.go` | `DefaultInstallerFunc`, `systemSuperuser` | `DefaultBootstrapFunc`, `systemAdmin` |
| `apis/backup.go` | `IsSuperuser()`, `SuperuserIPs` | `IsAdmin()`, `AdminIPs` |
| `apis/file.go` | `IsSuperuser()`, `SuperuserIPs` | `IsAdmin()`, `AdminIPs` |
| `apis/health.go` | `HasSuperuserAuth()` | `HasAdminAuth()` |
| `apis/collection.go` | `HasSuperuserAuth()` | `HasAdminAuth()` |

---

## 4. Core Package Terminology

| File | Legacy Term | Status |
|------|-------------|--------|
| `core/app.go` | Package comment: "backbone of PocketBase" | **ADAPT** - internal comment |
| `core/base.go` | `pb_data`, `registerSuperuserHooks()` | **ADAPT** - internal |
| `core/events.go` | `InstallerFunc`, `systemSuperuser` | **ADAPT** - internal |
| `core/event_request.go` | `HasSuperuserAuth()`, `IsSuperuser()` | **ADAPT** - internal |
| `core/collection_model.go` | "_superusers" reference | **ADAPT** - internal |
| `core/backup.go` | `sendSystemAlertToAllSuperusers` | **ADAPT** - internal |

**Note:** Core package terminology is INTERNAL implementation detail. Per ARSY-002, these must not leak into public APIs, but internal identifiers can remain for compatibility.

---

## 5. Reproducible Defects (Legacy Admin UI Still Shipped)

| Defect | Location | Evidence |
|--------|----------|----------|
| **Installer route** | `ui/src/router.js:129` | `#/pbinstall/{token}` still registered |
| **Login page** | `ui/src/auth/pageSuperuserLogin.js` | "Superuser login" title, `_superusers` collection |
| **Password reset** | `ui/src/auth/pageRequestSuperuserPasswordReset.js` | "Superuser" in function name and PBEvent |
| **API Preview** | `ui/src/apiPreview/*.js` (27 files) | All examples use `PocketBase` SDK |
| **Settings** | `ui/src/settings/application/superuserAccordion.js` | "Superuser IPs" section |
| **Logs** | `ui/src/logs/pageLogs.js` | `SUPERUSER_REQUESTS` filter presets |
| **Client SDK** | `ui/src/pb.js` | `new PocketBase()`, `__pb_superusers__` auth store |
| **Installer API** | `apis/installer.go` | `/pbinstall` route, `systemSuperuser` |
| **Middleware** | `apis/middlewares.go` | `pbRequireSuperuserAuth` middleware ID |

---

## 6. Canonical Studio App & Route Ownership

```
A-R-S-Y Studio (Single Admin Console)
├── Guest Routes (Public)
│   ├── /arsyinstall/{token} → Bootstrap/Onboarding
│   ├── /login → Owner/Admin Authentication
│   ├── /auth/* → Email verification, password reset, OAuth2
├── Authenticated Routes (Owner/Admin Only)
│   ├── /models → Model Designer (NEW - replaces /collections)
│   ├── /records → Record Management (per model)
│   ├── /tenants → Tenant Lifecycle (NEW)
│   ├── /modules → Module Registry (NEW)
│   ├── /migrations → Schema Migrations (NEW)
│   ├── /health → Runtime Health (NEW)
│   ├── /api-keys → API Key Management (NEW)
│   ├── /audit → Audit Log Viewer (NEW)
│   ├── /logs → Request Logs (EXISTING - rename filter)
│   ├── /settings → Application Settings (EXISTING - rename terms)
│   │   ├── /general → App name, URL, admin IPs
│   │   ├── /mail → SMTP/Sendmail config
│   │   ├── /storage → File storage (local/S3)
│   │   ├── /backups → Backup/Restore
│   │   ├── /crons → Scheduled jobs
│   │   ├── /import-export → Schema migration
│   │   └── /sql → SQL console
│   └── /diagnostics → System diagnostics (NEW)
```

### Route Ownership
| Route Prefix | Owner Package | API Layer |
|--------------|---------------|-----------|
| `/api/collections/*` | `apis/collection.go` | `apis` |
| `/api/records/*` | `apis/record_crud.go` | `apis` |
| `/api/backups/*` | `apis/backup.go` | `apis` |
| `/api/logs/*` | `apis/logs.go` | `apis` |
| `/api/crons/*` | `apis/cron.go` | `apis` |
| `/api/settings/*` | `apis/settings.go` | `apis` |
| `/api/tenants/*` | `arsy/tenancy` + `apis` | `arsy/httpapi` (NEW) |
| `/api/modules/*` | `arsy/engine/manifest` + `apis` | `arsy/httpapi` (NEW) |
| `/api/migrations/*` | `arsy/aeng/migration.go` + `apis` | `arsy/httpapi` (NEW) |
| `/api/audit/*` | `arsy/aeng/audit.go` + `apis` | `arsy/httpapi` (NEW) |
| `/api/keys/*` | `arsy/security` + `apis` | `arsy/httpapi` (NEW) |
| `/api/health/*` | `apis/health.go` | `apis` |

---

## 7. Verification Checklist (ARSY-010 Acceptance)

- [x] Inventory shipped Studio routes, screens, navigation, assets, API contracts
- [x] Verified model, record, tenant, module, permission, migration, runtime-health workflows (gaps identified)
- [x] Recorded remaining legacy admin UI/routes as reproducible defects (9 defects)
- [ ] Trace production startup to embedded web assets and API handlers
- [ ] Document canonical Studio app, route ownership, typed API boundary
- [ ] Implement only evidenced gaps in existing A-R-S-Y Studio

---

## 8. Next Steps

1. **Rename terminology** across UI (27+ files) - `Superuser` → `Admin/Owner`, `PocketBase` → `A-R-S-Y`, `_superusers` → `_admins`, `pbinstall` → `arsyinstall`
2. **Add missing routes** - Tenants, Modules, Models, Migrations, Health, API Keys, Audit, Diagnostics
3. **Update API middleware** - `RequireSuperuserAuth` → `RequireAdminAuth`
4. **Update installer** - `pbinstall` route, `DefaultInstallerFunc` → `DefaultBootstrapFunc`
5. **API Preview examples** - Replace `PocketBase` SDK with A-R-S-Y SDK in 27 files

**Evidence location:** `docs/architecture/pocketbase-transformation/studio-verification.md`