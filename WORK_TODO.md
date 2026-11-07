# A-R-S-Y — Complete Work Backlog
## Status: 2026-11-07 | Post-Sync with origin/main

---

## 📊 Repository Status

### Current State
- **Branch**: `master` → tracking `origin/main`
- **HEAD**: `b0a5b46` - feat(aeng): validate application relation metadata
- **Working Tree**: Clean ✓
- **Backup**: `backup-master-20261107` (old local commits)

### Active Pull Requests (3)
| PR # | Title | Status | Files |
|------|-------|--------|-------|
| 121 | Server management API + Console | OPEN (CI pending) | +736/-5 |
| 122 | SFTP upload + Job worker | OPEN (CI pending) | +973/-4 |
| 123 | Compose processor + Logs + Worker | OPEN (CI queued) | +429/-15 |

---

## 🎯 MVP-1 Task Plan (From `.plan/task-mvp-1.md`)

### Phase A: Foundation (Mostly Complete ✓)

#### ✅ Completed in Remote
- MVP1-002 to MVP1-018: Project setup, migrations, collections
- MVP1-019 to MVP1-022: Auth & onboarding
- MVP1-024 to MVP1-031: Collections & security
- MVP1-033 to MVP1-042: SSH & server readiness
- MVP1-045 to MVP1-050: Compose validation & state

#### ⏳ In Progress (PRs #121, #122, #123)
- MVP1-032: Sanitize server API responses
- MVP1-043: Server management API endpoints
- MVP1-044: Servers console screen
- MVP1-048: Secure SFTP upload
- MVP1-049: Validate Compose on target
- MVP1-051: Persistent deployment job claiming
- MVP1-052: Deployment worker lifecycle
- MVP1-053: Compose deployment execution
- MVP1-054: Persist deployment output
- MVP1-055: Deployment status & container inspection

---

### Phase B: Deployment Operations (TODO - High Priority)

#### 🔥 MVP1-056: Add deployment history and log API endpoints
**Status**: Not Started  
**Depends On**: PR #123 merged  
**Complexity**: Medium  
**Estimasi**: 4-6 jam

**Tasks**:
- [ ] Add `GET /api/arsy/v1/deployments` (list with filters)
- [ ] Add `GET /api/arsy/v1/deployments/:id` (single deployment)
- [ ] Add `GET /api/arsy/v1/deployments/:id/logs` (bounded log retrieval)
- [ ] Add `GET /api/arsy/v1/applications/:id/logs` (runtime container logs)
- [ ] Add pagination & filtering (by app, server, status, date range)
- [ ] Write API tests

**Files**:
- `arsy/httpapi/deployments.go` (new)
- `arsy/httpapi/deployments_test.go` (new)
- `arsy/httpapi/routes.go` (update)

---

#### 🔥 MVP1-057: Add safe stop and redeploy operations
**Status**: Not Started  
**Depends On**: MVP1-056  
**Complexity**: Medium  
**Estimasi**: 4-6 jam

**Tasks**:
- [ ] Add `POST /api/arsy/v1/deployments/:id/stop` (stop containers)
- [ ] Add `POST /api/arsy/v1/deployments/:id/redeploy` (redeploy same config)
- [ ] Implement `docker compose stop` executor
- [ ] Implement `docker compose restart` executor
- [ ] Add validation: prevent stop/redeploy on in-progress deployments
- [ ] Add concurrent operation locking per application
- [ ] Write operation tests

**Files**:
- `arsy/deployments/operations.go` (new)
- `arsy/deployments/operations_test.go` (new)
- `arsy/httpapi/deployments.go` (update)

---

#### 🔥 MVP1-058: Add configuration rollback
**Status**: Not Started  
**Depends On**: MVP1-057  
**Complexity**: Medium-High  
**Estimasi**: 6-8 jam

