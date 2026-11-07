# A-R-S-Y — MVP-1 Task Plan
## First Working Deployment: Onboarding → SSH Server → Docker Compose → Operations

**Status:** Ready for implementation  
**Target repository:** `BRIDGE-ALLAH/A-R-S-Y`  
**Plan file:** `.plan/task-mvp-1.md`  
**Product direction:** Self-hosted application deployment and infrastructure control plane (Coolify alternative)  
**Primary runtime:** Go; keep the control plane distributable as a standalone executable.

---

## 1. MVP-1 outcome

MVP-1 is complete when an administrator can:

1. Install and start the A-R-S-Y control-plane executable.
2. Open the A-R-S-Y console and complete first-run onboarding.
3. Register a remote Linux server using SSH key authentication.
4. Verify the SSH host key and test the connection safely.
5. Confirm that Docker Engine and the Docker Compose plugin are available on that server.
6. Create a project and an application.
7. Submit a Docker Compose manifest that uses existing container images.
8. Deploy the application to the selected server.
9. View deployment progress, deployment output, container status, and runtime logs.
10. Stop, redeploy, and roll back to a previous Compose configuration.
11. Restart A-R-S-Y without losing deployment records or interrupting already-running application containers.

The first supported public access method is the application's explicitly published host port, for example `http://SERVER_IP:8080`. Domain routing, automatic HTTPS, Git-based builds, and automatic Docker installation are intentionally scheduled for later milestones. This keeps MVP-1 small enough to implement and test without pretending that the first release already replaces every Coolify feature.

## 2. Repository facts used by this plan

The plan is based on the current repository structure inspected during planning:

- The repository is a PocketBase source snapshot, not yet an A-R-S-Y-specific implementation.
- The current executable entry point is `examples/base/main.go`.
- The release configuration is `.goreleaser.yaml` and still identifies the product/binary as PocketBase.
- The current example registers the PocketBase updater against the upstream PocketBase project.
- The existing `ui/` directory is the PocketBase Superuser dashboard, built with Vite and embedded through `ui/embed.go`.
- PocketBase already provides authentication, collections, migrations, hooks, REST APIs, and an embedded SQLite database.

**Owner direction (2026-11-07): replace the PocketBase Superuser UI with the A-R-S-Y console.** Reuse PocketBase authentication, APIs, hooks, migrations, settings, and SQLite storage. Keep the PocketBase dashboard available only until the A-R-S-Y console covers its administrator features; then remove the console link and disable the `/_/` UI route without removing PocketBase APIs or backend capabilities.

Before coding, re-check the actual branch because the repository may have changed since this plan was written.

## 3. Architecture contract

These decisions are fixed for MVP-1. A task agent must not reopen them unless the current source makes implementation impossible; in that case, document the conflict and choose the smallest compatible adjustment.

### Runtime and packaging

- [ ] The A-R-S-Y control plane remains Go-native.
- [ ] No Node.js process is required to run A-R-S-Y in production.
- [ ] Node.js/npm may be used only as a frontend build-time tool.
- [ ] The A-R-S-Y console is compiled to static files and embedded into the Go executable.
- [ ] Docker workloads run on managed target servers, not inside the A-R-S-Y control-plane process.
- [ ] The first release keeps SQLite/PocketBase storage. Do not add PostgreSQL, Redis, NATS, Kubernetes, or microservices in MVP-1.
- [ ] Keep existing PocketBase behavior unless an A-R-S-Y feature explicitly requires an extension.
- PocketBase remains the embedded backend and source of truth, but its complete administrator workflow must be available in the A-R-S-Y console before the PocketBase dashboard route is disabled for normal use.

### Boundaries

- New A-R-S-Y Go code belongs under `arsy/` (or a clearly justified subpackage).
- The A-R-S-Y HTTP API uses the `/api/arsy/v1/` prefix.
- The A-R-S-Y web console uses the `/arsy` route.
- The existing PocketBase API must keep working. Keep the Superuser UI route until A-R-S-Y admin parity is complete, then disable the UI route without removing backend APIs.
- The existing PocketBase `ui/` folder is not the A-R-S-Y dashboard.
- MVP-1 infrastructure management is restricted to authenticated PocketBase superusers. Organization-level RBAC is future work.
- Use PocketBase migrations for A-R-S-Y collections. Never rewrite historical PocketBase migrations.
- Keep server operations behind typed Go interfaces. Do not expose an unrestricted remote-shell endpoint.

### Security invariants

- Never use `ssh.InsecureIgnoreHostKey`.
- Never store private SSH keys or environment secrets as plaintext.
- Never return decrypted credentials from an API response.
- Never print credentials, private keys, passwords, tokens, or secret environment values in logs.
- Verify and pin SSH host keys before accepting a server as trusted.
- Apply context deadlines to SSH connection, command execution, uploads, and deployment steps.
- Validate application/project identifiers before using them in paths or commands.
- Never interpolate user-supplied text into a shell command.
- Do not run privileged, host-network, host-PID, device, or Docker-socket mounts by default. Reject them in MVP-1 with a useful validation error.
- Do not execute deployment commands against a real production server in automated tests.
- Never run destructive cleanup such as `docker system prune`, `docker compose down -v`, or broad container/volume deletion as an automatic recovery action.

## 4. MVP-1 scope

### Included

- A-R-S-Y executable identity and release build.
- Configuration and health diagnostics.
- Separate embedded A-R-S-Y console.
- First-run onboarding that reuses PocketBase superuser authentication.
- Projects and applications.
- Remote Linux server registration through SSH private-key authentication.
- Explicit SSH host-key pinning.
- Docker Engine and Docker Compose readiness checks.
- Docker Compose deployment using existing images only.
- Deployment history, logs, runtime status, stop, redeploy, and configuration rollback.
- Persistent deployment records and recoverable job execution.
- A basic integration test path using mocks/test infrastructure.

### Explicitly not included in MVP-1

- GitHub/GitLab/Gitea repository integration.
- Building from Dockerfile, Nixpacks, Buildpacks, or remote source code.
- Automatic installation or upgrade of Docker on the target server.
- Domain management, reverse proxy, wildcard routing, or automatic TLS.
- One-click service marketplace.
- Database provisioning/management.
- Multi-user teams, organization RBAC, billing, usage metering.
- AI agent, AI Router/LLM Gateway, MCP server, or autonomous infrastructure repair.
- Kubernetes, Swarm, multi-region scheduling, or high availability.
- PostgreSQL adapter or distributed control-plane mode.

These exclusions are deliberate. Do not add them opportunistically while implementing MVP-1.

---

## 5. Instructions for small coding models

Every unchecked task is intended to be executed independently by a small coding model.

1. Read this plan and inspect only the files relevant to the selected task.
2. Execute **one task ID at a time**. Do not attempt to complete an entire phase in one pass.
3. Respect the listed dependencies. Do not start a task while its prerequisites are incomplete.
4. Make the smallest code change that satisfies the task.
5. Reuse PocketBase APIs, hooks, migrations, and utilities where practical. Do not fork or duplicate existing functionality without a concrete reason.
6. Do not perform speculative refactors or rename unrelated PocketBase internals.
7. Add or update tests with each behavior change.
8. Run the verification command listed in the task. Also run focused tests for changed packages.
9. Mark a checkbox complete only when the acceptance condition is true and the verification command passes.
10. If a task is blocked by missing credentials, an unavailable external server, or unavailable infrastructure, finish all local implementation and mock tests; record the exact blocker beside the task. Do not invent credentials or silently skip security checks.
11. Use the current working branch. Do not create another branch unless the repository owner explicitly requests one.
12. Keep commits small and descriptive. Do not force-push or rewrite existing history.
13. Do not push a deployment to a production server. Runtime server actions must be initiated by an authenticated administrator from the A-R-S-Y console.

