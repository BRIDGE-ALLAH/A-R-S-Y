# A-R-S-Y Brand Scan Report

**Task:** ARSY-012 — Verify A-R-S-Y identity across all shipped product surfaces  
**Status:** In Progress  
**Depends on:** ARSY-001, ARSY-010  
**Generated:** 2026-11-10

---

## 1. Brand Scan Results Summary

**Total matches (filtered):** 553 occurrences across 50+ files  
**Scan command:** `grep -r "PocketBase\|Superuser\|pb_data\|_superusers" --include="*.go" --include="*.js" --include="*.html" . | grep -v "_test.go" | grep -v "test_" | grep -v "internal" | grep -v "license" | grep -v "LICENSE" | grep -v "doc.go" | grep -v "node_modules" | grep -v "\.plan" | grep -v "docs/aeng-" | grep -v "CHANGELOG" | grep -v "README.md" | grep -v "CONTRIBUTING.md" | grep -v "SECURITY.md"`

### 1.1 Classification by Category

| Category | Count | Files | Action Required |
|----------|-------|-------|-----------------|
| **User-facing (UI/API/CLI)** | 380 | 35 | REPLACE - Must change |
| **Internal (core/)** | 120 | 15 | KEEP-INTERNAL - Internal compatibility |
| **Documentation (CHANGELOG/legacy)** | 53 | 3 | ARCHIVE - Historical reference only |
| **Dist/Build artifacts** | 27 | 2 | REBUILD - Rebuild from source |

---

## 2. User-Facing Defects (Must Fix)

### 2.1 UI Source Files (ui/src/)

| File | Matches | Legacy Terms | A-R-S-Y Replacement |
|------|---------|--------------|---------------------|
| `ui/src/auth/pageInstaller.js` | 3 | "Setup your PocketBase instance", `_superusers` | "Setup your A-R-S-Y instance", `_admins` |
| `ui/src/auth/pageSuperuserLogin.js` | 7 | `pageSuperuserLogin`, "Superuser login", `_superusers` | `pageOwnerLogin`, "Owner login", `_admins` |
| `ui/src/auth/pageRequestSuperuserPasswordReset.js` | 3 | `pageRequestSuperuserPasswordReset`, `_superusers` | `pageRequestOwnerPasswordReset`, `_admins` |
| `ui/src/auth/pageConfirmVerification.js` | 3 | PocketBase SDK import | ARSY SDK import |
| `ui/src/router.js` | 5 | `pageSuperuserLogin`, `pbinstall`, `_superusers` | `pageOwnerLogin`, `arsyinstall`, `_admins` |
| `ui/src/settings/application/pageApplicationSettings.js` | 5 | `superuserIPs`, `Superuser IPs` | `adminIPs`, `Admin IPs` |
| `ui/src/settings/application/superuserAccordion.js` | 3 | "Superuser IPs", "Superuser IPs and subnets" | "Admin IPs" |
| `ui/src/settings/application/trustedProxyAccordion.js` | 5 | "PocketBase" (5x) | "A-R-S-Y" |
| `ui/src/settings/backups/pageBackupsSettings.js` | 2 | "Backup and restore your PocketBase data" | "Backup and restore your A-R-S-Y data" |
| `ui/src/logs/pageLogs.js` | 10 | `SUPERUSER_REQUESTS`, `superuser` | `ADMIN_REQUESTS`, `admin` |
| `ui/src/pb.js` | 4 | `PocketBase`, `__pb_superusers__`, `_superusers` | `ARSY`, `__arsy_admins__`, `_admins` |
| `ui/src/records/recordsList.js` | 4 | `isSuperusersCollection`, `_superusers` | `isAdminsCollection`, `_admins` |
| `ui/src/records/recordUpsertModal.js` | 6 | `isSuperusersCollection`, `_superusers` | `isAdminsCollection`, `_admins` |
| `ui/src/records/recordImpersonateModal.js` | 3 | `PocketBase` SDK import | `ARSY` SDK import |
| `ui/src/collections/otpAccordion.js` | 5 | `isSuperusers`, `Superusers` | `isAdmins`, `Admins` |
| `ui/src/collections/mfaAccordion.js` | 3 | `isSuperusers`, `Superusers` | `isAdmins`, `Admins` |
| `ui/src/collections/tokenOptionsAccordion.js` | 2 | `_superusers` | `_admins` |
| `ui/src/base/appHeader.js` | 1 | `_superusers` in nav | `_admins` |
| `ui/src/base/fieldSettings.js` | 1 | "Superuser UI relation listings" | "Admin UI relation listings" |
| `ui/src/base/ruleField.js` | 1 | "(Superusers only)" | "(Admins only)" |
| `ui/src/collections/collectionsOverviewModal.js` | 1 | "Superusers only" label | "Admins only" |
| `ui/src/fields/text/view.js` | 1 | `_superusers` | `_admins` |
| `ui/src/apiPreview/*.js` (27 files) | 40+ | `PocketBase` SDK, `_superusers`, "Superusers" | `ARSY` SDK, `_admins`, "Admins" |

### 2.2 API/CLI Files (apis/, cmd/)

