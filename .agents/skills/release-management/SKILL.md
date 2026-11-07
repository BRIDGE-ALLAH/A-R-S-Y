---
name: release-management
description: Use when preparing a version bump, changelog, Git tag, GitHub Release, first release, release candidate, or investigating release/tag/permission failures.
---

# Release Management

## Guardrails

A GitHub Release is not permission to deploy to production or publish to an external package registry. Follow existing repository release policy and automation. Do not trigger production deployment, publish packages, rotate secrets or alter infrastructure without explicit authorization. Never rewrite an existing published tag automatically.

## Readiness and identity

1. Inspect canonical `owner/repo`, intended origin, `fork`, `parent`, `source`, default branch and write permissions. A fork flag alone is not a release blocker; verify permissions and target. Never push to upstream without authorization or detach a fork automatically.
2. Inspect package manifests, monorepo versioning, lockfiles, all local/remote tags, GitHub Releases, workflows and previous changelog format.
3. Fetch tags from intended origin. No existing tags means first-release initialization, not failure.
4. Confirm clean/reviewed candidate state and that all intended changes/PRs are merged into the release branch.

## Compare-range evidence

- If a previous release tag exists, establish `PREVIOUS_TAG..CANDIDATE_SHA`; verify the previous tag is an ancestor when applicable.
- For first release, use an explicitly recorded initial history range or all commits up to the candidate; do not pretend a previous tag exists.
- Inventory commits, merged PRs and linked Issues in the range. Include all meaningful user-facing changes; mark internal-only exclusions deliberately.
- Verify candidate SHA belongs to intended default/release branch.

## Version and changelog

Use existing project convention; otherwise SemVer:
- PATCH: backward-compatible fixes.
- MINOR: backward-compatible capabilities.
- MAJOR: breaking changes.
- Prerelease: alpha/beta/rc when project maturity requires it.

In monorepos, bump only affected packages. Update manifests/lockfiles consistently. Do not invent or bump versions without meaningful release content.

Maintain professional Keep a Changelog-style sections as applicable: Unreleased, Added, Changed, Deprecated, Removed, Fixed, Security. Include version/date, user-facing changes, Issue/PR references, breaking changes and migration notes. Derive notes from compare-range evidence, not memory. Never list unmerged features as shipped.

## Pre-release gates

- Required tests/build/lint/typecheck/security and CI pass for candidate SHA.
- Version is consistent across relevant manifests/artifacts.
- Changelog covers the compare range.
- Release notes are reviewed and accurate.
- Tag does not already exist or point elsewhere.
- Required assets and checksums are ready when applicable.
- Any tag-triggered workflow is understood; do not trigger unauthorized production deployment.

## Publish GitHub Release

1. Follow repo policy on whether version/changelog commit must be on default branch before tagging.
2. Create tag at exact verified release commit using existing naming convention (often `vX.Y.Z`).
3. Push tag to intended origin.
4. Create GitHub Release with accurate, formatted notes and relevant assets.
5. Verify release URL, tag, target commit SHA, draft/prerelease state, assets and workflow results.
6. If registry publishing is part of automation, verify publication only when explicitly authorized and permitted.

## Failure handling

Inspect exact CLI/API/workflow error. Classify authentication, write permission, branch/tag protection, workflow permissions, missing tool, fork-network restriction or external failure. Fix only authorized in-scope causes. Do not disable protection or fabricate a tag/release. If blocked, report exact permission or setting required.

## Rollback/recovery

Before high-risk release side effects, identify supported correction path. Prefer correcting release notes/metadata over moving a published tag. Do not delete published tags/releases or unpublish artifacts without explicit authorization and a documented impact assessment.

## Output
Readiness GO/NO-GO; canonical repo and target; previous tag or FIRST RELEASE; compare range and candidate SHA; version/bump; merged PR/Issue inventory; changelog coverage; test/CI evidence; tag; release URL/assets; workflow status; blockers. Never claim a release succeeded until verified remotely.
