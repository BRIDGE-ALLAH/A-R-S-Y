# A-R-S-Y Release Gates & Transformation Scorecard

**Task:** ARSY-003 — Define release gates and transformation scorecard  
**Status:** In Progress  
**Depends on:** ARSY-002 (completed)  
**Generated:** 2026-11-10

---

## 1. Measurable Completion Criteria

### 1.1 Engine Completeness (Layer 1-3)

| Component | Criteria | Verification |
|-----------|----------|--------------|
| **Model Descriptor** | Round-trip deterministic; invalid metadata rejected | `arsy/engine/manifest` tests + golden files |
| **Field Types** | All 22 built-in types validated; defaults normalized | `core/field*_test.go` table-driven |
| **Relations** | 4 types (belongs-to, has-one, has-many, M2M); cascade policies work | `arsy/engine/relations` contract tests |
| **Validation Pipeline** | Order: authorize → normalize → defaults → validate → persist → events | `arsy/engine/validation` + integration tests |
| **Unit of Work** | TX-scoped CRUD; nested calls no accidental commits; hooks ordered | `arsy/engine/uow` tests |
| **Repository Mapper** | Typed CRUD, query, count, exists, bulk, upsert, soft-delete | Contract tests per adapter |
| **Database Adapters** | SQLite + PostgreSQL pass same contract suite | CI matrix: sqlite, pg |
| **Schema Planner** | Deterministic plans; dry-run no side effects; drift detection | Golden tests for SQLite/PG |
| **Migration Executor** | Dependencies, checksums, retries, per-tenant progress | `arsy/aeng/migration_test.go` |
| **Model/Field Auth** | Default deny; projection before serialization; no client allowlists | `arsy/aeng/security_test.go` + adversarial tests |
| **Record Policies** | Row-level scope in query AST; tenant context resolution; audited bypass | Adversarial cross-tenant tests |
| **Audit Trail** | Immutable events; correlation IDs; redaction; append-only | `arsy/aeng/audit_test.go` |

### 1.2 Web & Studio (Layer 6)

| Component | Criteria | Verification |
|-----------|----------|--------------|
| **Studio Routes** | All admin workflows served (models, records, tenants, modules, permissions, migrations, health) | Playwright E2E |
| **No Legacy UI** | Zero PocketBase Superuser UI bundle/routes shipped/reachable | Binary scan + route enumeration |
| **Identity** | A-R-S-Y branding on all user-facing surfaces | Brand scan CI |
| **Responsive/Accessible** | Mobile/desktop; keyboard nav; ARIA labels | Axe-core + visual regression |

### 1.3 Installer & Bootstrap (Layer 0→1)

| Component | Criteria | Verification |
|-----------|----------|--------------|
| **Clean Install** | Fresh data dir → owner + first app via A-R-S-Y APIs/UI | E2E test |
| **Resume/Recovery** | Interrupted setup resumes; no duplicate owners | Concurrency tests |
| **Unattended Mode** | Non-interactive install with env vars | CI test |
| **Upgrade** | In-place binary swap; migrations run; data preserved | Upgrade matrix test |
| **Rollback** | Failed upgrade reverts cleanly | Rollback test |
| **Backup/Restore** | Tenant + platform backups; integrity verified | `arsy/tenancy/backup_test.go` |

### 1.4 CLI (Layer 7)

| Component | Criteria | Verification |
|-----------|----------|--------------|
| **Command Tree** | `serve`, `install`, `doctor`, `config`, `owner`, `app`, `module`, `model`, `migration`, `tenant`, `backup`, `version` | `arsy --help` test |
| **JSON Output** | All commands support `--json`; stable exit codes | Contract tests |
| **No PocketBase CLI** | Zero PocketBase-branded commands/output | Binary scan |
| **Shell Completion** | `bash`/`zsh`/`fish` completion generates | Completion test |

### 1.5 Security & Tenancy (Layer 3)

| Component | Criteria | Verification |
|-----------|----------|--------------|
| **Tenant Isolation** | Cross-tenant CRUD/query/relations/background jobs fail | Adversarial test suite |
| **No Data Leaks** | Counts, errors, logs, pagination metadata never leak tenant data | Leak detection tests |
| **RLS (PG)** | Transaction-local context; no pool leakage; superuser bypass audited | PG isolation tests |
| **Secret Handling** | No passwords/tokens/keys in logs; redaction at boundaries | Security diagnostics test |

### 1.6 Developer Experience (Layer 5,7)

| Component | Criteria | Verification |
|-----------|----------|--------------|
| **Module Lifecycle** | Install/upgrade/rollback idempotent; dependency resolution; state consistent | `arsy/engine/manifest` tests |
| **Seed Data** | Idempotent reruns; no duplicates; demo data not in prod | `arsy/aeng/seed_test.go` |
| **SDK Generation** | TS/Go clients from OpenAPI; compile + contract tests | Generated client CI |
| **Code Generation** | `arsy app init`, `module init`, `model generate` produce working code | Scaffold test |

---

## 2. PocketBase-Facing vs Internal Upstream Tracking

| Category | PocketBase-Facing (Must Eliminate) | Internal Upstream (Can Retain) |
|----------|-----------------------------------|-------------------------------|
| **User-Facing Routes** | `/_/`, `/api/collections/`, `/api/records/`, `/pbinstall/` | — |
| **UI Bundles** | PocketBase Superuser Vue app | A-R-S-Y Studio Vue app |
| **CLI Commands** | `superuser`, `migrate` (PB-style) | `arsy` command tree |
| **Terminology** | "PocketBase", "Superuser", "pb_data", "_superusers" | Internal identifiers only |
| **Config Keys** | `PB_*` env vars | `ARSY_*` env vars |
| **Error Messages** | "PocketBase" in user-visible errors | Internal debug logs only |
| **Migrations** | Upstream 1717233556-1717233559 (preserve for compat) | A-R-S-Y migrations (new) |
| **Database Schema** | `_collections`, `_params`, `_superusers` tables | A-R-S-Y tables (`arsy_*`) |

