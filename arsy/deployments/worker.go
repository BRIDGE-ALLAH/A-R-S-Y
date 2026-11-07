package deployments

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bridge-allah/a-r-s-y/core"
)

type DeploymentProcessor func(context.Context, *core.Record) error

type Worker struct {
	App           core.App
	Process       DeploymentProcessor
	PollInterval  time.Duration
	RecoveryAfter time.Duration
}

// ProcessNext claims and processes at most one queued deployment.
func (w Worker) ProcessNext(ctx context.Context) (bool, error) {
	if ctx == nil {
		return false, errors.New("worker context is required")
	}
	if w.App == nil || w.Process == nil {
		return false, errors.New("worker app and processor are required")
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	record, err := ClaimNextQueuedDeployment(w.App)
	if errors.Is(err, ErrNoQueuedDeployments) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	processErr := w.Process(ctx, record)
	if processErr != nil {
		next := StateFailed
		summary := "Deployment failed. Review the saved deployment diagnostics before retrying."
		if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			next = StateCancelled
			summary = "Deployment was cancelled before completion."
		}
		if _, updateErr := UpdateDeploymentState(w.App, record.Id, next, summary); updateErr != nil {
			return true, fmt.Errorf("deployment failed and its terminal state could not be saved: %w", updateErr)
		}
		return true, nil
	}

	latest, err := w.App.FindRecordById("arsy_deployments", record.Id)
	if err != nil {
		return true, fmt.Errorf("reload processed deployment: %w", err)
	}
	if State(latest.GetString("status")) != StateVerifying {
		_, updateErr := UpdateDeploymentState(w.App, record.Id, StateFailed, "Deployment processor returned before verification completed.")
		if updateErr != nil {
			return true, updateErr
		}
		return true, nil
	}
	if _, err := UpdateDeploymentState(w.App, record.Id, StateSucceeded, "Deployment completed and verification passed."); err != nil {
		return true, err
	}
	return true, nil
}

// Run recovers stale jobs once on startup, then polls and processes queued work
// until the supplied context is cancelled.
func (w Worker) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("worker context is required")
	}
	if w.App == nil || w.Process == nil {
		return errors.New("worker app and processor are required")
	}
	recoveryAfter := w.RecoveryAfter
	if recoveryAfter <= 0 {
		recoveryAfter = 5 * time.Minute
	}
	if _, err := RecoverInterruptedDeployments(w.App, time.Now().Add(-recoveryAfter)); err != nil {
		return fmt.Errorf("recover interrupted deployments: %w", err)
	}
	interval := w.PollInterval
	if interval <= 0 {
		interval = 2 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		processed, err := w.ProcessNext(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			return err
		}
		if processed {
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