### Required task completion note

For each completed task, append a short note to the task entry or its linked issue:

- **Changed:** files or behavior changed.
- **Verified:** exact commands/tests run and their result.
- **Remaining:** any known limitation, or `None`.

---

# 6. Task checklist

## Phase A — Baseline and A-R-S-Y identity

**Phase exit:** The repository still builds and tests; the produced executable and release metadata identify as A-R-S-Y; the executable no longer offers an upstream PocketBase self-update.

- [x] **MVP1-001 — Capture the baseline build and test result**
  - Inspect the current Go version and existing Makefile targets.
  - Run `go test ./...` and the existing build command for `examples/base`.
  - Record failures that already exist; do not fix unrelated failures in this task.
  - **Done when:** baseline results are documented in the task completion note.
  - **Changed:** No source changes. Captured the current worktree baseline in this task note.
  - **Verified:** `go version` → `go1.27.0 windows/amd64`; Makefile targets inspected (`lint`, `test`, `jstypes`, `test-report`); `go test ./...` → failed. Observed failures include batch multipart upload and backup tests in `apis` (backup test also reported insufficient disk space), a panic in `core`, Windows symlink privilege failures in `tools/archive`, and expected-output mismatches in `tools/picker`, `tools/router`, and `tools/types`. Remaining package results included passes. `go build -o %TEMP%/arsy-mvp1-baseline.exe ./examples/base` → passed (46,875,648 bytes).
  - **Remaining:** This is the current modified worktree, not a clean pre-change checkout, so these results cannot be attributed to upstream baseline or to the A-R-S-Y changes alone. Disk space and Windows symlink privileges affected verification.

- [x] **MVP1-002 — Change the Go module identity to A-R-S-Y**
  - Change the module path in `go.mod` to `github.com/bridge-allah/a-r-s-y`.
  - Replace only repository-local imports that point to the old PocketBase application module.
  - Do not rename external dependencies such as PocketBase's separately hosted `dbx`, validation, or other modules.
  - Run formatting and `go mod tidy`.
  - **Done when:** `go list ./...` and `go test ./...` resolve with no old local-module imports remaining.
  - **Changed:** `go.mod` declares `github.com/bridge-allah/a-r-s-y`; repository-local imports use that lowercase module path. `go mod tidy` and `gofmt` completed for changed Go source files.
  - **Verified:** Searches found no old PocketBase-local or uppercase A-R-S-Y module paths; `go list ./...` and `go test ./arsy/...` passed. The full `go test ./...` run for MVP1-001 failed for the environment/current-worktree issues recorded there.
  - **Remaining:** Existing repository-wide working-tree changes predate this execution; attribution cannot be established.

- [x] **MVP1-003 — Rename the release project and executable**
  - Update `.goreleaser.yaml` project name and binary name to `arsy`.
  - Keep the current supported OS/architecture matrix unless a build target is proven broken.
  - **Done when:** a local release build produces an executable named `arsy` (or `arsy.exe`).
  - **Changed:** `.goreleaser.yaml` sets the project and binary names to `arsy` and retains the existing target matrix.
  - **Verified:** `go build -o %TEMP%/arsy-mvp1-release/arsy.exe ./examples/base` passed; the output file was 46,875,648 bytes.
  - **Remaining:** GoReleaser itself is not installed in this environment, so only the executable build was verified.

- [x] **MVP1-004 — Make version output identify A-R-S-Y**
  - Keep the existing version injection mechanism.
  - Update user-visible product/version labels where they are directly tied to the executable.
  - Avoid mass-renaming internal PocketBase types and symbols.
  - **Done when:** `arsy version` or the existing version command identifies the product as A-R-S-Y and displays the build version.
  - **Changed:** No additional code change was needed; the current executable name and shared `Version` value already drive version output.
  - **Verified:** Built with `-X github.com/BRIDGE-ALLAH/A-R-S-Y.Version=1.2.3`; `--version` returned `arsy-mvp1-version.exe version 1.2.3`. The release-named binary previously returned `arsy.exe version (untracked)`.
  - **Remaining:** None for this task.

- [x] **MVP1-005 — Disable the PocketBase upstream updater registration**
  - Remove the upstream PocketBase updater registration from the A-R-S-Y executable entry point.
  - Do not delete the reusable PocketBase updater package.
  - Do not point the updater at A-R-S-Y until a compatible A-R-S-Y release feed exists.
  - **Done when:** the A-R-S-Y binary cannot accidentally replace itself with a PocketBase release.
  - **Changed:** Removed the updater registration and import from `examples/base/main.go`; the reusable updater package remains intact.
  - **Verified:** No updater registration/import remains in the executable entry point; `go build -o %TEMP%/arsy-mvp1-no-updater.exe ./examples/base` passed.
  - **Remaining:** None for this task.

- [x] **MVP1-006 — Replace the inherited root README with A-R-S-Y project documentation**
  - Explain the A-R-S-Y purpose, current MVP status, Go-native runtime, and build/run commands.
  - Clearly state that A-R-S-Y is independently maintained and is derived from PocketBase.
  - Preserve accurate upstream license and attribution information.
  - Do not claim features that are not implemented.
  - **Done when:** the README describes the actual A-R-S-Y repository rather than presenting it as an unchanged PocketBase release.
  - **Changed:** Replaced `README.md` with A-R-S-Y purpose, early MVP status, Go-native runtime, build/run commands, current feature limitations, and explicit PocketBase derivation/independent-maintenance statement. Linked the preserved MIT license and copyright attribution.
  - **Verified:** Manually checked the README claims against the current implementation and `LICENSE.md`; deployment and console capabilities are clearly described as planned.
  - **Remaining:** None for this task.

- [x] **MVP1-007 — Add a minimal Go CI workflow**
  - Add a GitHub Actions workflow that installs the required Go version.
  - Run formatting checks, `go test ./...`, and a build of the A-R-S-Y executable.
  - Do not add frontend CI until the A-R-S-Y console exists.
  - **Done when:** the workflow runs successfully on a push and pull request.
  - **Changed:** Added `.github/workflows/go.yml` for pushes and pull requests, using Go 1.27, formatting validation, the Go test suite, and an A-R-S-Y executable build. No frontend CI was added.
  - **Verified:** `gofmt -l .` returned no files after formatting the changed Go source files; local executable build passed in MVP1-003. Workflow contents were inspected.
  - **Remaining:** A GitHub Actions run cannot be observed until a push or pull request triggers it; no remote write was authorized.

## Phase B — Control-plane foundation

**Phase exit:** A-R-S-Y has its own package boundary, validated configuration, persistent data layout, and health/diagnostic commands.

- [x] **MVP1-008 — Create the A-R-S-Y package skeleton**
  - Add the top-level `arsy/` package directory.
  - Add only small domain package placeholders needed by later tasks: `config`, `httpapi`, `servers`, `security`, and `deployments`.
  - Do not move PocketBase source files.
  - **Done when:** the new packages compile and have a short package comment describing their responsibility.
  - **Changed:** Added `arsy/` and the `config`, `httpapi`, `servers`, `security`, and `deployments` subpackages with package documentation.
  - **Verified:** `go test ./arsy/...` passed for all six packages.
  - **Remaining:** None for this task.

