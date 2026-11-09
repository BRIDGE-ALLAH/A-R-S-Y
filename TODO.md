# A-R-S-Y Transformation Roadmap

> **Single source of truth for the product transformation backlog.** Every checkbox starts unchecked. Only check an item after its implementation is merged and its acceptance evidence is recorded.
>
> **Product goal:** continue developing the already-transformed **A-R-S-Y Intelligent Enterprise ORM + Full-stack BaaS**. The current product direction is that the web experience, administration UI, setup flow and installer are already A-R-S-Y-native; this backlog must not treat them as unbuilt replacements. Verify that claim with repository/runtime evidence, then focus implementation on engine completeness, integration, security, database adapters, modules and developer experience. PocketBase-derived code may remain only as internal implementation lineage where required; do not reintroduce PocketBase as a user-facing product or blindly delete upstream-derived internals.

## 0. Product contract and non-negotiable rules

- **A-R-S-Y is the product.** The web, installer and user-facing product surfaces are considered transformed; tasks verify and harden them rather than rebuild them. PocketBase is internal code lineage only, not a user-facing product dependency.
- **One application engine:** model metadata, schema, validation, security, queries, transactions, lifecycle, modules and APIs must have one coherent A-R-S-Y contract.
- **Manifest-first and deterministic:** modules and apps are declared through versioned manifests. AI may draft or repair manifests, but deterministic validators and policy enforcement decide what can run.
- **Database adapters, not database leakage:** ORM and business logic must not depend directly on SQLite-specific SQL or PocketBase collection internals.
- **Security by default:** deny by default, server-derived field permissions, tenant isolation, parameterized queries, no implicit `SELECT *` for restricted access, no client-authoritative policy.
- **No premature infrastructure:** do not add Kafka, Kubernetes operator, ML query optimizer, GraphQL, or other infrastructure until a measured product requirement and an accepted design justify it.
- **No false completion:** code, tests, CI and documented evidence are required before checking a task.
- **Compatibility and licensing:** preserve data migration paths, upstream notices, third-party licenses and security fixes. Remove obsolete product surfaces only after dependency/runtime/build/package checks.

## 1. Task format and completion policy

Each task contains four mandatory, unchecked checklists:
- **SPEC** — exact scope, behavior, inputs, outputs and explicit non-goals.
- **DESIGN** — affected packages, data flow, interfaces, migration/compatibility strategy and implementation approach.
- **RULE** — invariants, security constraints, failure behavior and forbidden implementation shortcuts.
- **ACCEPTANCE** — observable, testable outcomes with named test/evidence artifacts.

**Task metadata is mandatory:** stable ID, outcome-based title, priority (P0–P3), size (XS/S/M; split anything larger than M), dependencies, explicit non-goals and evidence location. Avoid vague tasks such as “improve”, “complete” or “support” without measurable outcomes.

**Definition of Ready:** scope and non-goals are clear; dependencies are identified; design boundaries and acceptance checks are reviewable; no unresolved architecture decision blocks implementation.

**Definition of Done:** all four checklists are checked; implementation/docs are merged to `main`; relevant unit, integration, security and CI checks pass; migration/compatibility impact is covered; evidence (PR, test run, screenshot, benchmark or report) is linked. Code merely written or a PR merely opened is not completion.

**Granularity:** one task produces one independently reviewable outcome, normally within one focused PR. Split cross-layer work into engine/API/UI/test tasks. Do not mark a parent complete while required child work remains.

---

# PHASE 0 — Baseline, inventory and product boundary

## [ ] ARSY-001 — Establish the PocketBase transformation inventory
**Depends on:** none
**Priority:** P0 · **Size:** S · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Inventory every upstream-derived package, route, command, UI page, installer, migration, asset, configuration key and build artifact.
- [ ] Classify each item as KEEP-INTERNAL, ADAPT, REPLACE, REMOVE-AFTER-MIGRATION, or LICENSE/SECURITY-REQUIRED.
- [ ] Record user-visible PocketBase terminology and entry points, including indirect links and error messages.

### DESIGN
- [ ] Add a machine-readable inventory under `docs/architecture/pocketbase-transformation/`.
- [ ] Map runtime entry points from executable startup through router, database, UI embedding and installer.
- [ ] Record package dependency graph and ownership boundary: A-R-S-Y domain/application vs infrastructure foundation.

### RULE
- [ ] Do not delete code based only on its name or upstream origin.
- [ ] Preserve license notices, security fixes, data compatibility and required migrations.
- [ ] Every REMOVE decision must cite a replacement and dependency proof.

### ACCEPTANCE
- [ ] Inventory covers all root packages and shipped web/install surfaces.
- [ ] Each classified item has an owner, target phase and verification method.
- [ ] Architecture review approves the transformation boundary.

## [ ] ARSY-002 — Define A-R-S-Y product architecture and public contract
**Depends on:** ARSY-001
**Priority:** P0 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Define product layers: Application Platform, Intelligent ORM Engine, Runtime/API, Data Adapters, Web Studio, CLI/Installer.
- [ ] Define public concepts: App, Module, Model, Field, Relation, Record, Tenant, Role, Policy, View, Action, Workflow.
- [ ] Define which legacy PocketBase concepts are internal compatibility details and must never leak into public APIs.

### DESIGN
- [ ] Publish a concise architecture decision record with package ownership and dependency direction.
- [ ] Define versioning policy for manifests, public APIs, CLI and database adapters.
- [ ] Define supported extension points and lifecycle contracts.

### RULE
- [ ] Domain packages must not import web UI, CLI or a concrete database driver.
- [ ] No second registry or competing metadata source of truth.
- [ ] Public APIs must use A-R-S-Y terminology and stable identifiers.

### ACCEPTANCE
- [ ] Architecture document includes context, container and package-level diagrams.
- [ ] Public vocabulary and compatibility policy are documented.
- [ ] New packages can be placed unambiguously in the architecture.

## [ ] ARSY-003 — Define release gates and transformation scorecard
**Depends on:** ARSY-002
**Priority:** P1 · **Size:** S · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Define measurable completion criteria for engine, web, installer, CLI, security, tenancy and developer experience.
- [ ] Track remaining PocketBase-facing product surfaces separately from internal upstream code.
- [ ] Define minimum supported OS, database versions, Go/Node toolchains and browser matrix.

### DESIGN
- [ ] Add a CI-generated scorecard or a maintained evidence table in `docs/architecture/`.
- [ ] Add explicit release-blocking checks for tests, security, migration compatibility and UI branding.

### RULE
- [ ] Do not claim “PocketBase removed” when only branding has changed.
- [ ] Do not block releases on optional infrastructure that is outside the supported product scope.

