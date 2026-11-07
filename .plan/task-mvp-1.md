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

**Do not replace the PocketBase Superuser UI.** Add a separate A-R-S-Y console and preserve the existing PocketBase admin UI at its existing route.

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

### Boundaries

- New A-R-S-Y Go code belongs under `arsy/` (or a clearly justified subpackage).
- The A-R-S-Y HTTP API uses the `/api/arsy/v1/` prefix.
- The A-R-S-Y web console uses the `/arsy` route.
- The existing PocketBase API and Superuser UI routes must keep working.
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

- [ ] **MVP1-001 — Capture the baseline build and test result**
  - Inspect the current Go version and existing Makefile targets.
  - Run `go test ./...` and the existing build command for `examples/base`.
  - Record failures that already exist; do not fix unrelated failures in this task.
  - **Done when:** baseline results are documented in the task completion note.
  - **Remaining:** A true pre-change baseline was not captured before implementation began. The first recorded online baseline is the post-foundation CI run; do not treat it as a pre-change comparison.

- [x] **MVP1-002 — Change the Go module identity to A-R-S-Y**
  - Change the module path in `go.mod` to `github.com/bridge-allah/a-r-s-y`.
  - Replace only repository-local imports that point to the old PocketBase application module.
  - Do not rename external dependencies such as PocketBase's separately hosted `dbx`, validation, or other modules.
  - Run formatting and `go mod tidy`.
  - **Done when:** `go list ./...` and `go test ./...` resolve with no old local-module imports remaining.
  - **Changed:** migrated `go.mod` and repository-local Go imports to `github.com/bridge-allah/a-r-s-y`; updated release version injection.
  - **Verified:** `go mod tidy`, `go list ./...`, and `go test ./...` passed in GitHub Actions migration run.
  - **Remaining:** None.

- [x] **MVP1-003 — Rename the release project and executable**
  - Update `.goreleaser.yaml` project name and binary name to `arsy`.
  - Keep the current supported OS/architecture matrix unless a build target is proven broken.
  - **Done when:** a local release build produces an executable named `arsy` (or `arsy.exe`).
  - **Changed:** renamed GoReleaser project and binary to `arsy`; CI builds the executable into `.builds/arsy`.
  - **Verified:** `CGO_ENABLED=0 go build -o /tmp/arsy-bin ./examples/base` passed in the migration workflow.
  - **Remaining:** Full multi-platform GoReleaser archive publishing has not been run.

- [x] **MVP1-004 — Make version output identify A-R-S-Y**
  - Keep the existing version injection mechanism.
  - Update user-visible product/version labels where they are directly tied to the executable.
  - Avoid mass-renaming internal PocketBase types and symbols.
  - **Done when:** `arsy version` or the existing version command identifies the product as A-R-S-Y and displays the build version.
  - **Changed:** root CLI metadata now identifies as A-R-S-Y and prefixes the injected version with the product name.
  - **Verified:** the migration workflow ran `/tmp/arsy-bin --version` and asserted that output contains `A-R-S-Y`.
  - **Remaining:** None.

- [x] **MVP1-005 — Disable the PocketBase upstream updater registration**
  - Remove the upstream PocketBase updater registration from the A-R-S-Y executable entry point.
  - Do not delete the reusable PocketBase updater package.
  - Do not point the updater at A-R-S-Y until a compatible A-R-S-Y release feed exists.
  - **Done when:** the A-R-S-Y binary cannot accidentally replace itself with a PocketBase release.
  - **Changed:** removed the PocketBase updater import and registration from `examples/base/main.go`.
  - **Verified:** re-fetched the updated entry point and confirmed it no longer registers or imports the updater; commit `4c2dcf3`.
  - **Remaining:** local Go build/test execution is not available in this session.

- [x] **MVP1-006 — Replace the inherited root README with A-R-S-Y project documentation**
  - Explain the A-R-S-Y purpose, current MVP status, Go-native runtime, and build/run commands.
  - Clearly state that A-R-S-Y is independently maintained and is derived from PocketBase.
  - Preserve accurate upstream license and attribution information.
  - Do not claim features that are not implemented.
  - **Done when:** the README describes the actual A-R-S-Y repository rather than presenting it as an unchanged PocketBase release.
  - **Changed:** replaced the inherited PocketBase README with A-R-S-Y purpose, status, architecture, build/run, configuration, security, and attribution documentation.
  - **Verified:** README reviewed against implemented functionality and current MVP scope.
  - **Remaining:** Installation instructions will be expanded as the deployment features land.

