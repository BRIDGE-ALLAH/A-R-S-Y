// Package uow defines the driver-neutral Unit of Work contract for the
// application engine (issue #110) and a small manager implementing it
// (issue #130).
//
// Contract:
//   - A unit of work wraps multiple model operations in one transaction.
//   - Begin/commit/rollback are driver-neutral; drivers adapt to SQL.
//   - Returned errors and panics trigger rollback.
//   - Context cancellation triggers rollback; no transaction resources leak.
//   - Nested transactions are rejected with ErrNested.
//   - Commit is attempted exactly once per successful run; rollback exactly
//     once per failed run.
package uow

import "errors"

var (
	// ErrNested is returned when Run is called with a context that already
	// carries an active transaction.
	ErrNested = errors.New("uow: nested transactions are not supported")

	// ErrNoTransactor is returned when the manager was built without a
	// transactor.
	ErrNoTransactor = errors.New("uow: no transactor configured")

	// ErrTxCompleted is returned when a terminal operation (Commit or
	// Rollback) is invoked twice on the same transaction.
	ErrTxCompleted = errors.New("uow: transaction already completed")
)

// Tx is one active transaction handle. Exactly one of Commit or Rollback
// must terminate the transaction; the second terminal call must return
// ErrTxCompleted.
type Tx interface {
	Commit() error
	Rollback() error
}

// Transactor begins driver transactions.
type Transactor interface {
	Begin() (Tx, error)
}

// Manager runs work inside transactions provided by a Transactor.
type Manager struct {
	transactor Transactor
}

// NewManager builds a Manager. A nil transactor yields a manager whose Run
// fails with ErrNoTransactor.
func NewManager(t Transactor) *Manager {
	return &Manager{transactor: t}
}