### ACCEPTANCE
- [ ] Every release gate has a reproducible command or review procedure.
- [ ] Scorecard distinguishes PASS, FAIL, NOT RUN and BLOCKED.
- [ ] CI reports the gate results on pull requests.

---

# PHASE 1 — Verify transformed A-R-S-Y product surfaces

## [ ] ARSY-010 — Verify A-R-S-Y Studio is the sole administration experience
**Depends on:** ARSY-001, ARSY-002
**Priority:** P1 · **Size:** S · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Inventory shipped Studio routes, screens, navigation, assets and API contracts.
- [ ] Verify model, record, tenant, module, permission, migration and runtime-health workflows.
- [ ] Record any remaining legacy admin UI or route as a reproducible defect.

### DESIGN
- [ ] Trace production startup to embedded web assets and API handlers.
- [ ] Document the canonical Studio app, route ownership and typed API boundary.
- [ ] Implement only evidenced gaps in the existing A-R-S-Y Studio.

### RULE
- [ ] Do not create a second admin console or rebuild a working Studio.
- [ ] UI visibility never replaces backend authorization.
- [ ] Preserve accessible, responsive loading/empty/error/denied states.

### ACCEPTANCE
- [ ] Production binary serves the maintained A-R-S-Y Studio build.
- [ ] Route and browser tests prove primary Studio workflows.
- [ ] No PocketBase Superuser UI bundle or route is shipped or reachable.
- [ ] Remaining gaps have linked, focused issues.

## [ ] ARSY-011 — Verify A-R-S-Y first-run, owner bootstrap and setup lifecycle
**Depends on:** ARSY-010, ARSY-020
**Priority:** P1 · **Size:** S · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Trace clean-data-directory setup through owner, workspace and first-app creation.
- [ ] Verify resume, recovery and already-initialized behavior.
- [ ] Enumerate setup states and externally reachable bootstrap endpoints.

### DESIGN
- [ ] Document the server-owned setup state machine and persistence boundaries.
- [ ] Verify one-time bootstrap credentials, expiry, replay protection and audit events.
- [ ] Implement only proven gaps in the existing A-R-S-Y flow.

### RULE
- [ ] Do not reintroduce PocketBase setup or superuser flows.
- [ ] Never ship default credentials or leave bootstrap endpoints open after initialization.
- [ ] Bootstrap operations must be idempotent and auditable.

### ACCEPTANCE
- [ ] Clean-install E2E creates the first owner and app through A-R-S-Y APIs/UI.
- [ ] Replay, expiry, concurrency and interrupted-setup tests pass.
- [ ] Repeated setup cannot create duplicate owners or bypass authentication.
- [ ] No legacy setup route or UI is shipped.

## [ ] ARSY-012 — Verify A-R-S-Y identity across all shipped product surfaces
**Depends on:** ARSY-001, ARSY-010
**Priority:** P1 · **Size:** S · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Scan web assets, API errors, emails, installer output, CLI help, logs, metadata and generated pages.
- [ ] Maintain an explicit allowlist for legally required upstream attribution and internal references.
- [ ] Classify each match as user-facing defect, legal attribution or internal-only reference.

### DESIGN
- [ ] Centralize product metadata and brand assets where duplication exists.
- [ ] Add a repeatable CI scan with a reviewed allowlist.
- [ ] Fix findings at source rather than fragile output-time replacement.

### RULE
- [ ] Never remove legally required copyright/license notices.
- [ ] Do not expose PocketBase as the product name or reproduce its original user workflow.
- [ ] Do not blindly rename internal symbols in ways that damage compatibility.

### ACCEPTANCE
- [ ] Brand scan passes for release artifacts and user-facing surfaces.
- [ ] Every allowlisted match has a documented reason.
- [ ] Screenshots/smoke tests verify A-R-S-Y identity in Studio, installer and CLI.
- [ ] No unexplained legacy product surface remains.

## [ ] ARSY-013 — Verify one canonical web application and production asset pipeline
**Depends on:** ARSY-010, ARSY-011
**Priority:** P1 · **Size:** S · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Identify frontend apps, packages, lockfiles, build scripts and embedded bundles.
- [ ] Confirm the canonical web app and whether duplicate admin entry points remain active.
- [ ] Verify development, production and release builds use the intended source.

### DESIGN
- [ ] Document the canonical frontend package, build command and embed contract.
- [ ] Remove duplicate paths only after import, runtime, CI and packaging evidence.
- [ ] Expose embedded frontend version/commit in build metadata.

### RULE
- [ ] Do not rebuild the already-transformed web application.
- [ ] Do not delete files based only on directory names.
- [ ] Keep frontend/backend API contracts versioned and testable.

### ACCEPTANCE
- [ ] One documented command builds the canonical Studio.
- [ ] Production binary embeds and serves that build.
- [ ] Clean-checkout build and browser smoke tests pass.
- [ ] No obsolete PocketBase web bundle is packaged or reachable.

# PHASE 2 — Installer, CLI and runtime bootstrap verification

## [ ] ARSY-020 — Verify and document the A-R-S-Y install/upgrade contract
**Depends on:** ARSY-001, ARSY-002
**Priority:** P1 · **Size:** S · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Inventory the existing installer, deployment modes, configuration, data paths and supported platforms.
- [ ] Document install, configure, start, health-check, upgrade, rollback, backup and uninstall behavior.
- [ ] Define supported artifact types and platform/database compatibility.

### DESIGN
- [ ] Map installer flow to runtime bootstrap, migrations and service lifecycle.
- [ ] Document install/upgrade state transitions and recovery points.
- [ ] Identify only missing contracts that require implementation.

### RULE
- [ ] Do not replace a working A-R-S-Y installer without evidence of a concrete defect.
- [ ] Never overwrite user data or secrets silently.
- [ ] Destructive operations require explicit confirmation and a safe backup strategy.

### ACCEPTANCE
- [ ] Installation and upgrade contract matches actual code.
- [ ] Fresh install, upgrade, rollback and uninstall test matrices exist.
- [ ] Unsupported platforms and adapters are stated explicitly.
- [ ] Every gap links to a focused implementation task.

## [ ] ARSY-021 — Verify and harden the existing first-party A-R-S-Y installer
**Depends on:** ARSY-020
**Priority:** P1 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Trace installer from launch/download through configuration, owner bootstrap and service startup.
- [ ] Verify interactive and unattended modes, supported OS targets and upgrade behavior.
- [ ] Capture reproducible installer failures and missing safety checks.

### DESIGN
- [ ] Preserve the existing installer architecture where it is sound.
- [ ] Add integrity verification, preflight checks, dry-run/reporting or recovery only where missing.
- [ ] Keep platform detection, config generation, filesystem setup, service registration and bootstrap testable.

