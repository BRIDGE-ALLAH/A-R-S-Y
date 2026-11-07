# Repository Engineering Rules

## Mission
Keep this repository correct, secure, maintainable, minimal and releasable. Prefer the smallest complete change that satisfies the requirement.

## First actions
- Read this file and any applicable nested `AGENTS.md`/`AGENTS.override.md`.
- Inspect `git status`, current branch, remotes and recent history before editing.
- Discover project commands from manifests, scripts and CI; never assume a framework or command.
- Inspect relevant Issues, PRs and local plans before starting overlapping work.

## Ownership
- Code is implementation truth.
- GitHub Issues are the official backlog.
- PRs record review and integration.
- Codex Plan/Tasks track the active session.
- `.plan/` is optional persistent state, not a second backlog.
- README explains the project; `docs/` holds necessary versioned technical docs; Wiki holds useful extended guides; CHANGELOG records released changes.

## Engineering
- Follow existing architecture, style, dependency and branch conventions.
- Fix root causes. Keep diffs focused and backward-compatible unless change is explicitly intended.
- Add/update meaningful tests for behavior changes and bugs. Test according to impact; do not run unrelated suites repeatedly.
- Never weaken tests, lint, security checks or CI to get green.
- Review `git diff --check`, changed files and secret exposure before commit.
- Preserve unrelated user modifications, untracked files and local data.

## Tracking
- Before implementing a new independent actionable finding discovered locally, search existing Open/Closed Issues and PRs; create or link the appropriate Issue first.
- Reuse existing PRs/tasks when they cover the same work. Verify merged changes on the target branch before treating them as done.
- Close Issues only after acceptance criteria and required verification are satisfied on the intended default branch.

## Cleanup & Documentation
- Delete files only after proving they are unused by static/dynamic imports, runtime discovery, build, CI, packaging, deployment, migrations or external integrations.
- Preserve licenses, notices, security policies, migrations, fixtures, snapshots, lockfiles and required configuration.
- Keep Markdown concise. Do not create empty templates, duplicate indexes, competing TODO lists or unnecessary plans.
- Update only documentation affected by actual behavior changes. Keep Wiki and repository docs complementary, not copied wholesale.

## Safety
- No destructive production data/schema/storage operation, production deployment, production secret/DNS/firewall/server change, irreversible external publication, security bypass or protected-branch force push without explicit approval.
- Use isolated environments and rollback plans for high-risk changes.
- Never bypass branch protection, required reviews or required CI.
- A missing Git tag is not a release blocker. A fork flag alone is not a release prohibition; verify canonical repository and permissions.

## Verification & reporting
- Evidence before completion claims: run the relevant command, inspect output and exit status.
- Distinguish PASS, FAIL, NOT RUN and BLOCKED.
- Report actual commits, PR/merge state, CI run/commit, release tag and URL. Never fabricate results.
- For repeatable workflows, load the matching skill under `.agents/skills/` instead of duplicating its full procedure here.