**Tasks**:
- [ ] Store each deployment's Compose manifest as immutable artifact
- [ ] Add `POST /api/arsy/v1/deployments/:id/rollback` endpoint
- [ ] Implement rollback: copy old manifest → create new deployment
- [ ] Add validation: can only rollback to successful deployments
- [ ] Add rollback history tracking
- [ ] Write rollback tests with mock deployments

**Files**:
- `arsy/deployments/rollback.go` (new)
- `arsy/deployments/rollback_test.go` (new)
- `arsy/migrations/mvp1.go` (update: add rollback tracking)
- `arsy/httpapi/deployments.go` (update)

**Acceptance**:
- Can rollback to any previous successful deployment
- Rollback creates new deployment record (preserves history)
- Cannot rollback during active deployment

---

### Phase C: Console UI Screens (TODO - High Priority)

#### 🎨 MVP1-059: Add the Overview screen
**Status**: Not Started  
**Depends On**: MVP1-056 (API)  
**Complexity**: Medium  
**Estimasi**: 4-6 jam

**Tasks**:
- [ ] Create `dashboard/src/pages/Overview.jsx`
- [ ] Add statistics cards:
  - Total servers (ready/total)
  - Total projects
  - Total applications
  - Recent deployments count
- [ ] Add recent deployments table (last 10)
- [ ] Add quick actions: "Add Server", "New Project", "Deploy App"
- [ ] Add auto-refresh every 30s
- [ ] Write component tests

**Files**:
- `dashboard/src/pages/Overview.jsx` (new)
- `dashboard/src/components/StatCard.jsx` (new)
- `dashboard/src/main.js` (update routing)