### RULE
- [ ] Do not build a second installer or restore PocketBase installation flows.
- [ ] Never expose secrets in logs, shell history or process arguments.
- [ ] Reruns must be safe; existing data must not be silently replaced.

### ACCEPTANCE
- [ ] Clean-machine install launches A-R-S-Y without PocketBase UI/setup dependencies.
- [ ] Unattended install and safe-rerun tests pass on supported platforms.
- [ ] Failure produces an actionable report and preserves existing data.
- [ ] Installer artifact and product identity checks pass.

## [ ] ARSY-022 — Create the A-R-S-Y CLI command tree
**Depends on:** ARSY-020
**Priority:** P1 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Define commands for `serve`, `install`, `doctor`, `config`, `owner`, `app`, `module`, `model`, `migration`, `tenant`, `backup` and `version`.
- [ ] Replace legacy command names and help output with A-R-S-Y product language.
- [ ] Provide JSON output for automation and stable exit codes.

### DESIGN
- [ ] Use a single command registry and shared config loader.
- [ ] Separate command handlers from application services.
- [ ] Add shell completion generation and command reference docs.

### RULE
- [ ] CLI commands must call application services, not duplicate business logic.
- [ ] Destructive commands require explicit confirmation or a safe non-interactive flag.
- [ ] Secrets must be read from secure files, environment or secret providers, not echoed.

### ACCEPTANCE
- [ ] All documented commands have help, examples and tests.
- [ ] `arsy --help`, `arsy doctor` and `arsy version --json` work.
- [ ] No PocketBase-branded CLI command or output is exposed.
- [ ] CLI integration tests run in CI.

## [ ] ARSY-023 — Implement doctor, diagnostics and recovery commands
**Depends on:** ARSY-022
**Priority:** P2 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Diagnose config, storage, database connectivity, schema version, permissions, ports, secrets and embedded assets.
- [ ] Provide safe repair suggestions and explicit repair commands.
- [ ] Support redacted diagnostic bundle export.

### DESIGN
- [ ] Implement read-only diagnostics first; add narrowly scoped repair actions separately.
- [ ] Reuse health checks from runtime and adapter interfaces.
- [ ] Produce human-readable and JSON reports.

### RULE
- [ ] Diagnostics must not mutate data by default.
- [ ] Secret values and PII must be redacted.
- [ ] Repairs must be idempotent and auditable.

### ACCEPTANCE
- [ ] Tests cover healthy, misconfigured, unavailable and corrupted-state scenarios.
- [ ] Diagnostic output contains actionable remediation and no secret values.
- [ ] Recovery procedures are documented and tested.

## [ ] ARSY-024 — Unify executable startup and embedded web/runtime boot
**Depends on:** ARSY-010, ARSY-021, ARSY-022
**Priority:** P1 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Establish one production entry point for API, Studio, workers, modules and database lifecycle.
- [ ] Remove reliance on legacy PocketBase Superuser startup flags or routes.
- [ ] Support graceful shutdown, readiness and liveness.

### DESIGN
- [ ] Create an explicit application composition root with dependency injection.
- [ ] Start and stop services in deterministic dependency order.
- [ ] Expose health/readiness contracts for adapters, workers and migrations.

### RULE
- [ ] No package-level global mutable runtime state for application services.
- [ ] Startup failure must fail closed and release acquired resources.
- [ ] Graceful shutdown must stop accepting work before closing storage.

### ACCEPTANCE
- [ ] One documented binary command starts the full A-R-S-Y product.
- [ ] Startup/shutdown integration tests cover partial failure and cancellation.
- [ ] Health endpoints reflect actual readiness, not only process liveness.

---

# PHASE 3 — A-R-S-Y engine and metadata source of truth

## [ ] ARSY-030 — Finalize the canonical model descriptor
**Depends on:** ARSY-002
**Priority:** P0 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Define stable model identity, namespace, version, labels, fields, relations, indexes, constraints, hooks and security metadata.
- [ ] Define metadata compatibility and deterministic serialization.
- [ ] Support built-in, custom and module-provided models.

### DESIGN
- [ ] Extend the existing registry/descriptor into the canonical metadata contract.
- [ ] Add typed sub-descriptors with validation and immutable snapshots.
- [ ] Add deterministic fingerprinting for model and module descriptors.

### RULE
- [ ] Registry metadata is the source of truth; database introspection is evidence, not the application schema authority.
- [ ] Invalid descriptors cannot be partially registered.
- [ ] Names and IDs are stable and collision-checked.

### ACCEPTANCE
- [ ] Descriptor round-trip and fingerprint tests are deterministic.
- [ ] Invalid field/relation/security metadata is rejected with actionable errors.
- [ ] Existing registry, field and relation tests remain green.

## [ ] ARSY-031 — Integrate field types, defaults and validators into model registration
**Depends on:** ARSY-030
**Priority:** P0 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Define built-in scalar, temporal, numeric, enum, JSON, file, relation and computed field types.
- [ ] Define required, nullable, default, min/max, length, pattern, uniqueness and custom validation behavior.
- [ ] Define normalization and typed conversion rules.

### DESIGN
- [ ] Register field type handlers in a deterministic field registry.
- [ ] Run metadata validation before schema planning and runtime registration.
- [ ] Share validation rules across API, import, seed and internal service paths.

### RULE
- [ ] Defaults are server-owned and deterministic.
- [ ] Unknown types and invalid values fail closed.
- [ ] Never trust frontend validation as backend validation.

### ACCEPTANCE
- [ ] Table-driven tests cover every supported field type and invalid input.
- [ ] API, seed and import paths produce identical normalized values.
- [ ] Field type compatibility is documented and versioned.

## [ ] ARSY-032 — Integrate relation metadata and relation lifecycle
**Depends on:** ARSY-030, ARSY-031
**Priority:** P0 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Support belongs-to, has-one, has-many and many-to-many relations.
- [ ] Define cascade, restrict, set-null and relation ownership semantics.
- [ ] Define relation loading and expansion contracts.

### DESIGN
- [ ] Add relation resolver and schema constraint planner.
- [ ] Add repository methods for relation reads/writes.
- [ ] Integrate relation validation into model registration and persistence.

### RULE
- [ ] Relation targets must exist and be tenant-compatible.
- [ ] Cascades must be explicit and transaction-bound.
- [ ] Prevent unbounded recursive expansion.

### ACCEPTANCE
- [ ] Tests cover valid/invalid relation graphs and deletion policies.
- [ ] CRUD tests prove relation writes and reads are transactionally consistent.
- [ ] Query tests cover eager loading and bounded expansion.

