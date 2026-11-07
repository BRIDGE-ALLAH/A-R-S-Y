package deployments

import (
	"errors"
	"fmt"
	"time"

	"github.com/bridge-allah/a-r-s-y/core"
)

var ErrNoQueuedDeployments = errors.New("no queued deployments are available")

var interruptedDeploymentStates = []State{
	StatePreparing,
	StateUploading,
	StateValidating,
	StateDeploying,
	StateVerifying,
}

// ClaimNextQueuedDeployment atomically claims the oldest queued deployment
// inside a PocketBase transaction so concurrent workers cannot claim it twice.
func ClaimNextQueuedDeployment(app core.App) (*core.Record, error) {
	if app == nil {
		return nil, errors.New("PocketBase app is required")
	}
	var claimed *core.Record
	err := app.RunInTransaction(func(txApp core.App) error {
		records, err := txApp.FindRecordsByFilter("arsy_deployments", `status = "queued"`, "created", 1, 0)
		if err != nil {
			return fmt.Errorf("find queued deployments: %w", err)
		}
		if len(records) == 0 {
			return ErrNoQueuedDeployments
		}
		record := records[0]
		next, err := Transition(State(record.GetString("status")), StatePreparing)
		if err != nil {
			return err
		}
		record.Set("status", string(next))
		record.Set("started_at", time.Now().UTC().Format(time.RFC3339Nano))
		if err := txApp.Save(record); err != nil {
			return fmt.Errorf("claim deployment: %w", err)
		}
		claimed = record
		return nil
	})
	if err != nil {
		return nil, err
	}
	return claimed, nil
}

// UpdateDeploymentState applies the state machine and persists lifecycle
// timestamps and a bounded operator-facing summary.
func UpdateDeploymentState(app core.App, recordID string, next State, summary string) (*core.Record, error) {
	if app == nil {
		return nil, errors.New("PocketBase app is required")
	}
	record, err := app.FindRecordById("arsy_deployments", recordID)
	if err != nil {
		return nil, fmt.Errorf("find deployment: %w", err)
	}
	current := State(record.GetString("status"))
	transitioned, err := Transition(current, next)
	if err != nil {
		return nil, err
	}
	record.Set("status", string(transitioned))
	if summary != "" {
		if len(summary) > 2000 {
			summary = summary[:2000]
		}
		record.Set("summary", summary)
	}
	if next == StatePreparing && record.GetString("started_at") == "" {
		record.Set("started_at", time.Now().UTC().Format(time.RFC3339Nano))
	}
	if next == StateSucceeded || next == StateFailed || next == StateCancelled {
		record.Set("finished_at", time.Now().UTC().Format(time.RFC3339Nano))
	}
	if err := app.Save(record); err != nil {
		return nil, fmt.Errorf("save deployment state: %w", err)
	}
	return record, nil
}

// RecoverInterruptedDeployments marks stale in-progress jobs as failed after a
// control-plane restart. It deliberately does not replay commands on servers.
func RecoverInterruptedDeployments(app core.App, staleBefore time.Time) (int, error) {
	if app == nil {
		return 0, errors.New("PocketBase app is required")
	}
	if staleBefore.IsZero() {
		return 0, errors.New("stale deployment cutoff is required")
	}
	cutoff := staleBefore.UTC().Format(time.RFC3339Nano)
	recovered := 0
	for _, state := range interruptedDeploymentStates {
		filter := fmt.Sprintf(`status = "%s" && updated < "%s"`, state, cutoff)
		records, err := app.FindRecordsByFilter("arsy_deployments", filter, "updated", 0, 0)
		if err != nil {
			return recovered, fmt.Errorf("find interrupted %s deployments: %w", state, err)
		}
		for _, record := range records {
			next, err := Transition(State(record.GetString("status")), StateFailed)
			if err != nil {
				continue
			}
			record.Set("status", string(next))
			record.Set("finished_at", time.Now().UTC().Format(time.RFC3339Nano))
			record.Set("summary", "Deployment interrupted by a control-plane restart. Inspect the target host before retrying.")
			if err := app.Save(record); err != nil {
				return recovered, fmt.Errorf("mark interrupted deployment failed: %w", err)
			}
			recovered++
		}
	}
	return recovered, nil
}