- [x] **MVP1-007 — Add a minimal Go CI workflow**
  - Add a GitHub Actions workflow that installs the required Go version.
  - Run formatting checks, `go test ./...`, and a build of the A-R-S-Y executable.
  - Do not add frontend CI until the A-R-S-Y console exists.
  - **Done when:** the workflow runs successfully on a push and pull request.
  - **Changed:** added GitHub Actions CI for Go tests/build, dashboard asset build, and the existing PocketBase UI build.
  - **Verified:** GitHub Actions run `42336003844` passed all jobs, including `go test ./...`, dashboard build, Go binary build, and diagnostics.
  - **Remaining:** Docker integration coverage is tracked separately in MVP1-068.

## Phase B — Control-plane foundation

**Phase exit:** A-R-S-Y has its own package boundary, validated configuration, persistent data layout, and health/diagnostic commands.

- [x] **MVP1-008 — Create the A-R-S-Y package skeleton**
  - Add the top-level `arsy/` package directory.
  - Add only small domain package placeholders needed by later tasks: `config`, `httpapi`, `servers`, `security`, and `deployments`.
  - Do not move PocketBase source files.
  - **Done when:** the new packages compile and have a short package comment describing their responsibility.
  - **Changed:** added the `arsy/config`, `arsy/httpapi`, `arsy/servers`, `arsy/security`, and `arsy/deployments` packages with responsibility comments.
  - **Verified:** `go test ./...` and the standalone build passed in GitHub Actions run `42336003844`.
  - **Remaining:** None.

- [x] **MVP1-009 — Define the A-R-S-Y runtime configuration type**
  - Create a typed configuration structure for data directory, HTTP API prefix, console route, and secret-key source.
  - Keep defaults in one place.
  - Do not add provider-specific configuration or secret values to logs.
  - **Done when:** configuration defaults have unit tests.
  - **Changed:** added typed runtime configuration with centralized defaults for data directory, API prefix, console path, and secret-key environment variable name.
  - **Verified:** `go test ./arsy/config` passed, including default-value coverage in GitHub Actions run `42336003844`.
  - **Remaining:** None.

- [x] **MVP1-010 — Load and validate A-R-S-Y environment configuration**
  - Support documented `ARSY_*` environment variables for A-R-S-Y-specific settings.
  - Reject malformed values with actionable error messages.
  - Do not change PocketBase's existing flags or environment variable semantics.
  - **Done when:** valid configuration loads and invalid configuration fails before the server starts.
  - **Changed:** added `ARSY_DATA_DIR`, `ARSY_API_PREFIX`, `ARSY_CONSOLE_PATH`, and `ARSY_SECRET_KEY_ENV` loading and validation; documented them in README.
  - **Verified:** table-driven environment and malformed-value tests passed in GitHub Actions run `42336003844`.
  - **Remaining:** None.

- [x] **MVP1-011 — Establish the A-R-S-Y data-directory layout**
  - Define stable subdirectories for application data, encrypted temporary files, and deployment manifests.
  - Create directories with restrictive permissions where secrets or private files may exist.
  - Keep existing PocketBase database/data migration behavior compatible.
  - **Done when:** startup creates the expected directories and tests verify permissions on supported systems.
  - **Changed:** added private runtime, temporary, and manifest directories with owner-only permissions.
  - **Verified:** `go test ./arsy/config` confirmed directory creation and mode `0700`; full CI passed in run `42336003844`.
  - **Remaining:** None.

- [x] **MVP1-012 — Add an A-R-S-Y liveness endpoint**
  - Register `GET /api/arsy/v1/health`.
  - Return only product name, version, and a simple alive status.
  - Do not include database credentials, environment values, filesystem paths, or secrets.
  - **Done when:** an HTTP test confirms the endpoint works and the response contains no sensitive configuration.
  - **Changed:** registered `GET /api/arsy/v1/health` with only product, version, and alive status.
  - **Verified:** `TestHealthEndpointIsMinimal` passed and asserted the response contains exactly those three fields in run `42336003844`.
  - **Remaining:** None.

