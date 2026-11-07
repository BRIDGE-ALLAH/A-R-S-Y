package uow

import (
	"context"
	"errors"
	"fmt"
)

type ctxKey struct{}

// Run executes fn inside one transaction. On success the transaction is
// committed exactly once. A returned error, a panic, or context cancellation
// triggers exactly one rollback.
func (m *Manager) Run(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if m == nil || m.transactor == nil {
		return ErrNoTransactor
	}
	if ctx.Value(ctxKey{}) != nil {
		return ErrNested
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	tx, err := m.transactor.Begin()
	if err != nil {
		return fmt.Errorf("uow: begin: %w", err)
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			_ = tx.Rollback()
			panic(recovered)
		}
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, ErrTxCompleted) {
				err = errors.Join(err, fmt.Errorf("uow: rollback: %w", rbErr))
			}
		}
	}()

	txCtx := context.WithValue(ctx, ctxKey{}, tx)
	if err = fn(txCtx); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("uow: commit: %w", err)
	}
	return nil
}

// TxFromContext returns the transaction carried by ctx, if any.
func TxFromContext(ctx context.Context) (Tx, bool) {
	tx, ok := ctx.Value(ctxKey{}).(Tx)
	return tx, ok
}