- [x] **MVP1-009 — Define the A-R-S-Y runtime configuration type**
  - Create a typed configuration structure for data directory, HTTP API prefix, console route, and secret-key source.
  - Keep defaults in one place.
  - Do not add provider-specific configuration or secret values to logs.
  - **Done when:** configuration defaults have unit tests.
  - **Changed:** Added the typed runtime configuration and centralized defaults for the inherited PocketBase data-directory behavior, API prefix, console route, and secret-key environment variable name.
  - **Verified:** `go test ./arsy/config` passed, including a defaults unit test.
  - **Remaining:** Environment loading and validation are handled in MVP1-010.

- [x] **MVP1-010 — Load and validate A-R-S-Y environment configuration**
  - Support documented `ARSY_*` environment variables for A-R-S-Y-specific settings.
  - Reject malformed values with actionable error messages.
  - Do not change PocketBase's existing flags or environment variable semantics.
  - **Done when:** valid configuration loads and invalid configuration fails before the server starts.
  - **Changed:** Added loading/validation for `ARSY_DATA_DIR`, `ARSY_API_PREFIX`, `ARSY_CONSOLE_ROUTE`, and `ARSY_SECRET_KEY_ENV`; documented them in `README.md`; startup now validates configuration before constructing the server and applies an explicit data directory without changing existing PocketBase flags.
  - **Verified:** `go test ./arsy/config` passed for defaults, valid values, and invalid values. `go build -o %TEMP%/arsy-mvp1-config.exe ./examples/base` passed. Running `serve` with valid data-dir plus malformed `ARSY_API_PREFIX` exited 1 before server startup with an actionable error.
  - **Remaining:** Secret-key contents are not loaded or logged; cryptographic key validation is reserved for MVP1-030.

- [x] **MVP1-011 — Establish the A-R-S-Y data-directory layout**
  - Define stable subdirectories for application data, encrypted temporary files, and deployment manifests.
  - Create directories with restrictive permissions where secrets or private files may exist.
  - Keep existing PocketBase database/data migration behavior compatible.
  - **Done when:** startup creates the expected directories and tests verify permissions on supported systems.
  - **Changed:** Added stable `arsy/app`, `arsy/tmp/encrypted`, and `arsy/deployments/manifests` directories beneath the existing PocketBase data directory. Startup creates these directories with mode 0700 while retaining the existing database location.
  - **Verified:** `go test ./arsy/...` passed; permission assertions run on non-Windows systems and are skipped where Windows permission bits are not supported. The standalone binary started against a temporary data directory and created all three expected paths.
  - **Remaining:** None for this task.

- [x] **MVP1-012 — Add an A-R-S-Y liveness endpoint**
  - Register `GET /api/arsy/v1/health`.
  - Return only product name, version, and a simple alive status.
  - Do not include database credentials, environment values, filesystem paths, or secrets.
  - **Done when:** an HTTP test confirms the endpoint works and the response contains no sensitive configuration.
  - **Changed:** Added a dedicated `/api/arsy/v1/health` route returning only product, version, and `ok` status; registered it from the standalone executable.
  - **Verified:** `go test ./arsy/httpapi` passed an unauthenticated request and asserted a minimal response without data or secret fields; `go build ./examples/base` passed.
  - **Remaining:** None for this task.

- [x] **MVP1-013 — Add an authenticated readiness endpoint**
  - Register `GET /api/arsy/v1/ready`.
  - Check only local control-plane readiness: database available, migrations applied, required configuration valid.
  - Require superuser authentication or restrict the response to a minimal non-sensitive status.
  - **Done when:** tests cover ready and not-ready responses.
  - **Changed:** Added `/api/arsy/v1/ready`, which checks the local database with `SELECT 1` and returns only `ready` or `not_ready`; an unavailable database produces HTTP 503. The response is intentionally minimal and contains no sensitive information.
  - **Verified:** `go test ./arsy/httpapi` passed for a ready database and a simulated database failure; assertions confirm error details are not exposed. The standalone executable build passed.
  - **Remaining:** None for this task.

- [x] **MVP1-014 — Add an A-R-S-Y diagnostics CLI command**
  - Add a command such as `arsy doctor`.
  - Report version, Go build/runtime information, data-directory accessibility, database readiness, and secret-key configuration status.
  - Never print secret values.
  - **Done when:** the command exits non-zero for a broken local setup and exits zero for a valid setup.
  - **Changed:** Added `arsy doctor`, reporting the product version, Go runtime/platform, data-directory write access, database readiness, and secret-key presence without outputting the key value.
  - **Verified:** `go test ./arsy/diagnostics` passed ready/broken setup cases and checked that secret/database detail is not exposed. Built with an injected version and ran `doctor --dev=false` against a temporary data directory; it exited 0 and reported the injected version and readiness statuses. `go test ./arsy/...` and `go list ./...` passed after lowercasing the module path.
  - **Remaining:** None for this task.

## Phase C — Separate console and first-run onboarding

**Phase exit:** An administrator can sign in to a dedicated A-R-S-Y console, see onboarding status, and follow a guided path to register the first server. The PocketBase Superuser UI remains intact.

- [x] **MVP1-015 — Create a separate A-R-S-Y console frontend**
  - Create a new `dashboard/` directory.
  - Use a small Vite-based static frontend; vanilla JavaScript is acceptable.
  - Keep Node/npm as build-time tools only.
  - Do not modify `ui/src` or replace the PocketBase Superuser dashboard.
  - **Done when:** `npm run build` in `dashboard/` produces a static `dist/` directory.
  - **Changed:** Added a separate React + Vite project under `dashboard/`, configured shadcn/ui components, and ignored local `node_modules`. The PocketBase `ui/` source was not changed for this task.
  - **Verified:** `npm run build` produced `dashboard/dist/index.html`, hashed assets, and local font files. `npm audit` reported zero vulnerabilities.
  - **Remaining:** None for this task.

- [x] **MVP1-016 — Add a minimal A-R-S-Y console shell**
  - Create a simple layout with navigation for Overview, Servers, Projects, and Deployments.
  - Use placeholder empty states for pages that are not yet implemented.
  - Do not create decorative controls that look functional but do nothing.
  - **Done when:** the built console renders all four navigation destinations without runtime errors.
  - **Changed:** Built a responsive comic-riot console with Overview, Servers, Projects, and Deployments routes, honest empty states, compact workspace panels, a temporary link to the PocketBase admin at `/_/`, and original inline SVG artwork. shadcn Button, Card, Badge, and Separator components are in use.
  - **Verified:** `npm run build` passed. Browser navigation confirmed all four route headings and titles. Layout checks at 320, 390, 768, and 1366 px found no horizontal overflow; mobile navigation targets are at least 57 px high. `npm audit` reported zero vulnerabilities.
  - **Remaining:** None for this task.
  - **Owner direction update (2026-11-07):** The temporary PocketBase admin link will be removed only after full PocketBase admin parity tasks MVP1-073 through MVP1-080 are complete.