- [x] **MVP1-013 — Add an authenticated readiness endpoint**
  - Register `GET /api/arsy/v1/ready`.
  - Check only local control-plane readiness: database available, migrations applied, required configuration valid.
  - Require superuser authentication or restrict the response to a minimal non-sensitive status.
  - **Done when:** tests cover ready and not-ready responses.
  - **Changed:** registered `GET /api/arsy/v1/ready`, requiring superuser authentication and returning ready/not-ready state.
  - **Verified:** tests passed for unauthenticated (401), authenticated ready (200), and authenticated not-bootstrapped (503) responses in run `42336003844`.
  - **Remaining:** Readiness currently checks local bootstrap state; deeper dependency checks can be added as the control plane grows.

- [x] **MVP1-014 — Add an A-R-S-Y diagnostics CLI command**
  - Add a command such as `arsy doctor`.
  - Report version, Go build/runtime information, data-directory accessibility, database readiness, and secret-key configuration status.
  - Never print secret values.
  - **Done when:** the command exits non-zero for a broken local setup and exits zero for a valid setup.
  - **Changed:** added `arsy doctor` diagnostics for product/version, Go runtime, data directories, database readiness, and secret-key configuration without printing secret values.
  - **Verified:** GitHub Actions ran `ARSY_DATA_DIR=... ARSY_SECRET_KEY=... .builds/arsy doctor` successfully in run `42336003844`.
  - **Remaining:** None.

## Phase C — Separate console and first-run onboarding

**Phase exit:** An administrator can sign in to a dedicated A-R-S-Y console, see onboarding status, and follow a guided path to register the first server. The PocketBase Superuser UI remains intact.

- [x] **MVP1-015 — Create a separate A-R-S-Y console frontend**
  - Create a new `dashboard/` directory.
  - Use a small Vite-based static frontend; vanilla JavaScript is acceptable.
  - Keep Node/npm as build-time tools only.
  - Do not modify `ui/src` or replace the PocketBase Superuser dashboard.
  - **Done when:** `npm run build` in `dashboard/` produces a static `dist/` directory.
  - **Changed:** added a separate Vite-based vanilla-JavaScript dashboard in `dashboard/`; PocketBase `ui/` remains separate.
  - **Verified:** `npm install && npm run build` in `dashboard/` passed in GitHub Actions run `42336003844`.
  - **Remaining:** None.

- [x] **MVP1-016 — Add a minimal A-R-S-Y console shell**
  - Create a simple layout with navigation for Overview, Servers, Projects, and Deployments.
  - Use placeholder empty states for pages that are not yet implemented.
  - Do not create decorative controls that look functional but do nothing.
  - **Done when:** the built console renders all four navigation destinations without runtime errors.
  - **Changed:** added Overview, Servers, Projects, and Deployments navigation with explicit empty states.
  - **Verified:** Vite production build and embedded console route test passed in GitHub Actions run `42336003844`.
  - **Remaining:** Browser-level end-to-end interaction tests are not yet configured.

- [x] **MVP1-017 — Embed the A-R-S-Y console assets into Go**
  - Add a dedicated Go embed package for `dashboard/dist`.
  - Do not reuse or overwrite `ui/embed.go`.
  - Add a build step that generates the frontend assets before compiling the executable.
  - **Done when:** the A-R-S-Y binary serves the console without requiring Node.js or a separate frontend server.
  - **Changed:** added Go embed package for `dashboard/dist` and made CI/Makefile build dashboard assets before compiling the executable.
  - **Verified:** dashboard build, `go test ./...`, and standalone Go build passed in GitHub Actions run `42336003844`.
  - **Remaining:** None.

- [x] **MVP1-018 — Register the A-R-S-Y console route**
  - Serve the console at `/arsy` and its static assets under a predictable asset path.
  - Keep the existing PocketBase `/_/` route and all PocketBase API routes working.
  - Add SPA fallback only within the A-R-S-Y console route.
  - **Done when:** HTTP tests confirm both the A-R-S-Y console route and PocketBase admin route remain reachable.
  - **Changed:** registered `/arsy` and SPA/static asset handling without replacing PocketBase's existing `/_/` route.
  - **Verified:** `TestConsoleRouteServesEmbeddedDashboard` passed in GitHub Actions run `42336003844`.
  - **Remaining:** Full HTTP integration test against the complete PocketBase router remains useful before release.

