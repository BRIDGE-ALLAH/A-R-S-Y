---
name: pr-delivery
description: Use when triaging existing Pull Requests, preparing or updating a PR, addressing review feedback, merging changes, or reconciling Issues with merged work.
---

# Pull Request Delivery

## Before work

1. Identify canonical `owner/repo`, intended `origin`, default branch and current branch. Verify GitHub repository metadata when needed; do not infer fork status from remote names or missing tags.
2. Inspect Open, Merged and Closed PRs relevant to the Issue/branch. Read descriptions, linked Issues, changed files, commits, reviews, comments, checks and merge commit.
3. Inspect local Git status and preserve unrelated modifications.
4. Search for an existing PR before creating one.

## PR state handling

- **Open:** continue/update the existing branch/PR when it is the same work; inspect requested changes and conflicts.
- **Merged:** verify merge commit and target branch contains it. Treat it as integrated only after acceptance criteria are checked.
- **Closed, not merged:** inspect why it closed and what code landed; reopen/reuse only when appropriate, otherwise create a new Issue/PR without duplicating work.
- **Merged to non-default branch:** verify whether integration to the intended default branch remains required.

## Prepare/update

- Keep scope tied to Issue and acceptance criteria.
- Include concise summary, implementation details, test evidence, risk/migration notes and linked Issues.
- Follow repository PR template, branch naming and commit conventions.
- Run focused validation; review complete diff and `git diff --check`.
- Ensure no secrets, generated junk or unrelated files are included.

## CI and reviews

- Inspect checks on the exact head SHA.
- For failures, invoke `debug-quality` for root-cause triage.
- Address actionable review comments; explain any rejected suggestion with technical evidence.
- Re-run relevant tests after review changes.
- Do not bypass required status checks, approvals, merge queue or branch protection.

## Integration

- Push only to intended origin and authorized branch.
- Reuse existing PR; create a new PR only when needed.
- Merge only when required checks/reviews pass and repository policy permits.
- Never force-push protected branches or alter repository/fork settings automatically.
- Verify merge commit SHA and that the intended default branch contains it.
- Close linked Issues only when their acceptance criteria are met on that branch.
- Update task/plan state and relevant docs after verification.

## Output
List PR number/URL, state, head/base, head SHA, check result, review status, merge SHA, linked Issue state and any blocker. Never treat PR creation as merge or merge as release.