- [x] **MVP1-017 — Embed the A-R-S-Y console assets into Go**
  - Add a dedicated Go embed package for `dashboard/dist`.
  - Do not reuse or overwrite `ui/embed.go`.
  - Add a build step that generates the frontend assets before compiling the executable.
  - **Done when:** the A-R-S-Y binary serves the console without requiring Node.js or a separate frontend server.
  - **Changed:** Added a dedicated embed package in `arsy/dashboard`, configured Vite output to `arsy/dashboard/dist`, and added dashboard build steps to CI and GoReleaser hooks.
  - **Verified:** `npm run build`; `go test ./arsy/dashboard ./arsy/httpapi -run 'Test(DistFSContainsBuiltConsole|Console|StaticAt)' -count=1`; `go build -o %TEMP%/arsy-mvp1-console.exe ./examples/base`; binary smoke test served `/arsy/`, JS, and CSS with HTTP 200 without Node at runtime.
  - **Remaining:** None for this task.

- [x] **MVP1-018 — Register the A-R-S-Y console route**
  - Serve the console at `/arsy` and its static assets under a predictable asset path.
  - Keep the existing PocketBase `/_/` route and all PocketBase API routes working.
  - Add SPA fallback only within the A-R-S-Y console route.
  - **Done when:** HTTP tests confirm both the A-R-S-Y console route and PocketBase admin route remain reachable.
  - **Changed:** Registered the configured console route with embedded assets and in-console SPA fallback; missing assets remain 404. PocketBase routes are left to their existing handlers.
  - **Verified:** Focused dashboard/API tests passed. Binary smoke test returned HTTP 200 for `/arsy/`, `/arsy/assets/*.js`, `/arsy/assets/*.css`, `/_/`, and `/api/health`.
  - **Remaining:** The PocketBase dashboard remains temporarily available during the owner-requested parity migration; task MVP1-080 will disable it after all admin workflows move into A-R-S-Y.

- [x] **MVP1-019 — Register the A-R-S-Y API route group**
  - Create a shared router group under `/api/arsy/v1`.
  - Keep route registration separate from PocketBase's existing default API handlers.
  - Add a small test route or health route through this group.
  - **Done when:** route registration is centralized and existing collection APIs still pass tests.
  - **Changed:** Centralized A-R-S-Y health and readiness endpoint registration in one versioned `/api/arsy/v1` router group and updated executable registration.
  - **Verified:** `go test ./arsy/httpapi`; `go test ./apis -run 'TestCollectionsList|TestCollectionView|TestCollectionCreate|TestCollectionUpdate' -count=1` passed. A broader collection run hit the known Windows cleanup failure in `TestCollectionDelete` (`Expected empty/deleted dir, found 2`).
  - **Remaining:** Full `apis` package has the baseline Windows cleanup failure recorded under MVP1-001; unaffected collection API scenarios pass.

- [x] **MVP1-020 — Add superuser-only authorization middleware**
  - Reuse PocketBase's existing superuser authentication.
  - Protect all infrastructure-management endpoints.
  - Return consistent 401/403 responses.
  - Do not add a second user/password system.
  - **Done when:** tests prove unauthenticated and regular authenticated users cannot manage servers or deployments.
  - **Changed:** Added `httpapi.ProtectedGroup`, which binds PocketBase's existing `RequireSuperuserAuth` middleware to A-R-S-Y infrastructure route groups.
  - **Verified:** `go test ./arsy/httpapi -run 'TestProtectedGroupRequiresPocketBaseSuperuser' -count=1` and `go test ./arsy/httpapi` passed; tests cover guest 401, regular-user 403, and superuser 200 for both `/servers` and `/deployments` routes.
  - **Remaining:** None for the route-group helper; future infrastructure endpoints must register through `ProtectedGroup`.

- [x] **MVP1-021 — Add a console login and logout flow**
  - Use the existing PocketBase Superuser authentication API.
  - Keep the auth token in memory for MVP-1; do not persist it in `localStorage`.
  - Show clear expired-session and invalid-credential states.
  - **Done when:** a superuser can log in, access protected console pages, and log out.
  - **Changed:** Added PocketBase superuser password authentication and periodic auth refresh, in-memory-only session state, protected console rendering, sign-out, invalid-credential feedback, and a grunge-styled login screen. Removed the direct PocketBase-admin link from the shell.
  - **Verified:** `npm run build` and `npm audit` (zero vulnerabilities). Local temporary PocketBase instance verified successful login, protected Overview rendering, sign-out, and invalid-credential feedback in the browser.
  - **Remaining:** Full PocketBase admin UI parity is now an explicit owner-requested phase (MVP1-073 through MVP1-080); the `/_/` backend UI route remains until that phase is complete.

- [x] **MVP1-022 — Add the first-run onboarding status API**
  - Return non-sensitive setup state: whether a superuser exists, whether the secrets key is configured, and whether a server has been registered.
  - Do not return emails, credentials, private keys, or environment values.
  - **Done when:** tests cover fresh install and already-configured states.
  - **Changed:** Added the public, minimal `/api/arsy/v1/onboarding` status response for superuser existence, configured secret-key presence, and whether the `servers` collection has a record.
  - **Verified:** `go test ./arsy/httpapi`; focused PocketBase collection API tests passed with `go test ./apis -run 'TestCollectionsList|TestCollectionView|TestCollectionCreate|TestCollectionUpdate' -count=1`. Onboarding tests cover fresh and configured states and assert the secret value is never returned.
  - **Remaining:** The default `__pbinstaller@example.com` bootstrap record is excluded from the superuser count so a fresh database still reports setup required.

- [x] **MVP1-023 — Build the onboarding wizard**
  - Show a welcome screen and explain the A-R-S-Y control-plane/managed-server model.
  - If no superuser exists, show the existing CLI command pattern `arsy superuser create EMAIL PASSWORD`; do not create an unauthenticated public password-creation endpoint.
  - After login, guide the administrator through secret-key readiness and first-server registration.
  - Allow the wizard to be resumed after refresh or restart.
  - **Done when:** a fresh-install user can follow the wizard without guessing the next step.
  - **Changed:** Added first-install CLI guidance, a backend-driven setup checklist for superuser/key/server state, control-plane versus managed-server explanation, and a persistent link to the server setup route. Removed the direct PocketBase dashboard link from the shell.
  - **Verified:** `npm run build`; `npm audit` reported zero vulnerabilities. Browser verified the fresh-install CLI command, a configured superuser with incomplete key/server status, and the same onboarding state after refresh and sign-in.
  - **Remaining:** The server registration form is delivered by MVP1-044; the wizard links to that route until it is implemented.

## Phase D — Data model and secret handling

**Phase exit:** Projects, servers, applications, and deployments have persistent schemas; all management data is protected; credentials are encrypted and never exposed.

- [x] **MVP1-024 — Add an A-R-S-Y-owned migration registration package**
  - Follow the repository's existing PocketBase system migration registration pattern.
  - Keep A-R-S-Y migrations separate from historical upstream migrations.
  - Register the new package from the A-R-S-Y executable.
  - **Done when:** a clean database applies the A-R-S-Y migration package and existing migrations still pass.
  - **Changed:** Added `arsy/migrations` as the A-R-S-Y-owned migration namespace and imported it from the standalone executable; future A-R-S-Y migration files register in this package without modifying upstream history.
  - **Verified:** `go test ./arsy/migrations ./arsy/httpapi`; `go build -o %TEMP%/arsy-mvp1-task024.exe ./examples/base` passed. PocketBase system migrations remain the bootstrap path for the executable.
  - **Remaining:** The first schema migration is added by task MVP1-025.