- [x] **MVP1-019 — Register the A-R-S-Y API route group**
  - Create a shared router group under `/api/arsy/v1`.
  - Keep route registration separate from PocketBase's existing default API handlers.
  - Add a small test route or health route through this group.
  - **Done when:** route registration is centralized and existing collection APIs still pass tests.
  - **Changed:** centralized the versioned A-R-S-Y endpoints under a shared router group at the configured API prefix.
  - **Verified:** health/readiness route tests and the full Go CI suite passed in run `42336729656`.
  - **Remaining:** None.

- [ ] **MVP1-020 — Add superuser-only authorization middleware**
  - Reuse PocketBase's existing superuser authentication.
  - Protect all infrastructure-management endpoints.
  - Return consistent 401/403 responses.
  - Do not add a second user/password system.
  - **Done when:** tests prove unauthenticated and regular authenticated users cannot manage servers or deployments.

- [ ] **MVP1-021 — Add a console login and logout flow**
  - Use the existing PocketBase Superuser authentication API.
  - Keep the auth token in memory for MVP-1; do not persist it in `localStorage`.
  - Show clear expired-session and invalid-credential states.
  - **Done when:** a superuser can log in, access protected console pages, and log out.

- [ ] **MVP1-022 — Add the first-run onboarding status API**
  - Return non-sensitive setup state: whether a superuser exists, whether the secrets key is configured, and whether a server has been registered.
  - Do not return emails, credentials, private keys, or environment values.
  - **Done when:** tests cover fresh install and already-configured states.

- [ ] **MVP1-023 — Build the onboarding wizard**
  - Show a welcome screen and explain the A-R-S-Y control-plane/managed-server model.
  - If no superuser exists, show the existing CLI command pattern `arsy superuser create EMAIL PASSWORD`; do not create an unauthenticated public password-creation endpoint.
  - After login, guide the administrator through secret-key readiness and first-server registration.
  - Allow the wizard to be resumed after refresh or restart.
  - **Done when:** a fresh-install user can follow the wizard without guessing the next step.

## Phase D — Data model and secret handling

**Phase exit:** Projects, servers, applications, and deployments have persistent schemas; all management data is protected; credentials are encrypted and never exposed.

- [x] **MVP1-024 — Add an A-R-S-Y-owned migration registration package**
  - Follow the repository's existing PocketBase system migration registration pattern.
  - Keep A-R-S-Y migrations separate from historical upstream migrations.
  - Register the new package from the A-R-S-Y executable.
  - **Done when:** a clean database applies the A-R-S-Y migration package and existing migrations still pass.
  - **Changed:** added an A-R-S-Y-owned system migration package and registered it from the standalone executable.
  - **Verified:** clean test-app migration applied successfully; `go test ./arsy/migrations` passed in GitHub Actions run `42336729656`.
  - **Remaining:** None.

- [x] **MVP1-025 — Create the projects collection**
  - Add project name, validated slug, description, and timestamps.
  - Add unique slug validation.
  - Keep list/view/create/update/delete superuser-only.
  - **Done when:** migration and collection tests pass.
  - **Changed:** added the private `arsy_projects` collection with required name, slug pattern, description, and unique slug index.
  - **Verified:** migration tests confirmed collection creation and superuser-only rules in run `42336729656`.
  - **Remaining:** None.

- [x] **MVP1-026 — Create the servers collection**
  - Add server name, host, SSH port, SSH username, operating-system metadata, pinned host-key fingerprint, readiness state, and timestamps.
  - Keep credentials in a separate encrypted field or related protected record.
  - Do not expose private-key ciphertext in standard API responses.
  - **Done when:** schema tests verify required fields, validation, and private collection rules.
  - **Changed:** added the private `arsy_servers` collection with host, SSH port/user, OS/architecture, pinned fingerprint, readiness state, and hidden credential ciphertext field.
  - **Verified:** migration tests confirmed the collection and hidden credential field; full CI passed in run `42336729656`.
  - **Remaining:** SSH credentials are schema-only until encryption and SSH services are implemented.

- [x] **MVP1-027 — Create the applications collection**
  - Add project relation, server relation, application name, validated slug, Compose manifest, and lifecycle state.
  - Support one application mapped to one server in MVP-1.
  - Do not add Git repository fields or buildpack configuration yet.
  - **Done when:** application records can be created and validated through tests.
  - **Changed:** added the private `arsy_applications` collection with project/server relations, validated slug, Compose manifest, and lifecycle state.
  - **Verified:** migration tests confirmed project/server relation targets and private rules in run `42336729656`.
  - **Remaining:** Record-level application service validation is still to be implemented.

