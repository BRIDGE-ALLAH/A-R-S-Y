// Package hooks defines the transaction-aware model lifecycle hook contract
// (issue #109) and a deterministic registry implementing it (issue #129).
//
// Contract:
//   - Hooks run for create, update, and delete operations.
//   - Three phases: before (may veto persistence), after (in-transaction),
//     and after-commit (side effects only after successful commit).
//   - Hook order is registration order and is deterministic.
//   - Duplicate hook IDs for the same operation and phase are rejected.
//   - A failing before-hook prevents persistence from proceeding.
//   - After-commit hooks never run before a successful commit; they are
//     queued during the operation and flushed by the caller afterwards.
//   - Hook errors cannot bypass validation or authorization: callers must
//     run those layers before triggering persistence hooks.
//   - Re-entrant triggering from inside a hook is rejected.
package hooks

import "errors"

var (
	// ErrDuplicateHook is returned when a hook ID is registered twice for
	// the same operation and phase.
	ErrDuplicateHook = errors.New("hooks: duplicate hook registration")

	// ErrReentrant is returned when a hook triggers another hook phase
	// while a hook is already running.
	ErrReentrant = errors.New("hooks: re-entrant hook triggering")

	// ErrUnknownPhase is returned for an invalid phase value.
	ErrUnknownPhase = errors.New("hooks: unknown phase")
)