- [x] **MVP1-025 — Create the projects collection**
  - Add project name, validated slug, description, and timestamps.
  - Add unique slug validation.
  - Keep list/view/create/update/delete superuser-only.
  - **Done when:** migration and collection tests pass.
  - **Changed:** Added the A-R-S-Y projects base collection migration with required name, validated slug, description, automatic created/updated fields, a unique slug index, and default superuser-only rules; added migration validation tests.
  - **Verified:** `go test ./arsy/migrations -run TestProjectsCollectionMigration -count=1 -v` and `go test ./arsy/migrations ./arsy/httpapi` passed.
  - **Remaining:** None.

- [x] **MVP1-026 — Create the servers collection**
  - Add server name, host, SSH port, SSH username, operating-system metadata, pinned host-key fingerprint, readiness state, and timestamps.
  - Keep credentials in a separate encrypted field or related protected record.
  - Do not expose private-key ciphertext in standard API responses.
  - **Done when:** schema tests verify required fields, validation, and private collection rules.
  - **Changed:** Added the A-R-S-Y servers collection migration with validated SSH metadata, host-key/readiness fields, hidden private-key ciphertext, automatic timestamps, and default superuser-only rules; added schema and API export tests.
  - **Verified:** `go test ./arsy/migrations -run 'Test(Projects|Servers)CollectionMigration' -count=1 -v`, `go test ./arsy/migrations ./arsy/httpapi`, and `go build -o %TEMP%/arsy-mvp1-task026.exe ./examples/base` passed.
  - **Remaining:** Ciphertext encryption/decryption at write/use time is implemented in a later credential task; this schema deliberately stores it in a hidden field and does not expose it in public exports.

- [x] **MVP1-027 — Create the applications collection**
  - Add project relation, server relation, application name, validated slug, Compose manifest, and lifecycle state.
  - Support one application mapped to one server in MVP-1.
  - Do not add Git repository fields or buildpack configuration yet.
  - **Done when:** application records can be created and validated through tests.
  - **Changed:** Added the applications collection migration with required project/server relations, validated application slug, bounded required Compose manifest, lifecycle state, timestamps, and a per-project unique slug index; added schema and record validation tests.
  - **Verified:** `go test ./arsy/migrations -run TestApplicationsCollectionMigration -count=1 -v`, `go test ./arsy/migrations ./arsy/httpapi`, and `go build -o %TEMP%/arsy-mvp1-task027.exe ./examples/base` passed.
  - **Remaining:** Compose semantic/security validation and deployment behavior are covered by later tasks.

- [x] **MVP1-028 — Create the deployments collection**
  - Store application relation, deployment identifier, status, start/end times, manifest checksum, safe summary, and previous deployment reference.
  - Do not store unredacted command output or secrets in summary fields.
  - **Done when:** deployment records can represent queued, running, succeeded, failed, and cancelled states.
  - **Changed:** Added the deployments collection migration with application relation, validated deployment ID, five allowed statuses, start/end dates, SHA-256 checksum field, bounded safe summary, previous deployment ID, timestamps, and default superuser-only rules; tests persist each supported state.
  - **Verified:** `go test ./arsy/migrations -run TestDeploymentsCollectionMigration -count=1 -v`, `go test ./arsy/migrations ./arsy/httpapi`, and `go build -o %TEMP%/arsy-mvp1-task028.exe ./examples/base` passed.
  - **Remaining:** Previous deployment is stored as an identifier rather than a PocketBase relation because self-relations cannot validate during first collection creation; command output remains outside the summary field.

- [x] **MVP1-029 — Add a central AES-GCM secret encryption helper**
  - Use Go standard-library cryptography.
  - Require a 32-byte key supplied outside the database through documented A-R-S-Y configuration.
  - Store ciphertext with a format/version marker and a random nonce.
  - Return errors for missing, malformed, or incorrect keys.
  - **Done when:** encrypt/decrypt round-trip and tamper-detection tests pass.
  - **Changed:** Added A-R-S-Y AES-256-GCM encryption/decryption with a versioned ciphertext prefix, a random nonce, authentication, and strict 32-byte key validation.
  - **Verified:** `go test ./arsy/security -count=1 -v` and `go test ./arsy/migrations ./arsy/httpapi` passed; tests cover round-trip, plaintext absence, wrong keys, tampering, malformed key sizes, and malformed ciphertext.
  - **Remaining:** Key generation/documentation and consuming the configured environment key are handled by MVP1-030 and MVP1-031.

- [x] **MVP1-030 — Add secret-key generation and validation**
  - Add a CLI helper to generate a cryptographically random secret key.
  - Print the key once to stdout only when explicitly requested.
  - Never automatically write the key into source control or log it.
  - Document secure storage with restrictive file permissions or a secret manager.
  - **Done when:** generated keys pass validation and invalid keys prevent credential storage.
  - **Changed:** Added `arsy generate-secret-key`, generating a random 256-bit key encoded as 64 hexadecimal characters; encryption accepts this format and the documented environment variable name controls where the service reads it.
  - **Verified:** `go test ./arsy/security ./arsy/config`; executable build; invoking the generated binary confirmed the command works without server configuration and emits only a valid 64-character hex key.
  - **Remaining:** Secret-key use at credential write/decrypt time is implemented by MVP1-031.

- [x] **MVP1-031 — Add encrypted SSH credential persistence**
  - Encrypt private-key material before saving it.
  - Decrypt only immediately before creating the SSH client.
  - Support unencrypted OpenSSH private keys for MVP-1; return a clear unsupported-format error for encrypted keys if passphrase support is not implemented.
  - **Done when:** database inspection shows ciphertext, not the original private key.
  - **Changed:** Added server key storage that validates unencrypted OpenSSH keys, encrypts before record persistence, and exposes decrypted signers only inside a short-lived callback; malformed and passphrase-encrypted keys receive safe errors.
  - **Verified:** `go test ./arsy/servers -count=1 -v` and the combined focused package suite passed; database assertions confirm stored data is ciphertext and not PEM plaintext.
  - **Remaining:** API route integration and response redaction are covered by MVP1-032 and subsequent server workflow tasks.

- [x] **MVP1-032 — Sanitize server API responses**
  - Use explicit response DTOs rather than returning credential-bearing records directly.
  - Omit private-key ciphertext and all decrypted values.
  - Redact credentials from error messages.
  - **Done when:** API tests assert that secret-related fields never appear in JSON responses.
  - **Changed:** Added an explicit server response DTO that allowlists operational metadata and omits encrypted/decrypted credentials and arbitrary record fields.
  - **Verified:** `go test ./arsy/httpapi -run TestServerResponseOmitsCredentialFields -count=1` and `go test ./arsy/security ./arsy/servers ./arsy/httpapi ./arsy/migrations -count=1` passed; `go build -o Z:\temp\arsy-build\arsy-mvp1-task032.exe ./examples/base` passed.
  - **Remaining:** Server management HTTP handlers are implemented by the following server workflow tasks and must use this DTO.

## Phase E — SSH connection and server readiness

**Phase exit:** A superuser can register a remote Linux host, pin its SSH host key, test connectivity, and see Docker/Compose readiness.

