package uow

import (
	"context"
	"errors"
	"testing"
)

func TestRunCommitOnce(t *testing.T) {
	m := NewManager(&FakeTransactor{})
	var gotTx Tx
	err := m.Run(context.Background(), func(ctx context.Context) error {
		gotTx, _ = TxFromContext(ctx)
		return nil
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	fake, ok := gotTx.(*FakeTx)
	if !ok {
		t.Fatalf("expected *FakeTx in context")
	}
	if !fake.Committed || fake.RolledBack {
		t.Fatalf("expected committed exactly once, got %+v", fake)
	}
}

func TestRunRollbackOnReturnedError(t *testing.T) {
	m := NewManager(&FakeTransactor{})
	wantErr := errors.New("work failed")
	var gotTx Tx
	err := m.Run(context.Background(), func(ctx context.Context) error {
		gotTx, _ = TxFromContext(ctx)
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected work error, got %v", err)
	}
	if !gotTx.(*FakeTx).RolledBack {
		t.Fatal("expected rollback")
	}
}

func TestRunRollbackOnPanic(t *testing.T) {
	m := NewManager(&FakeTransactor{})
	var gotTx Tx
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic to propagate")
			}
		}()
		_ = m.Run(context.Background(), func(ctx context.Context) error {
			gotTx, _ = TxFromContext(ctx)
			panic("boom")
		})
	}()
	if !gotTx.(*FakeTx).RolledBack {
		t.Fatal("expected rollback on panic")
	}
}

func TestRunRollbackOnContextCancellation(t *testing.T) {
	m := NewManager(&FakeTransactor{})
	ctx, cancel := context.WithCancel(context.Background())
	var gotTx Tx
	err := m.Run(ctx, func(ctx context.Context) error {
		gotTx, _ = TxFromContext(ctx)
		cancel()
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if !gotTx.(*FakeTx).RolledBack {
		t.Fatal("expected rollback on cancellation")
	}
}

func TestRunNestedRejected(t *testing.T) {
	m := NewManager(&FakeTransactor{})
	err := m.Run(context.Background(), func(ctx context.Context) error {
		return m.Run(ctx, func(context.Context) error { return nil })
	})
	if !errors.Is(err, ErrNested) {
		t.Fatalf("expected ErrNested, got %v", err)
	}
}

func TestRunBeginFailure(t *testing.T) {
	m := NewManager(&FakeTransactor{BeginErr: ErrBegin})
	err := m.Run(context.Background(), func(context.Context) error { return nil })
	if !errors.Is(err, ErrBegin) {
		t.Fatalf("expected ErrBegin, got %v", err)
	}
}

func TestRunCommitFailure(t *testing.T) {
	m := NewManager(&FakeTransactor{})
	commitErr := errors.New("commit failed")
	m.transactor = &commitErrTransactor{err: commitErr}
	err := m.Run(context.Background(), func(context.Context) error { return nil })
	if !errors.Is(err, commitErr) {
		t.Fatalf("expected commit error, got %v", err)
	}
}

func TestRunNoTransactor(t *testing.T) {
	m := NewManager(nil)
	err := m.Run(context.Background(), func(context.Context) error { return nil })
	if !errors.Is(err, ErrNoTransactor) {
		t.Fatalf("expected ErrNoTransactor, got %v", err)
	}
}

func TestFakeTxRepeatedCompletion(t *testing.T) {
	tx := &FakeTx{}
	if err := tx.Commit(); err != nil {
		t.Fatalf("first commit: %v", err)
	}
	if err := tx.Commit(); !errors.Is(err, ErrTxCompleted) {
		t.Fatalf("expected ErrTxCompleted on second commit, got %v", err)
	}
	if err := tx.Rollback(); !errors.Is(err, ErrTxCompleted) {
		t.Fatalf("expected ErrTxCompleted on rollback after commit, got %v", err)
	}
}

type commitErrTransactor struct{ err error }

func (c *commitErrTransactor) Begin() (Tx, error) {
	return &FakeTx{CommitErr: c.err}, nil
}