## [ ] ARSY-033 — Implement ORM validation/default/normalization pipeline
**Depends on:** ARSY-031, ARSY-032
**Priority:** P0 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Define the ordered pipeline: authorize, normalize, apply defaults, validate, persist, emit lifecycle events.
- [ ] Support create, update, upsert and bulk operations.
- [ ] Return typed, field-addressable validation errors.

### DESIGN
- [ ] Implement a reusable pipeline shared by repositories and HTTP APIs.
- [ ] Define context for actor, tenant, operation, request ID and transaction.
- [ ] Add hooks only at documented lifecycle stages.

### RULE
- [ ] Authorization runs before exposing protected field validation details.
- [ ] Failed validation never partially persists data.
- [ ] Internal system fields cannot be overwritten by untrusted clients.

### ACCEPTANCE
- [ ] Tests prove pipeline order and rollback behavior.
- [ ] API, CLI import and module seed use the same validation pipeline.
- [ ] Error format is stable and documented.

## [ ] ARSY-034 — Integrate Unit of Work and lifecycle hooks into persistence
**Depends on:** ARSY-033
**Priority:** P0 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Support transaction-scoped create/update/delete and multi-model operations.
- [ ] Define before/after hooks and after-commit event handlers.
- [ ] Support nested service calls without accidental independent commits.

### DESIGN
- [ ] Integrate existing UoW and hook registries with the persistence mapper.
- [ ] Define transaction propagation and rollback-only behavior.
- [ ] Dispatch external side effects only after successful commit.

### RULE
- [ ] No after-commit event may run for a rolled-back transaction.
- [ ] Hook ordering is deterministic and recursion is bounded.
- [ ] Transactions must be tenant-scoped.

### ACCEPTANCE
- [ ] Tests cover commit, rollback, nested operations and hook failures.
- [ ] No partial writes occur across multi-model operations.
- [ ] Concurrency and race tests pass.

## [ ] ARSY-035 — Implement a complete repository and persistence mapper
**Depends on:** ARSY-033, ARSY-034, ARSY-040
**Priority:** P0 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Support typed CRUD, query, count, exists, bulk, upsert and soft-delete operations.
- [ ] Map descriptors to storage without leaking driver-specific types into domain services.
- [ ] Define record identity and system fields consistently.

### DESIGN
- [ ] Complete the existing persistence mapper behind repository interfaces.
- [ ] Separate query AST compilation from execution and row mapping.
- [ ] Add prepared statement reuse where supported by adapters.

### RULE
- [ ] All writes pass through authorization, validation and transaction lifecycle.
- [ ] No string-concatenated untrusted SQL.
- [ ] No implicit unrestricted projection.

### ACCEPTANCE
- [ ] Repository contract tests run against every supported adapter.
- [ ] CRUD and bulk tests prove typed mapping and rollback.
- [ ] Benchmarks record baseline latency and allocation counts.

---

# PHASE 4 — Database abstraction, schema and migrations

## [ ] ARSY-040 — Define the database adapter contract
**Depends on:** ARSY-002
**Priority:** P0 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Define interfaces for connection/pool, transactions, query execution, schema introspection, migrations and dialect capabilities.
- [ ] Define supported SQLite and PostgreSQL capability matrix.
- [ ] Define tenant-to-database routing contract.

### DESIGN
- [ ] Implement adapter interfaces in an infrastructure package.
- [ ] Keep dialect-specific SQL generation in adapter/dialect packages.
- [ ] Inject adapters through the application composition root.

### RULE
- [ ] Domain and ORM packages cannot import concrete SQLite/PostgreSQL drivers.
- [ ] Unsupported capabilities return explicit typed errors.
- [ ] Every adapter must pass the same repository contract suite.

### ACCEPTANCE
- [ ] SQLite adapter passes contract tests.
- [ ] PostgreSQL adapter passes integration tests before being declared production-ready.
- [ ] Dependency checks detect forbidden driver imports in domain packages.

## [ ] ARSY-041 — Route every A-R-S-Y data operation through tenant-aware providers
**Depends on:** ARSY-040
**Priority:** P0 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Identify and replace direct database opens in HTTP handlers, workers, CLI and provisioning.
- [ ] Resolve tenant storage from authenticated server-side context.
- [ ] Define pool lifecycle and connection limits.

### DESIGN
- [ ] Use one provider/router abstraction for all runtime database access.
- [ ] Pass request-scoped tenant context through repositories and transactions.
- [ ] Centralize pool creation, health checks and shutdown.

### RULE
- [ ] No per-request ad hoc database open/close in production handlers.
- [ ] No client-provided tenant ID may select storage without authorization.
- [ ] Cross-tenant access must fail closed.

### ACCEPTANCE
- [ ] Static search finds no unauthorized direct DB opens outside adapter/bootstrap tests.
- [ ] Integration tests cover HTTP, worker, CLI and migration access paths.
- [ ] Pool leak and concurrent tenant tests pass.

## [ ] ARSY-042 — Implement deterministic schema planner and executor
**Depends on:** ARSY-030, ARSY-040
**Priority:** P0 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Convert registered model descriptors into tables, columns, indexes, foreign keys and constraints.
- [ ] Compare desired schema with actual schema and produce a plan.
- [ ] Support dry-run and explainable migration steps.

### DESIGN
- [ ] Implement planner, dialect renderer, introspector and executor as separate components.
- [ ] Store migration identity, checksum, tenant and schema version.
- [ ] Generate stable migration plans from identical descriptors.

### RULE
- [ ] Never auto-drop or rewrite user data without an explicit approved migration.
- [ ] Schema changes must be tenant-aware and transactional where supported.
- [ ] Plans must be deterministic and reviewable.

### ACCEPTANCE
- [ ] Golden tests verify generated plans for SQLite and PostgreSQL.
- [ ] Drift tests detect missing/extra columns and mismatched constraints.
- [ ] Dry-run has no database side effects.

## [ ] ARSY-043 — Build versioned migration lifecycle and recovery
**Depends on:** ARSY-042
**Priority:** P0 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Support migration dependencies, checksums, applied state, retries and per-tenant progress.
- [ ] Define expand/contract strategy for zero-downtime compatible changes.
- [ ] Define backup and recovery requirements for destructive migrations.

### DESIGN
- [ ] Implement migration graph validation and ordered execution.
- [ ] Record execution logs and state transitions.
- [ ] Provide CLI and API status/reporting.

### RULE
- [ ] Applied migration content is immutable.
- [ ] Failed migrations must leave a recoverable state.
- [ ] Destructive migrations require explicit approval and tested backup.

### ACCEPTANCE
- [ ] Tests cover dependency cycles, checksum drift, retries and partial failure.
- [ ] Migration replay is idempotent where specified.
- [ ] Tenant migration status is independently queryable.