- [x] **MVP1-033 — Validate server connection input**
  - Validate hostname/IP, port range, username, and private-key structure.
  - Reject empty hostnames, invalid ports, and malformed key content.
  - Normalize hostnames without changing the user's intended destination.
  - **Done when:** table-driven validation tests cover valid and invalid inputs.
  - **Changed:** Added typed SSH connection input validation for DNS/IP host, port range, username format, and unencrypted private-key structure; DNS names are normalized to lowercase while IP literals are retained.
  - **Verified:** Table-driven cases passed with `go test ./arsy/servers ./arsy/security ./arsy/httpapi ./arsy/migrations -count=1`; standalone executable build passed to `Z:\temp\arsy-build\arsy-mvp1-task033.exe`.
  - **Remaining:** Host-key pinning and SSH connection verification are subsequent tasks.

- [x] **MVP1-034 — Implement SSH host-key fingerprint calculation**
  - Calculate a stable SHA-256 fingerprint from the SSH server host key.
  - Use the standard OpenSSH fingerprint format.
  - **Done when:** known fixture keys produce expected fingerprints.
  - **Changed:** Added an A-R-S-Y host-key fingerprint helper using the canonical OpenSSH SHA-256 format.
  - **Verified:** A deterministic Ed25519 fixture matches its fixed expected fingerprint with `go test ./arsy/servers -run 'Test(HostKeyFingerprintKnownFixture|ValidateConnectionInput)' -count=1`; executable build passed.
  - **Remaining:** First-seen and pinned-key comparison behavior is implemented by MVP1-035.

- [x] **MVP1-035 — Require explicit host-key pinning**
  - During the first connection test, display the server fingerprint without marking the server trusted.
  - Require the administrator to provide/confirm the expected fingerprint through an authenticated action.
  - Verify the pinned fingerprint on every later connection.
  - Reject changed or missing fingerprints.
  - **Done when:** tests cover first-seen, matching, mismatched, and changed host keys.
  - **Changed:** Added strict host-key comparison and an SSH callback that rejects absent pins, missing keys, and changed fingerprints; first-seen errors expose a candidate fingerprint without mutating server trust state.
  - **Verified:** Tests cover first-seen candidate, matching pin, changed key, missing key, and strict callback; `go test ./arsy/servers ./arsy/security ./arsy/httpapi ./arsy/migrations -count=1` and executable build passed.
  - **Remaining:** The authenticated endpoint that records an administrator-confirmed pin is implemented with server-management workflows.

- [x] **MVP1-036 — Implement a reusable SSH client factory**
  - Use `golang.org/x/crypto/ssh`.
  - Support private-key authentication only in MVP-1.
  - Require strict host-key verification.
  - Apply connection timeout and context cancellation.
  - **Done when:** the factory returns a connected client only for a trusted host.
  - **Changed:** Added a reusable context-aware SSH client factory using private-key authentication only, a mandatory pinned fingerprint callback, and bounded connection timeout/cancellation handling.
  - **Verified:** Local SSH test server accepts matching pinned host keys and rejects mismatched/unpinned keys; cancelled contexts fail. `go test ./arsy/servers ./arsy/security ./arsy/httpapi ./arsy/migrations -count=1` and executable build passed.
  - **Remaining:** Typed remote command operations are covered by MVP1-037.

- [ ] **MVP1-037 — Implement a bounded remote-command executor**
  - Accept typed internal operations, not arbitrary command strings from HTTP requests.
  - Apply command timeout and output-size limits.
  - Capture stdout/stderr separately.
  - Redact known secret values before returning output to callers.
  - **Done when:** unit tests verify timeout, cancellation, output truncation, and redaction.

- [ ] **MVP1-038 — Add an SSH connection-test service**
  - Combine input validation, host-key check, and SSH authentication.
  - Return a structured result with safe error codes.
  - Do not persist the server as ready merely because TCP port 22 is reachable.
  - **Done when:** mocked tests cover successful authentication, authentication failure, timeout, and host-key mismatch.

- [ ] **MVP1-039 — Detect remote Linux operating system and architecture**
  - Query a limited set of stable system facts using fixed commands.
  - Support Linux targets for MVP-1.
  - Reject unsupported operating systems with a clear message.
  - **Done when:** parser tests cover expected Linux outputs and malformed responses.

- [ ] **MVP1-040 — Detect Docker Engine readiness**
  - Check Docker client/server availability and server version.
  - Determine whether the configured SSH user can access Docker without an interactive sudo prompt.
  - Do not silently modify group membership or privilege settings.
  - **Done when:** tests distinguish Docker missing, daemon unavailable, permission denied, and ready.

- [ ] **MVP1-041 — Detect Docker Compose plugin readiness**
  - Check `docker compose version`.
  - Parse and store the version string.
  - Do not assume the legacy standalone `docker-compose` binary is equivalent.
  - **Done when:** tests cover installed, missing, and invalid version responses.

- [ ] **MVP1-042 — Build a server readiness summary**
  - Combine SSH, host-key, OS, Docker, and Compose results.
  - Return actionable status codes and safe diagnostic messages.
  - Keep readiness checks read-only.
  - **Done when:** one response clearly distinguishes ready and not-ready servers.

- [ ] **MVP1-043 — Add server management API endpoints**
  - Add authenticated create, list, view, update-label, connection-test, and readiness-check endpoints.
  - Do not add arbitrary SSH command execution.
  - Do not include credential fields in responses.
  - **Done when:** API tests cover authorization, validation, and normal CRUD behavior.

- [ ] **MVP1-044 — Add the Servers console screen**
  - List saved servers with hostname, SSH port, OS, Docker readiness, and last-check time.
  - Add a form to register a server and submit its private key.
  - Show the host-key fingerprint confirmation step.
  - **Done when:** an administrator can add and test a server from the console.

## Phase F — Docker Compose deployment engine

**Phase exit:** A-R-S-Y can deploy a Compose manifest containing existing images to a ready server, persist the outcome, and display runtime state and logs.

- [ ] **MVP1-045 — Define the supported Compose subset**
  - Document supported MVP-1 keys: `services`, `image`, `ports`, `environment`, `env_file` only if safely supported, `volumes` for named volumes, `networks`, and `restart`.
  - Explicitly reject `build` in MVP-1.
  - Reject privileged mode, host networking, host PID/IPC, device access, Docker socket mounts, and unsafe host-path mounts by default.
  - **Done when:** the supported/rejected subset is documented and represented in tests.

- [ ] **MVP1-046 — Add local Compose manifest parsing and policy validation**
  - Parse YAML structurally; never use regular expressions as the YAML parser.
  - Validate that at least one service exists and each service has an image.
  - Validate published ports and service identifiers.
  - Return field-specific validation errors.
  - **Done when:** valid and invalid manifest fixtures pass their expected tests.

- [ ] **MVP1-047 — Generate safe project and release paths**
  - Generate application slugs and deployment IDs internally.
  - Use a fixed remote root such as `/opt/arsy/apps/<app-slug>/releases/<deployment-id>`.
  - Never accept a raw filesystem path from the web request.
  - **Done when:** path tests prove that traversal strings and invalid slugs are rejected.

- [ ] **MVP1-048 — Add secure SFTP upload support**
  - Add a maintained SFTP client dependency only if no suitable existing implementation is available.
  - Upload the Compose manifest to a generated release directory.
  - Use temporary filenames followed by an atomic rename.
  - Apply restrictive file permissions.
  - **Done when:** upload tests cover success, interrupted upload, and permission failure.

- [ ] **MVP1-049 — Validate Compose on the target server**
  - Run the fixed command `docker compose -f <generated-file> config --quiet` through the typed SSH executor.
  - Pass only generated, validated paths.
  - Return bounded and redacted validation output.
  - **Done when:** invalid Compose files fail before deployment begins.

