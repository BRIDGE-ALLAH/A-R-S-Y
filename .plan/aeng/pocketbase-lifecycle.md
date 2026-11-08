# AENG-002: PocketBase App Lifecycle and Database Ownership

## Startup sequence (execution order)

1. **Entrypoint**: `examples/base/main.go` constructs `pocketbase.New()` (or `NewWithConfig`) from `github.com/bridge-allah/a-r-s-y`.
2. **App construction** (`pocketbase.go`, `core/base.go`): creates `BaseApp` with config (data dir, encryption key env handling), initializes logger, registers system migrations via blank imports / `core.SystemMigrations`.
3. **A-R-S-Y migration registration**: `arsy/migrations` package `init()` registers `baas_tenants`, `baas_tenant_memberships`, MVP1 collections (`arsy/migrations/registry.go`, `mvp1.go`, `baas_*.go`) into `core.SystemMigrations`. Import side-effects occur when packages are linked (e.g., via `arsy` imports in base app wiring).
4. **Bootstrap/serve**: `app.Start()` / `app.Bootstrap()` opens data dir, initializes SQLite connections (`data.db`, `auxiliary.db` patterns per PB), runs pending system migrations, loads settings, fires `OnBootstrap` hooks.
5. **HTTP server**: `apis.Serve()` / serve command registers routes, middlewares, static handlers; A-R-S-Y routes mounted via `arsy/httpapi.RegisterRoutes` / `RegisterARsyRoutes` (wired in app setup; see `arsy/httpapi/routes.go`).
6. **Shutdown**: serve context cancellation triggers graceful HTTP shutdown; `OnTerminate` hooks run; `BaseApp.ResetBootstrapState`/close paths release DB connections and cron/log resources. SQLite handles closed via `dbx`/`sql.DB`.

## Database ownership

| Resource | Owner | Opened | Closed |
|---|---|---|---|
| Control-plane SQLite (`pb_data/data.db`, aux) | `core.BaseApp` | Bootstrap | Terminate/reset; app owns connections. |
| Tenant SQLite files (`<tenantsRoot>/<id>.db`) | `arsy/tenancy.SQLiteProvider` | On `Open(ctx, tenantID)` / first use | `CloseTenant` eviction or `Provider.Close()`; per-tenant `sql.DB` with `MaxOpenConns(1)`. |
| BaaS directories | `arsy/config.EnsureBaaSDirectories` | Startup/config init | N/A (filesystem). |

## A-R-S-Y migration registration map

- Control-plane/system: `arsy/migrations/*.go` -> `core.SystemMigrations.Register(...)` with timestamped IDs (e.g., `20261107010000_baas_tenants.go`, `20261108150000_baas_tenant_memberships.go`).
- Execution: PB migration runner applies pending system migrations during bootstrap/`migrate up`.
- Tenant-local: not executed by PB system migrations; `arsy/tenancy/migration.go` (`MigrationRunner`) and `arsy/engine/bootstrap` apply schema to tenant DB handles.

## Multi-instance feasibility / risks

| Option | Feasibility | Risks / unknowns |
|---|---|---|
| Multiple full `PocketBase`/`BaseApp` instances in one process | Technically possible only with strict isolation (separate data dirs, ports, hook registries) | High risk: global hooks/settings collisions, migration state confusion, file locks, resource leaks, operational complexity. **Not recommended** for tenancy. |
| Separate tenant `sql.DB`/dbx handles to tenant SQLite files | **Feasible and implemented** (`arsy/tenancy.SQLiteProvider`) | Must enforce handle lifecycle (eviction on suspend/delete), single-writer pragmas, path safety, and never share handles across tenant IDs. |
| One PB app + row/collection scoping | Feasible for metadata, unsafe for strong isolation | Cross-tenant query bugs/leaks; not “dedicated database”. |

## Unknowns (explicit)
- Exact internal hook teardown ordering across all PB versions; rely on public `OnTerminate`/bootstrap contracts.
- Concurrent multi-process access to same tenant SQLite file (operator responsibility; WAL helps reads, single-writer constraints remain).

No production code changes. No unsafe concurrent app-instance tenancy proposed.
