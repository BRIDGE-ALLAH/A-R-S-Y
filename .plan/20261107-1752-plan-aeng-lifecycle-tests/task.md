 # Task — AENG lifecycle test batch (coding focus)

 Status: `IN_PROGRESS`
 Scope: `plan`
 Issues: `#102` (primary), reconcile `#127`, unblock path to `#131`

 ## Checklist
 - [x] Sync local `main` with `origin/main` and verify #132/#160 merged (`de2fab7`).
 - [x] Reconcile open AENG issues/PRs: #124/#125/#128/#130 closed; query AST+validation+compiler code exists in `arsy/engine/query/`; #127 OPEN but implementation appears present (needs evidence/issue update decision).
 - [ ] [S] `#102` Inspect existing test harness (`tests.NewTestApp`, `TempDirClone`, `Cleanup`, `Bootstrap`, `RunAllMigrations`) and identify minimal test locations/patterns. Acceptance: note exact APIs/files used; no production changes.
 - [ ] [M] `#102` Add deterministic, network-free isolated tests for temporary data-dir startup, migration registration/execution where practical, and clean shutdown/cleanup. Files (expected): new `*_test.go` under `tests/` or `core/` following repo conventions (choose smallest fit). Acceptance: uses temp dirs, no developer data touched, cleanup on success/failure, `go test ./...`-relevant scope passes.
 - [ ] [S] Run targeted tests (`go test` for touched packages), then broader affected scope if low risk. Record evidence.
 - [ ] [S] `#127` Evidence check: map acceptance criteria in issue #127 to `arsy/engine/query/query.go` + tests; if fully satisfied and merged, comment/close with evidence (no code redo). If gap exists, create follow-up task.
 - [ ] [M] `#131` (only after deps evidence) Prepare static Go model persistence mapper design slice in code: explicit adapter boundary, typed CRUD against isolated PocketBase/dbx test app, normalized errors. Depends on #101/#102 evidence; do not start if blocked.
 - [ ] [S] Commit/push PR for #102 with issue linkage; wait for CI; merge under policy; close #102 when verified on default branch.
 - [ ] [S] Update `report.md` with evidence, deviations, remaining risks.

