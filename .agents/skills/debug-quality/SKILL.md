---
name: debug-quality
description: Use when reproducing bugs, fixing failing tests/builds/CI, reviewing quality or security risks, or validating a code change before completion.
---

# Debugging & Quality

## Root-cause workflow

1. Read the Issue/PR, relevant instructions, recent diff and full relevant error context.
2. Reproduce the bug or failure when possible. Capture the baseline and exact command/output.
3. Locate the failing layer; trace data/control flow to the source. Compare with a working path.
4. State one testable root-cause hypothesis. Change the smallest relevant variable; do not stack guesses.
5. For behavior bugs, add a regression test that fails for the original defect and passes after the fix. Use a red/green check when practical; do not write tests that merely grep source text or mirror implementation constants.
6. Implement the smallest complete fix. Preserve public contracts unless the Issue requires a change.
7. Run the narrowest meaningful test first, then related integration/regression tests and checks required by project policy.
8. Inspect output and exit status. Diagnose failures before retrying. Do not rerun an unchanged failing command without new evidence or a changed condition.
9. At milestone/merge boundary, run broader required checks: tests, lint, typecheck, build, security checks where relevant and CI.
10. Review diff, error handling, edge cases, compatibility, secrets and regression risk.

## CI triage

For each failed check, record workflow/run, job/step, commit SHA, first actionable error and classification: change regression, pre-existing failure, flaky test, environment, external service or unknown. Fix in-scope root causes. Rerun and verify against the exact candidate commit. Do not disable, skip or weaken a required check.

## Risk-based review

- Auth/authz: check privilege boundaries, token/session behavior and negative paths.
- API: check schema, validation, error semantics and backward compatibility.
- Database: validate migration forward/rollback in isolated DB; never run destructive production migrations without approval.
- Concurrency: inspect race conditions, locking, retries and idempotency.
- UI: test changed user flow, keyboard/accessibility basics and responsive behavior where relevant.
- Dependencies: inspect advisories, compatibility and actual usage before upgrades/removal.

## Anti-loop

After three materially different attempts fail on the same root cause, stop that path. Reassess the hypothesis/architecture, summarize evidence and blocker, and continue unrelated safe tasks. Do not claim a fix without fresh verification.

## Output
Root cause; reproduction; changed files; regression test; commands and results; CI commit SHA/status; residual risks; blockers.
