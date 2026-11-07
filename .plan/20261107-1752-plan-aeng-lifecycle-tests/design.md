 # Design — AENG lifecycle test batch

 ## Context
 - `main` includes merged query compiler (#132/#160). Query AST/validation exists in `arsy/engine/query/` and tests pass; issue #127 remains OPEN (possible stale tracking).
 - `tests.NewTestApp*`, `TempDirClone`, `Cleanup`, and bootstrap hooks provide isolated PocketBase test apps.

 ## Decisions
 - Implement #102 as targeted regression/lifecycle tests using existing harness; prefer `tests` package or nearest conventional location to avoid production edits.
 - Treat #131 as blocked until #101/#102 evidence is satisfied; do not start mapper implementation now.
 - Use evidence-first reconciliation for #127 before writing code.

 ## Risks
 - Migration execution tests can be brittle if they depend on full app migration state; keep assertions minimal/practical.
 - Parallel test interference mitigated by temp cloned data dirs per app.

