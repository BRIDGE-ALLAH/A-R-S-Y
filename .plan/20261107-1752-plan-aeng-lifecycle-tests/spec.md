 # Spec — AENG lifecycle test batch

 ## FR
 - `REQ-102-1`: Provide isolated tests covering temporary data directory startup.
 - `REQ-102-2`: Cover migration registration/execution where practical using existing migrations/test app APIs.
 - `REQ-102-3`: Cover clean app/database shutdown and cleanup on success/failure paths.
 - `REQ-127-R`: Reconcile issue #127 acceptance criteria against existing `arsy/engine/query/` implementation/tests; produce evidence.

 ## NFR
 - Deterministic, no network, temp dirs only.
 - Compatible with `go test ./...` expectations for affected packages.

 ## Acceptance Criteria
 - New/updated tests pass in affected scope.
 - No production behavior changes for #102 unless justified and documented.
 - #127 either evidenced complete (close path) or gap documented.

