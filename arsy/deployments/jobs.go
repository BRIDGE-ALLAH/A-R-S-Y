package deployments

import (
	"errors"
	"fmt"
	"time"

	"github.com/bridge-allah/a-r-s-y/core"
)

var ErrNoQueuedDeployments = errors.New("no queued deployments are available")
var ErrDeploymentInProgress = errors.New("deployment in progress for this application")

var interruptedDeploymentStates = []State{
	StatePreparing,
	StateUploading,
	StateValidating,
	StateDeploying,
	StateVerifying,
}

// HasActiveDeployment checks if the application has any non-terminal deployment.
func HasActiveDeployment(app core.App, applicationID string) (bool, error) {
	if app == nil {
		return false, errors.New("PocketBase app is required")
	}
	records, err := app.FindRecordsByFilter(
		"arsy_deployments",
		`application = {:appID} && (status = "queued" || status = "preparing" || status = "uploading" || status = "validating" || status = "deploying" || status = "verifying")`,
		"",
		1, 0,
		map[string]any{"appID": applicationID},
	)
	if err != nil {
		return false, err
	}
	return len(records) > 0, nil
}

// FindDeploymentByIdempotencyKey returns the deployment with the given
// idempotency key, or nil if none exists.
func FindDeploymentByIdempotencyKey(app core.App, idempotencyKey string) (*core.Record, error) {
	if app == nil {
		return nil, errors.New("PocketBase app is required")
	}
	if idempotencyKey == "" {
		return nil, nil
	}
	record, err := app.FindFirstRecordByData("arsy_deployments", "idempotency_key", idempotencyKey)
	if err != nil {
		return nil, nil
	}
	return record, nil
}

// RecoverStaleDeploymentLocks marks any non-terminal deployment started
// more than 1 hour ago as failed. Called once on worker startup.
func RecoverStaleDeploymentLocks(app core.App) error {
	if app == nil {
		return errors.New("PocketBase app is required")
	}
	const staleThreshold = 1 * time.Hour
	cutoff := time.Now().UTC().Add(-staleThreshold)
	records, err := app.FindRecordsByFilter(
		"arsy_deployments",
		`(status = "queued" || status = "preparing" || status = "uploading" || status = "validating" || status = "deploying" || status = "verifying") && created < {:cutoff}`,
		"",
		1000, 0,
		map[string]any{"cutoff": cutoff.Format(time.RFC3339Nano)},
	)
	if err != nil {
		return fmt.Errorf("query stale deployments: %w", err)
	}
	for _, record := range records {
		record.Set("status", string(StateFailed))
		record.Set("summary", "Deployment lock recovered after process restart.")
		if err := app.Save(record); err != nil {
			return fmt.Errorf("recover stale deployment %q: %w", record.GetString("deployment_id"), err)
		}
	}
	return nil
}

// ClaimNextQueuedDeployment atomically claims the oldest queued deployment
// inside a PocketBase transaction so concurrent workers cannot claim it twice.
func ClaimNextQueuedDeployment(app core.App) (*core.Record, error) {
	if app == nil {
		return nil, errors.New("PocketBase app is required")
	}
	var claimed *core.Record
	err := app.RunInTransaction(func(txApp core.App) error {
		records, err := txApp.FindRecordsByFilter("arsy_deployments", `status = "queued"`, "created", 100, 0)
		if err != nil {
			return fmt.Errorf("find queued deployments: %w", err)
		}
		if len(records) == 0 {
			return ErrNoQueuedDeployments
		}
		var record *core.Record
		for _, candidate := range records {
			appID := candidate.GetString("application")
			active, err := HasActiveDeployment(txApp, appID)
			if err != nil {
				return err
			}
			if !active {
				record = candidate
				break
			}
		}
		if record == nil {
			return ErrNoQueuedDeployments
		}
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