**Design**:
```
┌─────────────────────────────────────────┐
│  Overview                               │
├───────────┬───────────┬───────────┬─────┤
│ Servers   │ Projects  │ Apps      │Deploys│
│ 2/3 Ready │    5      │   12      │  34   │
├───────────┴───────────┴───────────┴─────┤
│  Recent Deployments                     │
│  ┌─────────────────────────────────┐   │
│  │ App       Server    Status  Time│   │
│  │ web-api   prod-01   Success 2m  │   │
│  │ worker    prod-02   Running 5m  │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

---

#### 🎨 MVP1-060: Add Project management UI
**Status**: Not Started  
**Depends On**: MVP1-059  
**Complexity**: Low-Medium  
**Estimasi**: 3-4 jam

**Tasks**:
- [ ] Create `dashboard/src/pages/Projects.jsx`
- [ ] Add projects list table (name, description, app count, created)
- [ ] Add "New Project" modal with form (name, description)
- [ ] Add "Edit Project" modal
- [ ] Add "Delete Project" with confirmation (only if no apps)
- [ ] Add project detail view (list apps in project)
- [ ] Wire up to PocketBase projects collection API

**Files**:
- `dashboard/src/pages/Projects.jsx` (new)
- `dashboard/src/components/ProjectForm.jsx` (new)
- `dashboard/src/main.js` (update routing)

---

#### 🎨 MVP1-061: Add Application creation UI
**Status**: Not Started  
**Depends On**: MVP1-060  
**Complexity**: Medium  
**Estimasi**: 5-6 jam

**Tasks**:
- [ ] Create `dashboard/src/pages/Applications.jsx`
- [ ] Add applications list table (name, project, server, status)
- [ ] Create "New Application" wizard:
  - Step 1: Select project (dropdown)
  - Step 2: Select server (only show ready servers)
  - Step 3: Enter app name & description
  - Step 4: Paste Docker Compose YAML
  - Step 5: Validate & review
- [ ] Add YAML editor with syntax highlighting
- [ ] Add Compose validation (client-side basic check)
- [ ] Wire up to applications collection API

**Files**:
- `dashboard/src/pages/Applications.jsx` (new)
- `dashboard/src/components/ApplicationWizard.jsx` (new)
- `dashboard/src/components/ComposeEditor.jsx` (new)
- `dashboard/src/main.js` (update routing)

---

#### 🎨 MVP1-062: Add the Deploy action and progress view
**Status**: Not Started  
**Depends On**: MVP1-061  
**Complexity**: Medium-High  
**Estimasi**: 6-8 jam

**Tasks**:
- [ ] Add "Deploy" button on application detail page
- [ ] Create deployment modal with:
  - Compose manifest preview (read-only)
  - Deployment confirmation
  - Environment variables input (optional)
- [ ] Add real-time deployment progress view:
  - State: queued → uploading → validating → deploying → running
  - Live log streaming
  - Progress indicator
- [ ] Add WebSocket or polling for live updates
- [ ] Handle deployment errors gracefully
- [ ] Add "View Logs" action

**Files**:
- `dashboard/src/pages/ApplicationDetail.jsx` (new)
- `dashboard/src/components/DeploymentModal.jsx` (new)
- `dashboard/src/components/DeploymentProgress.jsx` (new)
- `dashboard/src/hooks/useDeploymentStatus.js` (new)

---

#### 🎨 MVP1-063: Add Application operations UI
**Status**: Not Started  
**Depends On**: MVP1-062  
**Complexity**: Medium  
**Estimasi**: 5-6 jam

**Tasks**:
- [ ] Add application detail page sections:
  - Current container status (docker compose ps)
  - Published ports (external access links)
  - Resource usage (if available)
  - Deployment history
- [ ] Add operation buttons:
  - Stop (with confirmation)
  - Restart
  - Redeploy
  - Rollback (select from history)
  - View Logs
- [ ] Add container logs viewer with:
  - Service filter
  - Log level filter
  - Auto-scroll toggle
  - Download logs
- [ ] Add deployment history timeline

**Files**:
- `dashboard/src/pages/ApplicationDetail.jsx` (update)
- `dashboard/src/components/ContainerStatus.jsx` (new)
- `dashboard/src/components/LogViewer.jsx` (new)
- `dashboard/src/components/DeploymentHistory.jsx` (new)

---

#### 🎨 MVP1-064: Finish the onboarding path
**Status**: Not Started  
**Depends On**: All UI screens  
**Complexity**: Medium  
**Estimasi**: 4-5 jam

**Tasks**:
- [ ] Create onboarding wizard flow (multi-step):
  1. Welcome & intro
  2. Setup secrets key (if not configured)
  3. Add first server
  4. Verify SSH connection
  5. Check Docker/Compose readiness
  6. Create first project
  7. Deploy sample application (nginx hello-world)
- [ ] Add onboarding progress tracking
- [ ] Add "Skip" and "Resume later" options
- [ ] Add sample Compose file templates
- [ ] Show onboarding on first login
- [ ] Add "Restart Onboarding" in settings

**Files**:
- `dashboard/src/pages/Onboarding.jsx` (new)
- `dashboard/src/components/OnboardingStep.jsx` (new)
- `dashboard/src/data/sampleCompose.js` (new)
- `arsy/httpapi/onboarding.go` (update: track completion)

---

### Phase D: Testing (TODO - Medium Priority)

#### 🧪 MVP1-065: Add unit tests for A-R-S-Y config and security
**Status**: Not Started  
**Complexity**: Low  
**Estimasi**: 3-4 jam

**Tasks**:
- [ ] Test config defaults and validation
- [ ] Test invalid configuration handling
- [ ] Test secret encryption/decryption cycle
- [ ] Test tamper detection (modified ciphertext)
- [ ] Test key rotation scenarios
- [ ] Test credential redaction in logs
- [ ] Achieve >80% coverage for `arsy/config` and `arsy/security`

**Files**:
- `arsy/config/config_test.go` (expand)
- `arsy/security/encryption_test.go` (expand)
- `arsy/security/key_test.go` (new)

---

#### 🧪 MVP1-066: Add SSH tests using a local test SSH server
**Status**: Not Started  
**Complexity**: Medium  
**Estimasi**: 4-6 jam

**Tasks**:
- [ ] Set up embedded test SSH server (using `gliderlabs/ssh`)
- [ ] Generate deterministic test SSH keys
- [ ] Test host-key verification (accept/reject)
- [ ] Test authentication (key-based)
- [ ] Test command execution and output capture
- [ ] Test connection timeout handling
- [ ] Test SFTP upload/download
- [ ] Achieve >80% coverage for `arsy/servers`

**Files**:
- `arsy/servers/testing/sshd.go` (new: test SSH server)
- `arsy/servers/client_test.go` (expand)
- `arsy/servers/connection_test.go` (expand)

---

#### 🧪 MVP1-067: Add Compose parser and deployment worker tests
**Status**: Not Started  
**Complexity**: Medium  
**Estimasi**: 4-6 jam

**Tasks**:
- [ ] Test Compose YAML parsing (valid/invalid)
- [ ] Test MVP-1 subset validation (reject unsupported features)
- [ ] Test deployment worker job claiming (concurrency)
- [ ] Test worker state transitions
- [ ] Test worker failure recovery
- [ ] Test deployment log persistence (bounded buffer)
- [ ] Use fake/mock SSH executor
- [ ] Achieve >80% coverage for `arsy/deployments`

**Files**:
- `arsy/deployments/compose_test.go` (expand)
- `arsy/deployments/worker_test.go` (new)
- `arsy/deployments/testing/executor.go` (new: mock executor)

---

#### 🧪 MVP1-068: Add a Docker integration test profile
**Status**: Not Started  
**Complexity**: Medium-High  
**Estimasi**: 5-8 jam

**Tasks**:
- [ ] Create integration test suite requiring Docker
- [ ] Use `testcontainers-go` for SSH server container
- [ ] Test real SFTP upload to container
- [ ] Test real `docker compose` execution
- [ ] Test container inspection and logs
- [ ] Add build tag `-tags=integration`
- [ ] Document how to run integration tests
- [ ] Add to CI workflow (optional Docker job)

**Files**:
- `arsy/integration_test.go` (new)
- `.github/workflows/ci.yml` (update: add integration job)
- `README.md` (update: document integration tests)

---

### Phase E: Release Preparation (TODO - Low Priority)

#### 📦 MVP1-069: Add frontend build verification
**Status**: Not Started  
**Complexity**: Low  
**Estimasi**: 1-2 jam

**Tasks**:
- [ ] Add CI job for frontend build
- [ ] Run `npm ci` (clean install)
- [ ] Run `npm run build` for dashboard
- [ ] Verify embedded files in `arsy/console/dist/`
- [ ] Check build artifact size
- [ ] Add to PR checks

**Files**:
- `.github/workflows/ci.yml` (update)

---

#### 📦 MVP1-070: Add release build checks
**Status**: Not Started  
**Complexity**: Low  
**Estimasi**: 2-3 jam

**Tasks**:
- [ ] Build for Linux amd64
- [ ] Build for Linux arm64
- [ ] Build for macOS (if needed)
- [ ] Verify binary size (<100MB uncompressed)
- [ ] Test binary execution (smoke test)
- [ ] Verify version info embedded
- [ ] Update `.goreleaser.yaml` if needed

**Files**:
- `.github/workflows/release.yml` (new)
- `.goreleaser.yaml` (verify)

---

#### 📚 MVP1-071: Write MVP-1 installation and operation documentation
**Status**: Not Started  
**Complexity**: Low  
**Estimasi**: 3-4 jam

**Tasks**:
- [ ] Write `docs/installation.md`:
  - Prerequisites (Linux server, Docker, Compose)
  - Download binary
  - First run & superuser creation
  - Secrets key setup
- [ ] Write `docs/server-requirements.md`:
  - SSH key authentication setup
  - Docker Engine installation
  - Docker Compose plugin
  - Firewall/network considerations
- [ ] Write `docs/first-deployment.md`:
  - Add server walkthrough
  - Create project
  - Deploy sample app
  - View logs
  - Common issues
- [ ] Update main `README.md` with quickstart

**Files**:
- `docs/installation.md` (new)
- `docs/server-requirements.md` (new)
- `docs/first-deployment.md` (new)
- `README.md` (update)

---

#### ✅ MVP1-072: Run the complete MVP-1 acceptance checklist
**Status**: Not Started  
**Complexity**: Low  
**Estimasi**: 2-3 jam

**Tasks**:
- [ ] Run `go fmt ./...` and verify no changes
- [ ] Run `go vet ./...` and fix issues
- [ ] Run `go test ./...` and verify all pass
- [ ] Run `golangci-lint run` and fix critical issues
- [ ] Build binary and verify it starts
- [ ] Run through onboarding wizard manually
- [ ] Add a server, create project, deploy sample app
- [ ] Verify logs, stop, restart, rollback
- [ ] Document any issues found
- [ ] Create MVP-1 release tag

**Files**:
- `.plan/task-mvp-1.md` (update: mark all complete)
- `CHANGELOG.md` (add MVP-1 entry)

---

## 🚧 Open GitHub Issues (Beyond MVP-1)

### MVP-1 Extensions (Post-Release)

| # | Title | Complexity | Estimasi |
|---|-------|------------|----------|
| 149 | [MVP1-075] Prevent overlapping deployments | Medium | 4-6h |
| 143 | [MVP1-074] SSH credential rotation | Medium-High | 6-8h |
| 142 | [MVP1-073] Encrypt Compose manifests at rest | Medium | 4-6h |

---

### Application Engine (AENG) — Advanced Features

**Foundation (Done)**: Registry, Fields, Relations  
**Remaining** (17 issues):

| Category | Issues | Total Est. |
|----------|--------|-----------|
| Core Engine | AENG-029 to AENG-032 | 3-4 days |
| Add-on System | AENG-033, AENG-034 | 2-3 days |
| Security & Runtime | AENG-035, AENG-036 | 2 days |
| Documentation | AENG-020, AENG-021, AENG-022 | 1 day |
| Definitions | AENG-012 to AENG-020 | 2 days |

**Total**: ~10-13 days of work

**Details**:
- **AENG-029**: Unit of Work & transaction manager
- **AENG-030**: Static Go model persistence mapper
- **AENG-031**: Safe query compiler (PocketBase + dbx)
- **AENG-032**: Versioned schema migration executor
- **AENG-033**: Add-on runtime registry & lifecycle
- **AENG-034**: Idempotent module seed data runner
- **AENG-035**: Enforce model/record security metadata
- **AENG-036**: Validate view/action metadata
- **AENG-020 to AENG-022**: Documentation & glossary
- **AENG-012 to AENG-019**: Contracts & definitions

---

### BaaS Platform Features (BAAS) — Multi-Tenancy

**Status**: Not Started  
**Remaining** (4 issues):

| # | Title | Complexity | Estimasi |
|---|-------|------------|----------|
| 138 | [BAAS-021] Tenant admin & membership APIs | High | 2-3 days |
| 139 | [BAAS-022] Tenant database lifecycle | High | 2-3 days |
| 140 | [BAAS-023] Durable operations outbox | Medium-High | 1-2 days |
| 141 | [BAAS-024] Control-plane backup/restore | Medium | 1-2 days |

**Total**: ~6-10 days of work

**Purpose**: Multi-tenant BaaS architecture (future milestone)

---

## 📋 Prioritized Work Queues

### Queue 1: MVP-1 Completion (Critical Path)
**Goal**: Working end-to-end deployment system  
**Timeline**: 1-2 weeks

1. ✅ Merge PR #121, #122, #123 (when CI passes)
2. 🔥 MVP1-056: Deployment history & logs API (4-6h)
3. 🔥 MVP1-057: Stop & redeploy operations (4-6h)
4. 🔥 MVP1-058: Configuration rollback (6-8h)
5. 🎨 MVP1-059: Overview screen (4-6h)
6. 🎨 MVP1-060: Projects management UI (3-4h)
7. 🎨 MVP1-061: Application creation UI (5-6h)
8. 🎨 MVP1-062: Deploy action & progress (6-8h)
9. 🎨 MVP1-063: Application operations UI (5-6h)
10. 🎨 MVP1-064: Onboarding wizard (4-5h)
11. 🧪 MVP1-065 to MVP1-068: Tests (16-24h)
12. 📦 MVP1-069 to MVP1-072: Release prep (8-12h)

**Total**: ~65-90 hours (~2 weeks at 40h/week)

---

### Queue 2: Security & Stability (Post-MVP-1)
**Goal**: Production-ready hardening  
**Timeline**: 3-5 days

1. MVP1-073: Encrypt manifests at rest (4-6h)
2. MVP1-074: SSH credential rotation (6-8h)
3. MVP1-075: Prevent overlapping deployments (4-6h)
4. Add comprehensive error handling
5. Add metrics & observability
6. Security audit & penetration testing

---

### Queue 3: Application Engine (Future)
**Goal**: Metadata-driven app framework  
**Timeline**: 2-3 weeks

1. Complete AENG definitions (AENG-012 to AENG-020)
2. Core engine (AENG-029 to AENG-032)
3. Add-on system (AENG-033, AENG-034)
4. Security runtime (AENG-035, AENG-036)
5. Documentation (AENG-020 to AENG-022)

---

### Queue 4: BaaS Platform (Future)
**Goal**: Multi-tenant architecture  
**Timeline**: 2 weeks

1. BAAS-021: Tenant APIs
2. BAAS-022: Database lifecycle
3. BAAS-023: Operations outbox
4. BAAS-024: Backup/restore

---

## 🎯 Recommended Next Steps

### Option A: Validate Current State (2-3 hours)
```bash
# 1. Build dashboard
cd dashboard && npm install && npm run build

