package uow

import "errors"

// FakeTx is a deterministic in-memory transaction used in tests. It records
// terminal state and fails repeated terminal calls with ErrTxCompleted.
type FakeTx struct {
	Committed  bool
	RolledBack bool
	// CommitErr, when set, is returned by Commit.
	CommitErr error
}

func (t *FakeTx) Commit() error {
	if t.Committed || t.RolledBack {
		return ErrTxCompleted
	}
	if t.CommitErr != nil {
		return t.CommitErr
	}
	t.Committed = true
	return nil
}

func (t *FakeTx) Rollback() error {
	if t.Committed || t.RolledBack {
		return ErrTxCompleted
	}
	t.RolledBack = true
	return nil
}

// FakeTransactor hands out FakeTx instances in order.
type FakeTransactor struct {
	// BeginErr, when set, is returned by Begin.
	BeginErr error
}

func (f *FakeTransactor) Begin() (Tx, error) {
	if f.BeginErr != nil {
		return nil, f.BeginErr
	}
	return &FakeTx{}, nil
}

// ErrBegin is a deterministic begin failure for tests.
var ErrBegin = errors.New("uow: fake begin failure")