## [ ] ARSY-044 — Complete PostgreSQL production adapter and isolation strategy
**Depends on:** ARSY-040, ARSY-041, ARSY-042
**Priority:** P1 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Choose and document a supported isolation model: database-per-tenant, schema-per-tenant or shared schema with enforced RLS.
- [ ] Define connection pooling, migrations, backups, quotas and operational limits.
- [ ] Keep SQLite as a supported development/single-node adapter if it meets product requirements.

### DESIGN
- [ ] Implement the selected isolation strategy end-to-end, not just a connection pool.
- [ ] Add provisioning and tenant context integration.
- [ ] Add real PostgreSQL integration environment in CI.

### RULE
- [ ] No claim of PostgreSQL multi-tenancy until adversarial isolation tests pass.
- [ ] RLS context must be transaction-local and cannot leak across pooled connections.
- [ ] Superuser bypass must be explicit, restricted and audited.

### ACCEPTANCE
- [ ] Cross-tenant read/write tests prove isolation under concurrent pooled connections.
- [ ] Provision, migrate, backup and restore work for a tenant.
- [ ] PostgreSQL adapter support matrix and operational guide are published.

---

# PHASE 5 — Security and enterprise tenancy

## [ ] ARSY-050 — Enforce server-derived model and field authorization
**Depends on:** ARSY-030, ARSY-035
**Priority:** P0 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Define model CRUD permissions, field read/write permissions and role/policy evaluation.
- [ ] Define server-side field projection for get/list/query/export.
- [ ] Define system-field ownership and explicit administrative bypass.

### DESIGN
- [ ] Build authorization context from authenticated principal, tenant, model and operation.
- [ ] Resolve allowed fields from trusted registered metadata and policy only.
- [ ] Apply field projection in repository/query layer before serialization.

### RULE
- [ ] Client-supplied allowlists never grant permission.
- [ ] Default deny; unknown model, field, operation or policy fails closed.
- [ ] Restricted reads never fall back to `SELECT *`.

### ACCEPTANCE
- [ ] Tests prove clients cannot grant themselves fields via request DTOs.
- [ ] Get, list, query, export and relation expansion all enforce field policy.
- [ ] Security regression tests cover unauthorized read/write and mass assignment.

## [ ] ARSY-051 — Enforce record-level policies and tenant isolation
**Depends on:** ARSY-041, ARSY-044, ARSY-050
**Priority:** P0 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Define row-level scope rules for read, update and delete.
- [ ] Define tenant context resolution for API, jobs, events, imports and admin tasks.
- [ ] Define audited cross-tenant administration.

### DESIGN
- [ ] Apply row policy in query AST/repository execution, not only HTTP middleware.
- [ ] Integrate adapter-level isolation where supported.
- [ ] Add explicit privileged system context.

### RULE
- [ ] Missing tenant context is an error in tenant-scoped operations.
- [ ] Never trust tenant identifiers from request bodies as authorization.
- [ ] Privileged bypass requires capability and audit reason.

### ACCEPTANCE
- [ ] Adversarial tests attempt cross-tenant CRUD, query, relation expansion and background jobs.
- [ ] No tenant data leaks through counts, errors, logs or pagination metadata.
- [ ] Concurrency tests prove tenant context cannot leak between pooled requests.

## [ ] ARSY-052 — Add audit trail, secret handling and security diagnostics
**Depends on:** ARSY-050, ARSY-051
**Priority:** P1 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Define immutable audit events for authentication, authorization, data mutation, module changes and admin operations.
- [ ] Define secret storage, rotation, redaction and diagnostic export.
- [ ] Define retention and access policies.

### DESIGN
- [ ] Add structured audit event schema and append-only persistence.
- [ ] Add correlation IDs across HTTP, jobs, hooks and migrations.
- [ ] Add redaction at logging and export boundaries.

### RULE
- [ ] Never log passwords, tokens, encryption keys or raw secret-bearing payloads.
- [ ] Audit writes cannot be modified by ordinary application users.
- [ ] Audit failure policy must be explicit for security-critical operations.

### ACCEPTANCE
- [ ] Tests prove redaction and append-only behavior.
- [ ] Critical operations emit actor, tenant, target, action, outcome and correlation ID.
- [ ] Security diagnostics can be shared without leaking secrets.

---

# PHASE 6 — Query engine and API generation

## [ ] ARSY-060 — Complete validated query AST and dialect compiler
**Depends on:** ARSY-030, ARSY-040, ARSY-050
**Priority:** P1 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Support filters, boolean groups, sorting, projection, pagination and relation predicates.
- [ ] Define typed operators by field type.
- [ ] Define query limits and complexity budget.

### DESIGN
- [ ] Parse external query input into a validated AST.
- [ ] Compile AST using adapter dialect and bound parameters.
- [ ] Apply authorization projection and row scope before execution.

### RULE
- [ ] No raw SQL from untrusted clients.
- [ ] Invalid operators, fields and sort expressions are rejected.
- [ ] Every query has bounded pagination and execution budget.

### ACCEPTANCE
- [ ] Fuzz tests cover parser, AST validation and compiler.
- [ ] Golden SQL tests cover SQLite and PostgreSQL.
- [ ] Security tests cover injection, unauthorized fields and tenant scope.

## [ ] ARSY-061 — Implement relation-aware querying and eager loading
**Depends on:** ARSY-032, ARSY-060
**Priority:** P1 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Support relation filters, nested projections, eager loading and bounded expansion.
- [ ] Prevent N+1 behavior for common list/detail workloads.
- [ ] Define cycle detection and maximum expansion depth.

### DESIGN
- [ ] Add relation planner and batched loader.
- [ ] Integrate relation policy at each traversed model.
- [ ] Add query-count instrumentation.

### RULE
- [ ] Every relation traversal re-checks tenant and field/record permissions.
- [ ] Expansion depth, row count and query count are bounded.
- [ ] Never reveal related records through unauthorized relation metadata.

### ACCEPTANCE
- [ ] Tests cover nested relations, cycles, forbidden relations and query budgets.
- [ ] Benchmarks demonstrate bounded query count for list/detail flows.
- [ ] Both supported adapters pass relation contract tests.

## [ ] ARSY-062 — Generate secure CRUD APIs and OpenAPI from model metadata
**Depends on:** ARSY-035, ARSY-050, ARSY-060
**Priority:** P1 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Generate versioned REST CRUD endpoints from registered models.
- [ ] Generate OpenAPI schemas, typed errors, pagination and auth requirements.
- [ ] Support custom service endpoints without bypassing core policies.

### DESIGN
- [ ] Implement route registration from model descriptors and API policy metadata.
- [ ] Reuse the same repository and validation pipeline as internal services.
- [ ] Generate OpenAPI deterministically.