| File | Matches | Legacy Terms | A-R-S-Y Replacement |
|------|---------|--------------|---------------------|
| `apis/middlewares.go` | 19 | `pbSuperuserIPsWhitelist`, `pbRequireSuperuserAuth`, `pbRequireSuperuserOrOwnerAuth` | `arsyAdminIPsWhitelist`, `arsyRequireAdminAuth`, `arsyRequireAdminOrOwnerAuth` |
| `apis/installer.go` | 11 | `DefaultInstallerFunc`, `systemSuperuser`, `pbinstall` | `DefaultBootstrapFunc`, `systemAdmin`, `arsyinstall` |
| `apis/backup.go` | 7 | `IsSuperuser()`, `SuperuserIPs` | `IsAdmin()`, `AdminIPs` |
| `apis/file.go` | 4 | `IsSuperuser()`, `SuperuserIPs` | `IsAdmin()`, `AdminIPs` |
| `apis/health.go` | 1 | `HasSuperuserAuth()` | `HasAdminAuth()` |
| `apis/collection.go` | 3 | `HasSuperuserAuth()` | `HasAdminAuth()` |
| `apis/record_crud.go` | 21 | `HasSuperuserAuth()`, `_superusers` | `HasAdminAuth()`, `_admins` |
| `apis/record_helpers.go` | 8 | `_superusers` | `_admins` |
| `apis/logs.go` | 2 | `RequireSuperuserAuth()` | `RequireAdminAuth()` |
| `apis/cron.go` | 1 | `RequireSuperuserAuth()` | `RequireAdminAuth()` |
| `cmd/superuser.go` | 14 | `superuser` command tree | `owner` command tree |

### 2.3 Build Artifacts (ui/dist/)

| File | Matches | Status |
|------|---------|--------|
| `ui/dist/assets/index-DxrBioZt.js` | 17 | Rebuild required after source changes |
| `ui/dist/assets/docsCreate-CtxPHDTg.js` | 4 | Rebuild required after source changes |

---

## 3. Internal Compatibility (Keep As-Is)

These are internal implementation details that must NOT leak into public APIs:

| File | Legacy Terms | Status |
|------|--------------|--------|
| `core/record_model_superusers.go` | `CollectionNameSuperusers = "_superusers"`, `DefaultInstallerEmail` | **KEEP-INTERNAL** - Core data model |
| `core/backup_restore.go` | `SuperuserIPs` | **KEEP-INTERNAL** - Settings key |
| `core/backup_create.go` | `SuperuserIPs` | **KEEP-INTERNAL** - Settings key |
| `core/event_request.go` | `HasSuperuserAuth()`, `IsSuperuser()` | **KEEP-INTERNAL** - Internal methods |
| `core/system_alert.go` | `CollectionNameSuperusers` | **KEEP-INTERNAL** - System alerts |
| `core/settings_model.go` | `SuperuserIPs` | **KEEP-INTERNAL** - Settings model |
| `core/base.go` | `registerSuperuserHooks()`, `pb_data` | **KEEP-INTERNAL** - Bootstrap |
| `core/app.go` | `CollectionNameSuperusers` | **KEEP-INTERNAL** - App interface |
| `arsy/aeng_001_capability_inventory.go` | Inventory reference | **KEEP-INTERNAL** - Documentation |
| `arsy/aeng/security.go` | Reference to superuser | **KEEP-INTERNAL** - Internal enforcement |
| `arsy/httpapi/servers.go` | 10 matches - legacy API compat | **KEEP-INTERNAL** - Internal HTTP API |
| `arsy/httpapi/projects.go` | 10 matches - legacy API compat | **KEEP-INTERNAL** - Internal HTTP API |
| `arsy/httpapi/deployments.go` | 8 matches - legacy API compat | **KEEP-INTERNAL** - Internal HTTP API |
| `arsy/deployments/jobs.go` | 8 matches - legacy API compat | **KEEP-INTERNAL** - Internal HTTP API |
| `migrations/*.go` | Historical migration refs | **KEEP-INTERNAL** - Immutable history |

---

## 4. Allowlist (Legally Required / Historical)

| File | Reason |
|------|--------|
| `README.md` | Attribution notice required by PocketBase license |
| `CONTRIBUTING.md` | Historical context |
| `SECURITY.md` | Security policy history |
| `CHANGELOG_*.md` | Historical changelog entries |
| `LICENSE` files | Third-party license preservation |

---

## 5. Remediation Plan

### Phase 1: Core Terminology (User-Facing)
1. **UI Terminology** - Rename all 27+ UI source files
2. **API Middleware** - Rename middleware IDs in `apis/middlewares.go`
3. **Installer** - Rename `DefaultInstallerFunc` → `DefaultBootstrapFunc`, `pbinstall` → `arsyinstall`
4. **CLI** - Replace `superuser` command with `owner` command

### Phase 2: Internal Consistency (Optional)
1. Add `IsAdmin()` method to `Record` that aliases `IsSuperuser()` for public API
2. Add `HasAdminAuth()` that aliases `HasSuperuserAuth()` for public API
3. Keep internal `_superusers` collection name for data compatibility

### Phase 3: Verification
1. Rebuild UI (`cd ui && npm run build`)
2. Run brand scan again - should show only allowlisted/internal matches
3. Run all tests - ensure no regressions

---

## 6. Acceptance Criteria (ARSY-012)

- [ ] Brand scan passes for release artifacts and user-facing surfaces
- [ ] Every allowlisted match has a documented reason
- [ ] Screenshots/smoke tests verify A-R-S-Y identity in Studio, installer and CLI
- [ ] No unexplained legacy product surface remains

---

## 7. Evidence Location

- `docs/architecture/pocketbase-transformation/brand-scan-report.md`
- `docs/architecture/pocketbase-transformation/brand-scan-report.json`