**Tracking Method:** CI brand scan (`grep -r "PocketBase\|Superuser\|pb_data\|_superusers" --include="*.go" --include="*.js" --include="*.html" | grep -v "test\|internal\|license"`)

---

## 3. Supported Platform Matrix

| Dimension | Minimum Supported | Tested In CI |
|-----------|------------------|--------------|
| **OS** | Linux (glibc 2.31+), macOS 12+, Windows 10+ | ubuntu-latest, macos-latest, windows-latest |
| **SQLite** | 3.38+ (modernc.org/sqlite v1.30+) | modernc.org/sqlite latest |
| **PostgreSQL** | 14+ (pgx v5) | PostgreSQL 16 |
| **Go Toolchain** | 1.22+ | Go 1.22, 1.23 |
| **Node.js** | 20+ (for UI build) | Node 20 LTS |
| **Browser** | Chrome 120+, Firefox 115+, Safari 17+, Edge 120+ | Playwright matrix |

---

## 4. CI-Generated Scorecard

Location: `docs/architecture/release-scorecard.md` (generated at `make scorecard`)

```markdown
# A-R-S-Y Release Scorecard — {{VERSION}} — {{DATE}}

## Engine (Layer 1-3)
| Component | Status | Evidence |
|-----------|--------|----------|
| Model Descriptor | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| Field Types (22) | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| Relations (4 types) | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| Validation Pipeline | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| Unit of Work | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| Repository Mapper (SQLite) | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| Repository Mapper (PostgreSQL) | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| Schema Planner | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| Migration Executor | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| Model/Field Auth | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| Record Policies | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| Audit Trail | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |

## Web & Studio (Layer 6)
| Component | Status | Evidence |
|-----------|--------|----------|
| Studio Workflows | {{PASS/FAIL/NOT RUN/BLOCKED}} | [E2E Run]({{URL}}) |
| No Legacy UI | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Brand Scan]({{URL}}) |
| A-R-S-Y Identity | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Brand Scan]({{URL}}) |
| Accessibility | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Axe Report]({{URL}}) |

## Installer & Bootstrap
| Component | Status | Evidence |
|-----------|--------|----------|
| Clean Install E2E | {{PASS/FAIL/NOT RUN/BLOCKED}} | [E2E Run]({{URL}}) |
| Resume/Recovery | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| Unattended Install | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| Upgrade Matrix | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| Rollback | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| Backup/Restore | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |

## CLI
| Component | Status | Evidence |
|-----------|--------|----------|
| Command Tree Complete | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| JSON Output Stable | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| No PB CLI | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Brand Scan]({{URL}}) |
| Shell Completion | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |

## Security & Tenancy
| Component | Status | Evidence |
|-----------|--------|----------|
| Cross-Tenant Isolation | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Adversarial Tests]({{URL}}) |
| No Data Leaks | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Leak Tests]({{URL}}) |
| PG RLS Enforced | {{PASS/FAIL/NOT RUN/BLOCKED}} | [PG Tests]({{URL}}) |
| Secret Redaction | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Security Tests]({{URL}}) |

## Developer Experience
| Component | Status | Evidence |
|-----------|--------|----------|
| Module Lifecycle | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| Seed Idempotency | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Test Run]({{URL}}) |
| SDK Generation | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Contract Tests]({{URL}}) |
| Code Generation | {{PASS/FAIL/NOT RUN/BLOCKED}} | [Scaffold Tests]({{URL}}) |

## Overall
**GATE: {{PASS/FAIL}}** — All release-blocking checks must be PASS for release.
```

---

## 5. Release-Blocking CI Checks

| Check | Job | Blocks Release If |
|-------|-----|-------------------|
| **Go Formatting** | `gofmt -l` | Any non-conforming file |
| **Go Vet** | `go vet ./...` | Any warning |
| **Unit Tests** | `go test ./... -count=1` | Any FAIL |
| **Race Detector** | `go test -race ./arsy/...` | Any race |
| **Engine Contract Tests** | `go test ./arsy/engine/...` | Any FAIL |
| **Adapter Contract Tests** | `go test ./arsy/tenancy/...` (sqlite + pg) | Any FAIL |
| **Security Tests** | `go test ./arsy/security/...` | Any FAIL |
| **Adversarial Tenancy** | `go test ./arsy/tenancy/... -run Adversarial` | Any FAIL |
| **Migration Verification** | `make verify-migrations` | Any drift/failure |
| **UI Build** | `cd ui && npm run build` | Build error |
| **Playwright E2E** | `npx playwright test` | Any FAIL |
| **Brand Scan** | `make brand-scan` | Any PocketBase match (non-allowlist) |
| **Diagnostics** | `go run ./cmd/arsy doctor` | Any FAIL |
| **Standalone Binary** | `go build -o arsy .` | Build error |
| **License Check** | `go-licenses check` | Non-compliant license |

---

## 6. Acceptance Checklist (ARSY-003)

- [x] Measurable completion criteria for all 6 product areas defined
- [x] PocketBase-facing vs internal upstream tracking method documented
- [x] Minimum supported platform matrix defined
- [x] CI-generated scorecard template created
- [x] 15 release-blocking checks specified with job mapping
- [ ] Scorecard generation script implemented (`make scorecard`)
- [ ] CI job reports gate results on PRs
- [ ] PASS/FAIL/NOT RUN/BLOCKED distinction enforced

**Evidence location:** `docs/architecture/release-gates.md`