### RULE
- [ ] Generated endpoints cannot bypass authorization, validation or tenant routing.
- [ ] No automatic exposure of internal/system fields.
- [ ] API compatibility changes require versioning and changelog entries.

### ACCEPTANCE
- [ ] New model manifest produces working CRUD API without hand-written routes.
- [ ] Contract tests compare runtime responses with generated OpenAPI.
- [ ] Security tests cover every generated CRUD operation.

## [ ] ARSY-063 — Add API keys, quotas, rate limits and webhook delivery
**Depends on:** ARSY-051, ARSY-062
**Priority:** P2 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Support scoped API keys, rotation, revocation and usage tracking.
- [ ] Support tenant/user/model quotas and configurable rate limits.
- [ ] Support signed, retryable and observable webhooks.

### DESIGN
- [ ] Store only hashed API key material and show plaintext only once.
- [ ] Implement rate-limit abstraction with safe local fallback.
- [ ] Implement webhook outbox, retries, signatures and delivery logs.

### RULE
- [ ] API keys are tenant- and scope-bound.
- [ ] Webhook retries are idempotent and secrets are redacted.
- [ ] Rate limits cannot be bypassed by changing untrusted headers.

### ACCEPTANCE
- [ ] Tests cover key lifecycle, scope enforcement and quota boundaries.
- [ ] Webhook signature and retry tests pass.
- [ ] Usage and rate-limit telemetry are queryable per tenant.

---

# PHASE 7 — Manifest-driven add-on and application runtime

## [ ] ARSY-070 — Define the A-R-S-Y module manifest specification
**Depends on:** ARSY-002, ARSY-030
**Priority:** P2 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Define manifest schema for module identity, version, dependencies, models, fields, relations, permissions, views, actions, workflows, seeds and migrations.
- [ ] Define SemVer compatibility and manifest schema evolution.
- [ ] Define deterministic validation errors.

### DESIGN
- [ ] Publish JSON Schema and Go typed structures.
- [ ] Implement parser, validator and canonical serializer.
- [ ] Provide example manifests for a minimal app and a CRM module.

### RULE
- [ ] Unknown critical manifest keys fail validation.
- [ ] Manifests cannot execute arbitrary code or raw SQL during validation.
- [ ] Version and dependency resolution must be deterministic.

### ACCEPTANCE
- [ ] Valid/invalid fixture suite covers every manifest section.
- [ ] Same manifest always produces the same canonical fingerprint.
- [ ] Manifest reference and authoring guide are published.

## [ ] ARSY-071 — Implement secure module registry and lifecycle
**Depends on:** ARSY-070, ARSY-043
**Priority:** P2 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Support discover, validate, install, enable, disable, upgrade, rollback and uninstall.
- [ ] Resolve module dependencies and detect conflicts.
- [ ] Track installed module versions per app/tenant.

### DESIGN
- [ ] Implement module registry, dependency resolver and lifecycle orchestrator.
- [ ] Integrate migrations, model registration, permissions and seeds as ordered lifecycle stages.
- [ ] Add dry-run install/upgrade plan.

### RULE
- [ ] Module lifecycle is transactional or explicitly resumable.
- [ ] Uninstall never silently deletes business data.
- [ ] Module capabilities are least-privilege and auditable.

### ACCEPTANCE
- [ ] Install/upgrade/rollback tests cover dependency and failure cases.
- [ ] Duplicate installs are idempotent.
- [ ] Module state remains consistent after interrupted lifecycle operations.

## [ ] ARSY-072 — Implement declarative views, actions and menu runtime
**Depends on:** ARSY-070, ARSY-010
**Priority:** P2 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Define view types: form, list, detail, kanban, calendar and dashboard.
- [ ] Define actions, menus, navigation, visibility rules and role access.
- [ ] Define safe declarative expressions and supported UI components.

### DESIGN
- [ ] Add metadata validators and runtime registries for views/actions.
- [ ] Render declarative metadata through A-R-S-Y Studio component registry.
- [ ] Bind actions to authorized API operations.

### RULE
- [ ] No arbitrary client script execution from untrusted module manifests.
- [ ] UI visibility is not a substitute for backend authorization.
- [ ] Unsupported view/action metadata fails installation.

### ACCEPTANCE
- [ ] Sample module produces navigation, list, form and detail screens from metadata.
- [ ] Permission-aware UI tests cover hidden and forbidden actions.
- [ ] Manifest compatibility tests cover view/action upgrades.

## [ ] ARSY-073 — Implement module seed data and deterministic app bootstrap
**Depends on:** ARSY-071, ARSY-033
**Priority:** P2 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Support ordered seed sets, references between seeded records and idempotent reruns.
- [ ] Support base system data and optional module demo data.
- [ ] Define seed ownership and upgrade behavior.

### DESIGN
- [ ] Implement seed runner through repository/UoW pipeline.
- [ ] Track seed version/checksum and execution state.
- [ ] Add preview and dry-run mode.

### RULE
- [ ] Seed data must pass normal authorization/validation rules in a controlled system context.
- [ ] Re-running a seed cannot create duplicates.
- [ ] Demo data must never be installed silently in production.

### ACCEPTANCE
- [ ] Fresh app bootstrap installs base metadata and seed data.
- [ ] Repeated seed execution is idempotent.
- [ ] Failure recovery does not leave partially seeded references.

---

# PHASE 8 — Full-stack App Builder and developer experience

## [ ] ARSY-080 — Build metadata-driven form and field component system
**Depends on:** ARSY-010, ARSY-031, ARSY-072
**Priority:** P2 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Map every supported field type to a UI component and editor configuration.
- [ ] Support required, readonly, hidden, conditional, help text and validation states.
- [ ] Support accessible keyboard and mobile interactions.

### DESIGN
- [ ] Create a field component registry and form renderer.
- [ ] Generate typed form state from model metadata.
- [ ] Reuse server validation error format.

### RULE
- [ ] Frontend metadata cannot grant backend permissions.
- [ ] Components must be accessible and responsive.
- [ ] Custom components must declare capabilities and version compatibility.

### ACCEPTANCE
- [ ] A model manifest renders a usable create/edit form automatically.
- [ ] All built-in field types have component tests.
- [ ] Accessibility and responsive smoke tests pass.

## [ ] ARSY-081 — Build generic list, search, filter and record detail views
**Depends on:** ARSY-060, ARSY-062, ARSY-080
**Priority:** P2 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Support configurable columns, sorting, filtering, saved views and pagination.
- [ ] Support record detail, related records and bulk actions.
- [ ] Respect field and row-level authorization.

### DESIGN
- [ ] Build generic data table and query-builder components.
- [ ] Bind all data access through generated typed API client.
- [ ] Support cursor pagination for large datasets.

