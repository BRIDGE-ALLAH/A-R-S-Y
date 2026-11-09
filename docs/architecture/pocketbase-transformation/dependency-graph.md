# Runtime Entry Points & Package Dependency Graph

**Generated:** 2026-11-09  
**Task:** ARSY-001 — Establish the PocketBase transformation inventory  
**Component:** Design - Runtime entry points and dependency graph

---

## 1. Executable Startup Flow

```
main() [pocketbase.go:NewWithConfig()]
  -> pocketbase.go:NewWithConfig(Config)
      -> core.NewBaseApp(BaseAppConfig) [core/base.go]
          -> core.initDataDir()
          -> core.openDB() [core/db_connect.go]
          -> core.loadSettings() [core/settings_model.go]
          -> core.runMigrations() [core/migrations_runner.go]
      -> pocketbase.go:RootCmd.AddCommand(cmd.NewServeCommand)
      -> pocketbase.go:RootCmd.AddCommand(cmd.NewSuperuserCommand)
      -> pocketbase.go:Execute()
          -> pb.Bootstrap() [if not skipped]
          -> pb.RootCmd.Execute() [cobra]
              -> serve command [cmd/serve.go]
                  -> apis.Serve() [apis/serve.go]
                      -> http.Server.ListenAndServe()
                      -> apis.registerRoutes() [apis/*.go]
              -> superuser command [cmd/superuser.go]
```

---

## 2. Package Dependency Graph (Layered)

### Layer 0: External Dependencies
```
github.com/spf13/cobra           -> CLI framework
github.com/google/uuid           -> UUID generation
ganigeorgiev.com/pocketbase/dbx  -> SQL builder (UPSTREAM)
ganigeorgiev.com/pocketbase/ozzo-validation/v4 -> Validation (UPSTREAM)
modernc.org/sqlite               -> SQLite driver
github.com/fatih/color           -> CLI colors
github.com/bridge-allah/a-r-s-y/tools/* -> Internal tools
```

### Layer 1: Foundation (`tools/`)
```
tools/hook       -> Event system (no deps)
tools/list       -> Slice utilities (no deps)
tools/routine    -> Goroutine helpers (no deps)
tools/osutils    -> OS utilities (no deps)
tools/inflector  -> String inflection (no deps)
tools/types      -> Type utilities (no deps)
tools/picker     -> Random selection (no deps)
tools/store      -> In-memory store (no deps)
tools/logger     -> Structured logging (no deps)
tools/archive    -> Archive handling (no deps)
tools/template   -> Template rendering (no deps)
tools/search     -> Search utilities (no deps)
tools/cron       -> Cron scheduler (deps: tools/routine)
tools/security   -> Crypto/tokens (deps: tools/types)
tools/tokenizer  -> JWT tokens (deps: tools/security)
tools/filesystem -> Local/S3 FS (deps: tools/security)
tools/mailer     -> SMTP/Sendmail (deps: tools/filesystem)
tools/subscriptions -> Realtime (deps: tools/hook, tools/routine)
tools/router     -> Request routing (deps: tools/list)
tools/dbutils    -> DB utilities (deps: dbx)
tools/auth       -> Auth helpers (deps: tools/tokenizer, tools/security)
```