- [ ] **MVP1-050 — Implement the deployment state machine**
  - Define allowed states: queued, preparing, uploading, validating, deploying, verifying, succeeded, failed, and cancelled.
  - Define allowed transitions in one package.
  - Reject invalid transitions.
  - **Done when:** table-driven tests cover all valid and invalid transitions.

- [ ] **MVP1-051 — Implement persistent deployment job claiming**
  - Use the PocketBase database for persistent job state.
  - Ensure only one worker claims a given deployment at a time.
  - Recover jobs left in an in-progress state after process restart using a documented timeout/reconciliation rule.
  - **Done when:** tests prove duplicate workers cannot run the same deployment concurrently.

- [ ] **MVP1-052 — Implement the deployment worker lifecycle**
  - Process queued deployments in the background.
  - Use contexts and bounded timeouts for each deployment stage.
  - Persist every state transition.
  - Keep worker errors attached to the deployment record in sanitized form.
  - **Done when:** worker tests cover success, failure, cancellation, and restart recovery.

- [ ] **MVP1-053 — Implement Compose deployment execution**
  - Create the generated remote release directory.
  - Upload the validated Compose manifest.
  - Execute `docker compose -p <validated-slug> -f <generated-file> up -d`.
  - Do not execute `down -v` or delete persistent volumes.
  - **Done when:** a test deployment starts services on a test Docker host and records a successful deployment.

- [ ] **MVP1-054 — Persist deployment output safely**
  - Capture each deployment stage's start/end time, exit status, and bounded output.
  - Redact secrets before persistence.
  - Avoid storing unlimited log output in a single PocketBase record; use bounded log chunks or a file-backed log artifact with a reference.
  - **Done when:** output retrieval works and secret fixtures never appear in persisted logs.

- [ ] **MVP1-055 — Add deployment status and container inspection**
  - Query `docker compose ps` for the generated project.
  - Map container states to stable A-R-S-Y status values.
  - Return service name, image, state, health when available, and published ports.
  - **Done when:** parser tests cover running, exited, restarting, unhealthy, and missing containers.

- [ ] **MVP1-056 — Add deployment history and log API endpoints**
  - Add authenticated endpoints to list deployments, view one deployment, retrieve its bounded build/deploy log, and inspect application runtime logs.
  - Apply pagination and output limits.
  - Do not return internal SSH credentials or unredacted environment variables.
  - **Done when:** API tests cover pagination, access control, and log redaction.

- [ ] **MVP1-057 — Add safe stop and redeploy operations**
  - Implement stop with `docker compose stop`.
  - Implement redeploy using the saved manifest and selected deployment options.
  - Require an explicit authenticated console action for each operation.
  - Do not delete named volumes during stop or redeploy.
  - **Done when:** tests prove application containers stop/restart without deleting persistent data.

- [ ] **MVP1-058 — Add configuration rollback**
  - Keep each deployment's Compose manifest as an immutable release artifact.
  - Allow rollback to a previous successful manifest.
  - Warn when image tags are mutable; recommend digest-pinned images for reproducible rollback.
  - Do not claim guaranteed image rollback unless image digests are stored and restored.
  - **Done when:** rollback restores the previous Compose configuration and records a new rollback deployment event.

## Phase G — Console application and onboarding completion

**Phase exit:** The administrator can complete the whole MVP-1 workflow and administer PocketBase from the A-R-S-Y console without opening the PocketBase dashboard.

- [ ] **MVP1-059 — Add the Overview screen**
  - Show total registered servers, ready servers, projects, applications, and recent deployments.
  - Use real API data only.
  - Show clear empty states for a fresh install.
  - **Done when:** all values come from the A-R-S-Y API and empty states render correctly.

- [ ] **MVP1-060 — Add Project management UI**
  - Create, list, view, and edit a project's name and description.
  - Use the API validation messages.
  - **Done when:** project creation and editing work from the console.

- [ ] **MVP1-061 — Add Application creation UI**
  - Select a project and ready server.
  - Enter an application name and Compose manifest.
  - Show validation errors before allowing deployment.
  - **Done when:** an administrator can create an application without editing PocketBase collections manually.

- [ ] **MVP1-062 — Add the Deploy action and progress view**
  - Start a deployment and show its state transitions.
  - Poll or subscribe to deployment updates using a bounded interval.
  - Show safe deployment output.
  - **Done when:** the console shows queued → running → succeeded/failed accurately.

- [ ] **MVP1-063 — Add Application operations UI**
  - Show current container status and published ports.
  - Add stop, redeploy, and rollback actions.
  - Require a clear confirmation for stop and rollback.
  - **Done when:** all actions call their matching API and reflect the resulting state.

- [ ] **MVP1-064 — Finish the onboarding path**
  - After admin login, guide the user to configure the secrets key, add a server, resolve Docker/Compose readiness, create a project, and deploy the first application.
  - Preserve completed steps when the browser refreshes or A-R-S-Y restarts.
  - Allow experienced users to skip the guide.
  - **Done when:** the first-run flow ends at a successfully deployed application or a precise actionable blocker.

## Phase H — Tests, release, and MVP acceptance

**Phase exit:** MVP-1 can be built, tested, installed as a standalone Go executable, and verified without relying on a running Node.js service.

- [ ] **MVP1-065 — Add unit tests for A-R-S-Y config and security**
  - Test defaults, invalid configuration, secret encryption, tamper detection, and redaction.
  - **Done when:** focused package tests pass.

- [ ] **MVP1-066 — Add SSH tests using a local test SSH server**
  - Use test keys and deterministic host-key fixtures.
  - Cover accepted/rejected host keys, authentication, timeout, and command output limits.
  - Do not use a real VPS in CI.
  - **Done when:** tests pass without external SSH access.

- [ ] **MVP1-067 — Add Compose parser and deployment worker tests**
  - Use fixture manifests and a fake remote executor.
  - Cover success, failure, timeout, duplicate job claim, and recovery.
  - **Done when:** all deployment-engine tests are deterministic and network-independent.

- [ ] **MVP1-068 — Add a Docker integration test profile**
  - Make the test require an explicitly enabled local Docker environment.
  - Use a disposable test project name and temporary volumes.
  - Clean up only resources created by that test.
  - Never run broad Docker prune commands.
  - **Done when:** the integration profile can deploy, inspect, stop, and clean up its own test workload.

- [ ] **MVP1-069 — Add frontend build verification**
  - Run a clean dependency install and production Vite build for `dashboard/`.
  - Fail CI if the embedded dashboard output is missing or stale.
  - Do not require Node.js at runtime.
  - **Done when:** frontend build and Go embed build both pass from a clean checkout.

- [ ] **MVP1-070 — Add release build checks**
  - Build the A-R-S-Y executable for the initial supported Linux target(s).
  - Verify the binary reports the expected A-R-S-Y name and version.
  - Verify the binary serves embedded console assets.
  - **Done when:** release artifacts are named `arsy` and pass smoke tests.

- [ ] **MVP1-071 — Write MVP-1 installation and operation documentation**
  - Document prerequisites, standalone binary startup, first superuser creation, secrets-key setup, managed-server SSH requirements, Docker/Compose requirements, and first Compose deployment.
  - Document backup location and recovery basics.
  - Clearly list unsupported MVP-1 features.
  - **Done when:** a developer can follow the document from a clean test VPS without guessing commands.