### RULE
- [ ] Never fetch unrestricted fields and hide them only in the browser.
- [ ] Bulk actions must be authorized per record.
- [ ] Query UI must enforce server-defined complexity limits.

### ACCEPTANCE
- [ ] New model appears in list/detail UI from metadata alone.
- [ ] Search, filter, sort and pagination tests pass.
- [ ] Restricted fields never appear in network payloads.

## [ ] ARSY-082 — Build kanban, calendar, dashboard and workflow UI primitives
**Depends on:** ARSY-072, ARSY-081
**Priority:** P3 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Support configurable kanban grouping, calendar date mapping and dashboard widgets.
- [ ] Support workflow transitions and action feedback.
- [ ] Define widget capability and permission metadata.

### DESIGN
- [ ] Add reusable view renderer and widget registry.
- [ ] Bind transitions to server-side workflow/action APIs.
- [ ] Add loading, error and empty states.

### RULE
- [ ] Client-side state cannot directly mutate protected workflow fields.
- [ ] Widgets must declare data requirements and permissions.
- [ ] Avoid unbounded dashboard queries.

### ACCEPTANCE
- [ ] Example module renders kanban, calendar and dashboard views.
- [ ] Workflow transition tests cover valid and forbidden transitions.
- [ ] UI tests verify bounded requests and permission-aware widgets.

## [ ] ARSY-083 — Create typed SDK and application developer tooling
**Depends on:** ARSY-062, ARSY-070
**Priority:** P2 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Generate TypeScript and Go clients from versioned API/OpenAPI contracts.
- [ ] Provide typed auth, query, CRUD, upload and realtime helpers.
- [ ] Define SDK compatibility and deprecation policy.

### DESIGN
- [ ] Add deterministic SDK generator and checked-in example output.
- [ ] Publish package usage examples and local development workflow.
- [ ] Add `arsy app init`, `arsy module init` and `arsy model generate`.

### RULE
- [ ] SDK must not bypass API authorization or encode secrets in generated source.
- [ ] Generated artifacts must be reproducible.
- [ ] Breaking SDK changes require explicit versioning.

### ACCEPTANCE
- [ ] Generated clients compile and pass contract tests.
- [ ] Quickstart creates an app and performs authenticated CRUD.
- [ ] CLI scaffolds a module that passes validation and tests.

## [ ] ARSY-084 — Add deterministic manifest/code generation and AI assistant boundary
**Depends on:** ARSY-070, ARSY-083
**Priority:** P3 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Generate model, view, permission, API and test scaffolds from validated manifests.
- [ ] Allow AI assistance to draft, explain and repair manifests.
- [ ] Define human approval for destructive or security-sensitive changes.

### DESIGN
- [ ] Implement deterministic generator templates and validation pipeline.
- [ ] Put AI behind a provider-neutral interface that returns proposed manifest changes.
- [ ] Run generated output through formatter, validators, tests and security policy.

### RULE
- [ ] AI never executes arbitrary SQL, deploys code or bypasses policy validation.
- [ ] Same manifest and generator version produce byte-stable output.
- [ ] Human approval is required for destructive schema or permission changes.

### ACCEPTANCE
- [ ] Generated module compiles and passes tests without manual edits.
- [ ] Invalid AI suggestions are rejected with actionable diagnostics.
- [ ] Golden tests prove deterministic generation.

---

# PHASE 9 — Tenant lifecycle and BaaS operations

## [ ] ARSY-090 — Implement durable tenant provisioning lifecycle
**Depends on:** ARSY-041, ARSY-043, ARSY-044
**Priority:** P1 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Support create, provision, activate, suspend, resume, backup, restore and delete/retention workflows.
- [ ] Track provisioning state, retries and operator intervention.
- [ ] Support tenant plans, quotas and feature entitlements.

### DESIGN
- [ ] Implement durable provisioning state machine and outbox/worker.
- [ ] Make database creation, migrations and seed execution resumable.
- [ ] Expose tenant lifecycle status in Studio and CLI.

### RULE
- [ ] Provisioning is idempotent and auditable.
- [ ] Tenant deletion follows retention and explicit approval policy.
- [ ] No tenant is activated before storage, schema and security checks pass.

### ACCEPTANCE
- [ ] Tests cover retries, partial failure and reconciliation.
- [ ] Tenant create-to-ready flow is observable end-to-end.
- [ ] Suspend/resume and backup/restore preserve isolation.

## [ ] ARSY-091 — Add usage metering, quotas and plan enforcement
**Depends on:** ARSY-063, ARSY-090
**Priority:** P3 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Define usage dimensions: API requests, storage, files, bandwidth, background jobs and active users.
- [ ] Define plan limits, grace behavior and overage policy.
- [ ] Define exportable usage reports.

### DESIGN
- [ ] Implement append-only usage events and periodic aggregation.
- [ ] Enforce quotas at service boundaries with configurable policy.
- [ ] Add plan/entitlement resolver.

### RULE
- [ ] Usage accounting must be tenant-scoped and tamper-evident.
- [ ] Quota enforcement must not create partial writes.
- [ ] Billing integration is optional and must not be coupled to core ORM transactions.

### ACCEPTANCE
- [ ] Usage tests cover aggregation, concurrency and quota boundaries.
- [ ] Tenant usage report reconciles with source events.
- [ ] Plan changes apply predictably and are audited.

## [ ] ARSY-092 — Implement backup, restore and disaster recovery
**Depends on:** ARSY-043, ARSY-090
**Priority:** P1 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Support tenant and platform backups, retention, encryption and integrity verification.
- [ ] Define RPO/RTO targets per deployment profile.
- [ ] Support restore to isolated environment before promotion.

### DESIGN
- [ ] Use adapter-specific backup interfaces behind a common lifecycle.
- [ ] Add scheduled backup worker and restore verification.
- [ ] Document recovery and data portability procedures.

### RULE
- [ ] Backup is not considered valid until restore verification passes.
- [ ] Backup encryption keys are managed outside backup payloads.
- [ ] Restore must preserve tenant identity and security boundaries.

### ACCEPTANCE
- [ ] Automated backup/restore integration tests pass for supported adapters.
- [ ] Recovery drills record measured RPO/RTO.
- [ ] Backup status and failures are visible in Studio and CLI.

---

# PHASE 10 — Quality, compatibility, packaging and release

## [ ] ARSY-100 — Establish full engine contract, integration and fuzz test suites
**Depends on:** ARSY-035, ARSY-044, ARSY-051, ARSY-062
**Priority:** P0 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Define test pyramid for unit, adapter contract, integration, security, browser E2E, migration and performance.
- [ ] Define critical golden paths and supported compatibility matrix.
- [ ] Define deterministic test data and isolated test environments.