### Layer 2: Core Domain (`core/`)
```
core/db_connect.go        -> DB connection (deps: dbx, modernc.org/sqlite)
core/db.go                -> DB abstraction (deps: dbx, core/db_connect)
core/db_builder.go        -> Query builder (deps: dbx)
core/db_model.go          -> Base model (deps: dbx)
core/db_table.go          -> Table operations (deps: dbx)
core/db_tx.go             -> Transactions (deps: dbx)
core/db_retry.go          -> Retry logic (deps: core/db)
core/field*.go            -> Field types (22 files, deps: core/db_model, validation)
core/collection_model.go  -> Collections (deps: core/field*, core/db)
core/record_model.go      -> Records (deps: core/collection_model, core/field*)
core/record_query.go      -> Record queries (deps: core/record_model, dbx)
core/record_tokens.go     -> Auth tokens (deps: tools/tokenizer)
core/auth_origin*.go      -> Auth origins (deps: core/record_model)
core/external_auth*.go    -> External auth (deps: core/record_model)
core/mfa_model.go         -> MFA (deps: core/record_model)
core/otp_model.go         -> OTP (deps: core/record_model)
core/settings_model.go    -> App settings (deps: core/db, core/record_model)
core/migrations_runner.go -> Migrations (deps: core/db, migrations/*)
core/events.go            -> Event system (deps: tools/hook)
core/view.go              -> Views (deps: core/collection_model)
core/backup.go            -> Backups (deps: core/db, tools/filesystem)
core/notify_watcher.go    -> FS watcher (deps: tools/routine)
core/log_model.go         -> Logs (deps: core/db)
```

### Layer 3: API Layer (`apis/`)
```
apis/serve.go             -> HTTP server (deps: core.App, apis/*)
apis/base.go              -> Handler wrapping (deps: core.App)
apis/middlewares*.go      -> Auth, CORS, rate limit, gzip (deps: core.App, tools/*)
apis/collection*.go       -> Collection APIs (deps: core/collection_model)
apis/record_crud.go       -> Record CRUD (deps: core/record_model, core/record_query)
apis/record_helpers.go    -> Record helpers (deps: core/record_model)
apis/batch.go             -> Batch operations (deps: apis/record_crud)
apis/record_auth*.go      -> Auth APIs (20+ files, deps: core/record_model, core/auth_origin*, tools/tokenizer)
apis/installer.go         -> Installer (deps: core/record_model, ui.DistDirFS)
apis/backup.go            -> Backup APIs (deps: core/backup)
apis/file.go              -> File APIs (deps: tools/filesystem)
apis/realtime.go          -> Realtime (deps: core.SubscriptionsBroker)
apis/logs.go              -> Log APIs (deps: core/log_model)
apis/settings.go          -> Settings APIs (deps: core/settings_model)
apis/sql.go               -> SQL API (deps: core.DB)
apis/cron.go              -> Cron API (deps: core.Cron)
apis/health.go            -> Health (deps: core.App)
apis/extensions.go        -> Extensions (deps: core.App)
```

### Layer 4: UI Embedding (`ui/`)
```
ui/embed.go               -> embed.FS directive (deps: ui/dist/*)
ui/embed_no_ui.go         -> No-UI variant
vite build                -> Builds ui/src/ -> ui/dist/
ui/src/                   -> Vue 3 app (deps: pocketbase JS SDK)
```

### Layer 5: A-R-S-Y Extensions (`arsy/`)
```
arsy/aeng/                -> Application Engine (deps: core/*, tools/*)
arsy/tenancy/             -> Multi-tenancy (deps: core/*, tools/*, arsy/aeng)
arsy/engine/              -> Engine components (deps: core/*, tools/*, arsy/aeng)
arsy/config/              -> Config (deps: core/settings_model)
arsy/console/             -> Console (deps: core.App)
arsy/controlplane/        -> Control plane (deps: core.App, arsy/tenancy)
arsy/deployments/         -> Deployments (deps: core.App, tools/ssh)
arsy/httpapi/             -> HTTP API extensions (deps: apis/*)
arsy/migrations/          -> Custom migrations (deps: core/migrations_runner)
arsy/security/            -> Security (deps: core/*, tools/security)
arsy/servers/             -> Server management (deps: core/App, arsy/tenancy)
```

---

## 3. Critical Paths

### Path A: HTTP Request → Record Read
```
HTTP Request
  -> apis.Serve() [apis/serve.go]
  -> middlewares.Auth() [apis/middlewares.go]
      -> core.App.FindAuthRecordByToken() [core/record_model_auth.go]
  -> apis.RecordCrud.GET() [apis/record_crud.go]
      -> core.App.FindRecordById() [core/record_model.go]
          -> core.DB().Select()... [core/db.go -> dbx]
          -> modernc.org/sqlite
```

