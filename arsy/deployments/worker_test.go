package deployments

import (
	"context"
	"errors"
	"testing"

	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestWorkerProcessesDeploymentThroughVerification(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	deployment := createDeploymentFixture(t, app, "worker-success", StateQueued)
	worker := Worker{
		App: app,
		Process: func(_ context.Context, record *core.Record) error {
			for _, next := range []State{StateUploading, StateValidating, StateDeploying, StateVerifying} {
				if _, err := UpdateDeploymentState(app, record.Id, next, "step completed"); err != nil {
					return err
				}
			}
			return nil
		},
	}
	processed, err := worker.ProcessNext(context.Background())
	if err != nil || !processed {
		t.Fatalf("ProcessNext() = (%v, %v)", processed, err)
	}
	latest, err := app.FindRecordById("arsy_deployments", deployment.Id)
	if err != nil {
		t.Fatal(err)
	}
	if latest.GetString("status") != string(StateSucceeded) || latest.GetString("finished_at") == "" {
		t.Fatalf("worker final state = %q", latest.GetString("status"))
	}
}

func TestWorkerPersistsFailureAndCancellation(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	failed := createDeploymentFixture(t, app, "worker-failed", StateQueued)
	worker := Worker{App: app, Process: func(context.Context, *core.Record) error {
		return errors.New("sensitive remote detail")
	}}
	if processed, err := worker.ProcessNext(context.Background()); err != nil || !processed {
		t.Fatalf("failed ProcessNext() = (%v, %v)", processed, err)
	}
	failedRecord, err := app.FindRecordById("arsy_deployments", failed.Id)
	if err != nil {
		t.Fatal(err)
	}
	if failedRecord.GetString("status") != string(StateFailed) || failedRecord.GetString("summary") == "sensitive remote detail" {
		t.Fatalf("failure state or summary was unsafe: %q / %q", failedRecord.GetString("status"), failedRecord.GetString("summary"))
	}

	cancelled := createDeploymentFixture(t, app, "worker-cancelled", StateQueued)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cancelWorker := Worker{App: app, Process: func(context.Context, *core.Record) error {
		return context.Canceled
	}}
	// A cancelled context is rejected before claiming, so this job remains queued.
	if processed, err := cancelWorker.ProcessNext(ctx); !errors.Is(err, context.Canceled) || processed {
		t.Fatalf("cancelled ProcessNext() = (%v, %v)", processed, err)
	}
	cancelledRecord, err := app.FindRecordById("arsy_deployments", cancelled.Id)
	if err != nil {
		t.Fatal(err)
	}
	if cancelledRecord.GetString("status") != string(StateQueued) {
		t.Fatalf("cancelled-before-claim job status = %q", cancelledRecord.GetString("status"))
	}
}

func TestWorkerMarksProcessorCancellationAsCancelled(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	deployment := createDeploymentFixture(t, app, "worker-cancel", StateQueued)
	ctx, cancel := context.WithCancel(context.Background())
	worker := Worker{App: app, Process: func(context.Context, *core.Record) error {
		cancel()
		return context.Canceled
	}}
	processed, err := worker.ProcessNext(ctx)
	if err != nil || !processed {
		t.Fatalf("ProcessNext() = (%v, %v)", processed, err)
	}
	latest, err := app.FindRecordById("arsy_deployments", deployment.Id)
	if err != nil {
		t.Fatal(err)
	}
	if latest.GetString("status") != string(StateCancelled) {
		t.Fatalf("cancelled deployment status = %q", latest.GetString("status"))
	}
}
