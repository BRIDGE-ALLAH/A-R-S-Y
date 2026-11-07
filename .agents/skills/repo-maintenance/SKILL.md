---
name: repo-maintenance
description: Use when auditing a repository backlog, reconciling GitHub Issues with code and plans, cleaning obsolete files/dependencies, or simplifying project structure safely.
---

# Repository Maintenance

## Scope
Own repository discovery, backlog reconciliation and safe cleanup. Do not own detailed bug debugging, PR integration, release publication or Wiki page editing; route those to their dedicated skills when needed.

## Workflow

1. **Baseline:** inspect `git status`, branch, remotes, default branch, package/workspace manifests, CI workflows and applicable instructions. Preserve unrelated changes.
2. **Inventory:** inspect GitHub Open/Closed Issues, Open/Merged/Closed PRs, `.plan` as either file or directory, `AGENTS.md`, root Markdown, `docs/`, `.github/` and relevant source/config directories. Use targeted `rg`/file listing; avoid dependency, generated, vendor and build trees.
3. **Classify:** record only actionable, evidence-backed findings: bug, security, feature, enhancement, technical debt, performance or documentation. Confirm each against current code and PR history.
4. **Deduplicate:** search existing Issues and PRs before creating anything. Reuse a matching Issue; do not create one Issue per file in a related cleanup batch.
5. **Issue first:** create a GitHub Issue before implementing each new independent actionable local finding. Include problem, evidence/path, bounded scope, acceptance criteria, priority/risk and verification. If GitHub is unavailable, mark blocked; do not silently implement unrelated backlog work.
6. **Plan:** use native Codex Plan/Tasks for substantial work. Tasks should be independently verifiable and linked to Issues. Simple changes need no elaborate plan. Keep `.plan/` only when durable cross-session progress is useful.
7. **Cleanup:** classify candidates as KEEP, UPDATE, MERGE, DELETE or REVIEW. Delete only when usage checks prove they are obsolete. Check static/dynamic imports, reflection, runtime discovery, entry points, scripts, CI, packaging, deployment, migrations, docs and external integrations. Check dependency usage before removing packages and regenerate lockfiles with the existing package manager.
8. **Preserve:** never casually remove license/legal notices, security policy, migrations, tests/fixtures/snapshots, lockfiles, required configs/workflows, runtime assets, backups or user-created changes. Retain uncertain files and explain why.
9. **Minimal docs:** prefer concise README + AGENTS.md + CHANGELOG; add technical docs only when useful. Do not create empty docs, duplicate roadmaps, TODO backlogs or redundant indexes. Update affected docs only.
10. **Verify:** run targeted import/reference checks, relevant tests, lint/typecheck/build and documentation-link checks where available. Inspect `git diff --check` and the complete diff. Confirm deleted files are not required.
11. **Deliver:** commit cleanup in logical batches, push to intended origin and follow PR/merge policy. Do not close Issues until verified on the default branch.

## Anti-slop
- No speculative cleanup based on filenames or age.
- No mass deletion without reference analysis.
- No cosmetic refactor outside scope.
- No duplicate Issues, tasks or docs.
- No claiming a file/dependency is unused without evidence.

## Output
Summarize findings, Issues, files/dependencies removed, files retained as uncertain, validation results, commits/PRs and blockers.
