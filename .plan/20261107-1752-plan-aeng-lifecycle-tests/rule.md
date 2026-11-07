 # Rules — AENG lifecycle test batch

 1. Preserve unrelated/uncommitted user work (`README.md`, `.agents/`, `AGENTS.md`, `GOAL.txt`).
 2. `#102` is **test-only** unless strictly required for testability; no production behavior changes without explicit evidence.
 3. Tests must use temporary directories and must not touch developer data or require network access.
 4. Do not redo verified/merged work (`#132` compiler, existing query validation/AST code). Reconcile stale issues with evidence instead.
 5. Do not bypass branch protection, required checks, or GitHub policy.
 6. Keep changes minimal and scoped; no unrelated refactors or dependency additions.
 7. Record deviations and verification evidence in `report.md`.