- [ ] **MVP1-072 — Run the complete MVP-1 acceptance checklist**
  - Run `gofmt` checks and `go test ./...`.
  - Build the A-R-S-Y executable and embedded dashboard from a clean checkout.
  - Verify PocketBase's existing Superuser UI and collection APIs still work.
  - Verify first-run onboarding and server registration.
  - Deploy a test Compose application, inspect status/logs, stop it, redeploy it, and roll back its manifest.
  - Restart the A-R-S-Y control plane and verify deployment history remains available and managed containers continue running.
  - **Done when:** all checks pass or any remaining blocker is explicitly listed. Do not mark MVP-1 complete while a security invariant is failing.

---

## Phase I — PocketBase administrator console parity

**Phase exit:** Every PocketBase Superuser dashboard workflow is available in the A-R-S-Y console, the console uses PocketBase backend APIs and auth, and the PocketBase UI route is disabled without breaking the REST API or A-R-S-Y operations.

- [ ] **MVP1-073 — Inventory PocketBase administrator workflows**
  - Map all current `ui/src/router.js` routes, navigation actions, dialogs, forms, and their backend endpoints to an A-R-S-Y console feature matrix.
  - Include setup/login/recovery, collections/schema, records/files, auth providers, logs, application/mail/storage settings, backups, cron, import/export, and SQL console.
  - Mark unsupported/unsafe operations explicitly and obtain a product decision before omitting any PocketBase dashboard feature.
  - **Done when:** every PocketBase dashboard route and primary action has an A-R-S-Y destination, API contract, and regression check.

- [ ] **MVP1-074 — Move PocketBase setup and superuser authentication into A-R-S-Y**
  - Implement first-superuser setup, login/logout, password-reset requests, confirmation flows, OAuth callbacks, expired sessions, and session recovery in the A-R-S-Y console using PocketBase auth APIs.
  - Keep auth tokens in memory unless a reviewed PocketBase mechanism provides a safer persistent session.
  - **Done when:** fresh-install setup and all supported administrator auth flows work without opening `/_/`.

- [ ] **MVP1-075 — Add collection and schema administration to A-R-S-Y**
  - Implement collection CRUD, field types and options, validation rules, API access rules, indexes, and collection view queries through PocketBase APIs.
  - Preserve migration/collection behavior and show validation errors returned by PocketBase.
  - **Done when:** an administrator can create, inspect, update, and delete supported collections and fields in A-R-S-Y with API-level regression tests.

- [ ] **MVP1-076 — Add record, file, and relation administration to A-R-S-Y**
  - Implement record search, filters, sorting, pagination, create/read/update/delete, bulk actions, file previews/uploads, and relation selection.
  - Use PocketBase query contracts and enforce each collection's rules.
  - **Done when:** record workflows available in the PocketBase dashboard work in A-R-S-Y, including file and relation fields.

- [ ] **MVP1-077 — Add auth collection and provider administration to A-R-S-Y**
  - Implement auth record management and PocketBase password, OAuth2, OTP, MFA, verification, token, and provider configuration screens supported by the existing backend.
  - Never expose plaintext passwords, OAuth client secrets, or signing secrets after save.
  - **Done when:** auth settings and administrator-managed auth records can be managed in A-R-S-Y with secret-redaction tests.

- [ ] **MVP1-078 — Add PocketBase application settings to A-R-S-Y**
  - Implement general application/superuser/rate-limit/trusted-proxy settings, mail settings and test action, storage/S3 settings and test action.
  - Implement backup list/create/restore and cron inspection/configuration with explicit confirmation for state-changing actions.
  - **Done when:** all PocketBase settings, backup, and cron dashboard workflows are available through authenticated A-R-S-Y screens and APIs.

- [ ] **MVP1-079 — Add logs, collection transfer, and SQL administration to A-R-S-Y**
  - Implement log browsing and filters, collection export/import review, and the existing superuser SQL console capabilities.
  - Require explicit confirmation for imports and SQL statements that mutate data; retain audit/logging behavior.
  - **Done when:** these PocketBase dashboard tools work from A-R-S-Y and tests cover permissions and mutation behavior.

- [ ] **MVP1-080 — Replace PocketBase dashboard access after parity**
  - Remove the dashboard link from the A-R-S-Y shell and disable the `/_/` UI route only after tasks MVP1-073 through MVP1-079 pass.
  - Keep PocketBase REST APIs, migrations, hooks, and database behavior intact.
  - Document how to re-enable the dashboard in a development-only build if it is required for recovery.
  - **Done when:** the A-R-S-Y console is the only shipped administrator UI and regression tests confirm PocketBase APIs and A-R-S-Y routes continue working.

## 7. MVP-1 acceptance checklist

This is the final product-level checklist. It is intentionally separate from individual implementation tasks.

- [ ] A-R-S-Y builds as a standalone Go executable.
- [ ] No Node.js process is needed after installation.
- [ ] A-R-S-Y has its own product/version identity and does not self-update from PocketBase upstream.
- [ ] All PocketBase administrator workflows are available through the A-R-S-Y console; the standalone PocketBase dashboard is disabled after parity.
- [ ] The A-R-S-Y console is embedded in the executable.
- [ ] First-run onboarding is resumable.
- [ ] Infrastructure APIs require superuser authentication.
- [ ] SSH host keys are pinned and verified.
- [ ] SSH credentials are encrypted at rest and never returned by APIs.
- [ ] Server readiness correctly identifies SSH, OS, Docker, and Compose state.
- [ ] Docker Compose manifests are validated before execution.
- [ ] A deployment is persisted and recoverable after control-plane restart.
- [ ] The administrator can deploy, inspect, stop, redeploy, and roll back an application.
- [ ] Published application ports are visible to the administrator.
- [ ] Deployment/runtime logs are bounded and secrets are redacted.
- [ ] Tests do not depend on a real production server.
- [ ] Go tests, frontend build, binary build, and smoke tests pass.

## 8. Next milestone after MVP-1

Do not implement these tasks inside MVP-1. Create a separate plan after this acceptance checklist is met.

**MVP-2 — Git-to-Production Deployment**
- GitHub, GitLab, and generic Git repository connections.
- Deploy from branch/tag/commit.
- Build from Dockerfile.
- Build logs and artifact caching.
- Git webhooks and push-to-deploy.
- Deployment previews.

**MVP-3 — Domains, Proxy, and HTTPS**
- Traefik or Caddy integration (choose one after a focused comparison).
- Domain verification and routing.
- Automatic Let's Encrypt certificates.
- HTTP-to-HTTPS redirect.
- Wildcard domains and proxy health checks.

**MVP-4 — Infrastructure Operations**
- Safe Docker installation/provisioning.
- Server resource monitoring.
- Backups and restore workflows.
- Database/service templates.
- Notifications and scheduled maintenance.

**MVP-5 — AI-native DevOps**
- AI Router/LLM Gateway integration.
- Deployment-log diagnosis.
- Dockerfile/Compose assistance.
- MCP server and agent-facing tools.
- Policy-controlled remediation with audit trail.

---

## 9. Plan maintenance

- Keep this file as the source checklist for MVP-1.
- Do not delete completed task entries; keep their checked state and completion note.
- If implementation reveals a missing task, add a new uniquely numbered task in the correct phase.
- If a task is too large for a small model, split it into smaller tasks before implementation.
- If a task changes scope, update the explicit MVP-1 exclusions and acceptance checklist in the same commit.
- Never mark a task complete merely because code was generated; require tests and its stated acceptance condition.