### Path B: Bootstrap → Migrations
```
pocketbase.go:Execute()
  -> pb.Bootstrap() [core/base.go]
      -> core.openDB() [core/db_connect.go]
      -> core.loadSettings() [core/settings_model.go]
      -> core.RunSystemMigrations() [core/migrations_runner.go]
          -> migrations/* (8 files)
      -> core.RunAppMigrations() [core/migrations_runner.go]
          -> arsy/migrations/*
```

### Path C: Installer Flow
```
serve command (no superuser)
  -> apis.Serve()
  -> apis.loadInstaller() [apis/installer.go]
      -> findOrCreateInstallerSuperuser() [core/record_model_superusers.go]
      -> DefaultInstallerFunc() [apis/installer.go]
          -> systemSuperuser.NewStaticAuthToken() [core/record_tokens.go]
          -> osutils.LaunchURL() [tools/osutils]
          -> UI: pageInstaller [ui/src/auth/pageInstaller.js]
```

---

## 4. Ownership Boundaries

| Domain | Owner Package | Public API |
|--------|--------------|------------|
| Database | `core/db*.go` | `core.App.DB()`, `ConcurrentDB()`, `NonconcurrentDB()` |
| Collections | `core/collection_model.go` | `core.App.FindCollectionByNameOrId()` |
| Records | `core/record_model.go` | `core.App.FindRecordById()`, `Save()`, `Delete()` |
| Auth | `core/record_model_auth.go` | `core.App.FindAuthRecordByEmail()`, `AuthWithPassword()` |
| Settings | `core/settings_model.go` | `core.App.Settings()`, `ReloadSettings()` |
| Migrations | `core/migrations_runner.go` | `core.App.RunSystemMigrations()`, `RunAppMigrations()` |
| Files | `tools/filesystem` | `core.App.NewFilesystem()`, `NewBackupsFilesystem()` |
| Realtime | `tools/subscriptions` | `core.App.SubscriptionsBroker()` |
| Tenancy | `arsy/tenancy` | `arsy/tenancy.Router`, `Provisioner` |
| Engine | `arsy/engine` | `arsy/engine/registry.Registry` |
| AENG | `arsy/aeng` | `arsy/aeng.ModelRegistry`, `MigrationExecutor`, `SeedRunner` |

---

## 5. Verification Methods

| Entry Point | Test Location | Verification |
|-------------|---------------|--------------|
| `pocketbase.go` | `pocketbase_test.go` | App creation, bootstrap, command registration |
| `cmd/serve.go` | `apis/serve_test.go` | HTTP server startup, route registration |
| `cmd/superuser.go` | `cmd/superuser_test.go` | Superuser CRUD, auth |
| `core/base.go` | `core/base_test.go` | Bootstrap, DB, settings, migrations |
| `apis/installer.go` | `apis/installer_test.go` | Installer flow, superuser creation |
| `ui/embed.go` | Build + smoke test | Vite build succeeds, assets embed |
| `arsy/tenancy` | `arsy/tenancy/*_test.go` | Provisioning, routing, isolation |
| `arsy/aeng` | `arsy/aeng/*_test.go` | Registry, migration, seed, security, audit |

---

## 6. Transformation Priority Order

1. **Terminology-only (ADAPT)**: `core/*.go`, `apis/*.go` - string replacements
2. **User-facing UI (REPLACE)**: `ui/src/auth/*.js` - rebuild with A-R-S-Y identity
3. **Installer (ADAPT+REPLACE)**: `apis/installer.go` + UI - rename route, terminology
4. **CLI commands (ADAPT)**: `cmd/*.go` - rename superuser → owner/admin
5. **Migrations (LICENSE)**: Preserve upstream, add A-R-S-Y migrations after
6. **Broken code (REMOVE)**: `arsy/aeng/audit_logger.go` - delete

---

**Evidence location:** `docs/architecture/pocketbase-transformation/dependency-graph.md`