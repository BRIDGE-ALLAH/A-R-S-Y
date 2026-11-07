---
name: docs-wiki
description: Use when auditing, cleaning, creating or updating README/docs/Markdown, GitHub Wiki pages, navigation, links or user/developer documentation.
---

# Documentation & GitHub Wiki

## Ownership

- README: project purpose, requirements, install and quick start.
- AGENTS.md: concise permanent repository instructions.
- `docs/`: essential technical documentation versioned with code and reviewed through PRs.
- GitHub Wiki: extended user/developer guides, tutorials, workflows, FAQ and troubleshooting.
- CHANGELOG: released changes only.
- GitHub Issues: work tracking, never a documentation backlog.

Use one authoritative location per topic; link between locations instead of copying entire pages.

## Audit

1. Inspect existing Markdown, `.plan`, README, docs, GitHub Wiki availability, pages, `_Sidebar.md`, assets, history and references.
2. Compare instructions with actual merged code and current configuration.
3. Classify KEEP, UPDATE, MERGE, DELETE or REVIEW.
4. Check links, commands, paths, package names, screenshots and feature status.
5. Do not treat missing documentation as a bug if no user/developer need exists. For new independent actionable doc work, create/link a GitHub Issue first.

## Minimal documentation

Prefer a concise README, AGENTS and CHANGELOG plus only necessary technical docs. Do not create empty templates, duplicate roadmap/TODO/index files or docs purely to satisfy a checklist. Remove obsolete plans only after checking references and active work.

## Update rules

- Update only pages affected by actual behavior/API/config/development/deployment changes.
- Never describe unreleased functionality as available.
- Include prerequisites, verified commands and expected results where useful.
- Keep examples runnable and version references accurate.
- Remove stale instructions, duplicate sections and broken links.

## Wiki workflow

1. Check whether Wiki is enabled and whether the authenticated identity can read/write it.
2. If disabled, do not enable it or change visibility/permissions without authorization. Report the limitation.
3. Treat Wiki as a separate Git repository. Inspect its actual remote/default branch/history before editing.
4. Preserve existing work. Update/create only justified pages; maintain simple navigation and `_Sidebar.md` when multiple pages warrant it.
5. Review Wiki diff, links and assets; commit and push separately to the Wiki remote.
6. Verify published pages. Never assume a code-repo push updates Wiki.
7. Do not force-push or overwrite concurrent Wiki changes.

## Validation

Run available Markdown lint/link checks. Verify relative paths and all changed commands against repository scripts/configuration where practical. Inspect diffs for duplicated or contradictory content.

## Output
List docs/Wiki retained, changed, merged, removed; reason for removal; link validation; Wiki availability; Wiki commit/push/publication status; blockers.