- [x] **MVP1-028 — Create the deployments collection**
  - Store application relation, deployment identifier, status, start/end times, manifest checksum, safe summary, and previous deployment reference.
  - Do not store unredacted command output or secrets in summary fields.
  - **Done when:** deployment records can represent queued, running, succeeded, failed, and cancelled states.
  - **Changed:** added the private `arsy_deployments` collection with application relation, unique deployment ID, lifecycle status, timestamps, manifest checksum, safe summary, and previous deployment ID.
  - **Verified:** clean migration and collection tests passed in run `42336729656`.
  - **Remaining:** Deployment state transitions and log storage are tracked in later tasks.

- [ ] **MVP1-029 — Add a central AES-GCM secret encryption helper**
  - Use Go standard-library cryptography.
  - Require a 32-byte key supplied outside the database through documented A-R-S-Y configuration.
  - Store ciphertext with a format/version marker and a random nonce.
  - Return errors for missing, malformed, or incorrect keys.
  - **Done when:** encrypt/decrypt round-trip and tamper-detection tests pass.

- [ ] **MVP1-030 — Add secret-key generation and validation**
  - Add a CLI helper to generate a cryptographically random secret key.
  - Print the key once to stdout only when explicitly requested.
  - Never automatically write the key into source control or log it.
  - Document secure storage with restrictive file permissions or a secret manager.
  - **Done when:** generated keys pass validation and invalid keys prevent credential storage.

- [ ] **MVP1-031 — Add encrypted SSH credential persistence**
  - Encrypt private-key material before saving it.
  - Decrypt only immediately before creating the SSH client.
  - Support unencrypted OpenSSH private keys for MVP-1; return a clear unsupported-format error for encrypted keys if passphrase support is not implemented.
  - **Done when:** database inspection shows ciphertext, not the original private key.

- [ ] **MVP1-032 — Sanitize server API responses**
  - Use explicit response DTOs rather than returning credential-bearing records directly.
  - Omit private-key ciphertext and all decrypted values.
  - Redact credentials from error messages.
  - **Done when:** API tests assert that secret-related fields never appear in JSON responses.

## Phase E — SSH connection and server readiness

**Phase exit:** A superuser can register a remote Linux host, pin its SSH host key, test connectivity, and see Docker/Compose readiness.

- [ ] **MVP1-033 — Validate server connection input**
  - Validate hostname/IP, port range, username, and private-key structure.
  - Reject empty hostnames, invalid ports, and malformed key content.
  - Normalize hostnames without changing the user's intended destination.
  - **Done when:** table-driven validation tests cover valid and invalid inputs.

- [ ] **MVP1-034 — Implement SSH host-key fingerprint calculation**
  - Calculate a stable SHA-256 fingerprint from the SSH server host key.
  - Use the standard OpenSSH fingerprint format.
  - **Done when:** known fixture keys produce expected fingerprints.

- [ ] **MVP1-035 — Require explicit host-key pinning**
  - During the first connection test, display the server fingerprint without marking the server trusted.
  - Require the administrator to provide/confirm the expected fingerprint through an authenticated action.
  - Verify the pinned fingerprint on every later connection.
  - Reject changed or missing fingerprints.
  - **Done when:** tests cover first-seen, matching, mismatched, and changed host keys.

- [ ] **MVP1-036 — Implement a reusable SSH client factory**
  - Use `golang.org/x/crypto/ssh`.
  - Support private-key authentication only in MVP-1.
  - Require strict host-key verification.
  - Apply connection timeout and context cancellation.
  - **Done when:** the factory returns a connected client only for a trusted host.

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

**Phase exit:** The administrator can complete the whole MVP-1 workflow from the A-R-S-Y console without using raw PocketBase collection editing.

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

## 7. MVP-1 acceptance checklist

This is the final product-level checklist. It is intentionally separate from individual implementation tasks.

- [ ] A-R-S-Y builds as a standalone Go executable.
- [ ] No Node.js process is needed after installation.
- [ ] A-R-S-Y has its own product/version identity and does not self-update from PocketBase upstream.
- [ ] The existing PocketBase admin interface remains available.
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