# 2. Build binary
cd .. && make build

# 3. Run tests
make test

# 4. Test run
.builds/arsy serve --dev
```

**Outcome**: Know what works and what's broken

---

### Option B: Complete MVP-1 Backend (2-3 days)
Focus on MVP1-056, 057, 058 (deployment operations)

**Outcome**: Full backend API ready for frontend

---

### Option C: Complete MVP-1 Frontend (1 week)
Focus on MVP1-059 to 064 (all UI screens)

**Outcome**: Complete user-facing application

---

### Option D: Full MVP-1 Sprint (1-2 weeks)
Complete all MVP-1 tasks (backend + frontend + tests + docs)

**Outcome**: Production-ready v1.0.0 release

---

## 📊 Progress Tracking

### MVP-1 Progress
- **Foundation**: ████████░░ 80% (41/51 tasks)
- **Backend API**: ██████░░░░ 60% (in PRs)
- **Frontend UI**: ░░░░░░░░░░ 0% (not started)
- **Testing**: ██░░░░░░░░ 20% (some tests exist)
- **Documentation**: ░░░░░░░░░░ 0% (not started)

**Overall MVP-1**: ████░░░░░░ 40%

---

## 🛠️ Development Commands

```bash
# Build dashboard
make dashboard

# Build binary
make build

# Run tests
make test

# Run with coverage
make test-report

# Lint
make lint

# Run dev server (with hot reload)
cd dashboard && npm run dev
.builds/arsy serve --dev

# Check PR status
gh pr list
gh pr view 121

# Check issues
gh issue list
gh issue view 149
```

---

## 📝 Notes

- All MVP-1 tasks should follow security invariants in `.plan/task-mvp-1.md`
- Never store secrets in plaintext
- Always use typed interfaces, never raw shell
- Test coverage target: >80% for core packages
- UI should work without JavaScript (progressive enhancement)
- All API responses must be non-sensitive (use DTOs)

---

**Last Updated**: 2026-11-07 13:53 WIB  
**Next Review**: After PR #121, #122, #123 merge