### DESIGN
- [ ] Add reusable contract suites for repository, adapter, auth and module lifecycle.
- [ ] Add fuzz targets for manifests, query AST and migration planner.
- [ ] Add Playwright E2E for Studio and installer smoke tests.

### RULE
- [ ] Never weaken tests or skip security tests to make CI green.
- [ ] No flaky network-dependent unit tests.
- [ ] Every bug fix adds a regression test.

### ACCEPTANCE
- [ ] CI runs required suites on every pull request.
- [ ] Race detector and fuzz smoke tests run on supported packages.
- [ ] Critical golden path passes from fresh install through app/module CRUD.

## [ ] ARSY-101 — Add performance baselines and regression budgets
**Depends on:** ARSY-100
**Priority:** P2 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Benchmark CRUD, filters, relations, transactions, startup, tenant routing and Studio load.
- [ ] Define realistic data volumes and concurrency profiles.
- [ ] Define regression thresholds based on measured baselines.

### DESIGN
- [ ] Add reproducible Go benchmarks and browser performance scripts.
- [ ] Store benchmark metadata and compare against main branch.
- [ ] Profile hot paths before optimization.

### RULE
- [ ] Do not introduce speculative caching or ML optimization without benchmark evidence.
- [ ] Performance tests must use representative workloads.
- [ ] Correctness and tenant isolation take priority over micro-optimizations.

### ACCEPTANCE
- [ ] Baselines are published with hardware/runtime context.
- [ ] CI flags material regressions.
- [ ] Performance reports identify p50/p95/p99, throughput and allocations where relevant.

## [ ] ARSY-102 — Verify and complete A-R-S-Y documentation and quickstart
**Depends on:** ARSY-011, ARSY-022, ARSY-062
**Priority:** P2 · **Size:** S · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Rewrite README, quickstart, API docs, configuration reference and operator guide around A-R-S-Y.
- [ ] Remove obsolete PocketBase product setup instructions while retaining required attribution/legal notices.
- [ ] Document install, upgrade, backup, restore, app/module authoring and security model.

### DESIGN
- [ ] Generate API reference from OpenAPI and CLI reference from command definitions where possible.
- [ ] Maintain one canonical quickstart and versioned docs.
- [ ] Add examples that are tested in CI.

### RULE
- [ ] Docs must describe current behavior, not planned features as implemented.
- [ ] Do not duplicate the same source of truth across multiple docs.
- [ ] All commands and code snippets must be reproducible.

### ACCEPTANCE
- [ ] Fresh-user quickstart succeeds from a clean environment.
- [ ] Documentation links and code examples are checked in CI.
- [ ] No obsolete Superuser UI/installer steps remain in user docs.

## [ ] ARSY-103 — Harden release artifacts and software supply chain
**Depends on:** ARSY-021, ARSY-022, ARSY-100
**Priority:** P1 · **Size:** M · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Define release binaries, container images, checksums, SBOM, provenance and supported upgrade paths.
- [ ] Define dependency update and vulnerability response process.
- [ ] Define signed release verification.

### DESIGN
- [ ] Build reproducible release pipeline for supported OS/architectures.
- [ ] Generate SBOM and provenance attestations.
- [ ] Add container and binary smoke tests.

### RULE
- [ ] Never publish secrets or private build artifacts.
- [ ] Release artifacts must be traceable to source commit and CI run.
- [ ] Preserve upstream and third-party license notices in distributions.

### ACCEPTANCE
- [ ] Clean release candidate installs and starts on supported targets.
- [ ] Checksums, SBOM and provenance are published with artifacts.
- [ ] Upgrade and rollback tests pass before stable release.

## [ ] ARSY-104 — Prove no PocketBase product surface remains in release artifacts
**Depends on:** ARSY-010, ARSY-011, ARSY-012, ARSY-013, ARSY-021, ARSY-022, ARSY-024
**Priority:** P1 · **Size:** S · **Evidence:** PR description + CI/test artifacts
**Non-goals:** No unrelated refactor, duplicate implementation, or destructive rewrite of upstream-derived internals.

### SPEC
- [ ] Audit shipped routes, UI bundles, installer, commands, settings, templates, assets and user-facing strings.
- [ ] Separate user-facing product surfaces from internal upstream-derived implementation code.
- [ ] Record remaining legacy surfaces with reproduction steps and remediation links.

### DESIGN
- [ ] Use dependency graph, route enumeration, source scans, binary/package inspection and runtime tests as independent evidence.
- [ ] Maintain a legal/internal attribution allowlist with rationale.
- [ ] Remove or disable only confirmed obsolete entry points in small changes.

### RULE
- [ ] Target no PocketBase product experience; do not rewrite or erase every upstream-derived line of code.
- [ ] Never blindly remove licenses, security fixes, migration history or required infrastructure.
- [ ] No silent breaking change to stored data or supported A-R-S-Y APIs.

### ACCEPTANCE
- [ ] No PocketBase Superuser UI, setup wizard, installer, CLI identity or user-facing route is shipped or reachable.
- [ ] A-R-S-Y web, installer and CLI are the only supported product experiences.
- [ ] Package/runtime scans pass with only documented legal/internal allowlist entries.
- [ ] Clean build, install smoke test, Studio E2E and regression suites pass.
- [ ] Audit report links evidence for every checked surface.

## Execution order

1. **Baseline and architecture:** ARSY-001 to ARSY-003 — inventory the transformed fork and define A-R-S-Y ownership boundaries.
2. **Verify existing product surfaces:** ARSY-010 to ARSY-024 — prove Studio, setup, installer, CLI and startup are A-R-S-Y-native; fix only evidenced gaps.
3. **Core engine:** ARSY-030 to ARSY-044 — canonical metadata, ORM, persistence, adapters and migrations.
4. **Security gates:** ARSY-050 to ARSY-052 — enforce model/field/record authorization and tenant isolation.
5. **Query and application platform:** ARSY-060 to ARSY-073 — query engine, API generation, manifests, module lifecycle and seed data.
6. **Full-stack developer experience:** ARSY-080 to ARSY-084 — metadata-driven UI, SDK, scaffolding and bounded AI assistance.
7. **BaaS operations:** ARSY-090 to ARSY-092 — provisioning, metering, backup and recovery.
8. **Release proof:** ARSY-100 to ARSY-104 and the Golden path — integration, performance, docs, supply chain and proof that no legacy product surface ships.

**Explicitly deferred until justified:** Kafka, Kubernetes operator, ML-based query optimizer, predictive analytics, event sourcing, GraphQL, service mesh, and large ERP module suite. Build the deterministic secure engine and complete first-party product experience first.
