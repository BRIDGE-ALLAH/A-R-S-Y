package deployments

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestWorkerRespectsContextTimeout(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	deployment := createDeploymentFixture(t, app, "worker-timeout", StateQueued)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond)
	worker := Worker{App: app, Process: func(context.Context, *core.Record) error {
		return nil
	}}
	_, err = worker.ProcessNext(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ProcessNext() with timeout error = %v, want context.DeadlineExceeded", err)
	}
	record, _ := app.FindRecordById("arsy_deployments", deployment.Id)
	if record.GetString("status") != string(StateQueued) {
		t.Fatalf("timed-out unclaimed job status = %q, want queued", record.GetString("status"))
	}
}

func TestClaimNextQueuedDeploymentEnforcesAtomicity(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	dep1 := createDeploymentFixture(t, app, "atomic-app1", StateQueued)
	serverCol, _ := app.FindCollectionByNameOrId("arsy_servers")
	server2 := core.NewRecord(serverCol)
	server2.Set("name", "Server2")
	server2.Set("host", "server2.test")
	server2.Set("ssh_port", 22)
	server2.Set("ssh_username", "deploy")
	if err := app.Save(server2); err != nil {
		t.Fatal(err)
	}
	appCol, _ := app.FindCollectionByNameOrId("arsy_applications")
	app2 := core.NewRecord(appCol)
	projCol, _ := app.FindCollectionByNameOrId("arsy_projects")
	proj2 := core.NewRecord(projCol)
	proj2.Set("name", "Project2")
	proj2.Set("slug", "atomic-proj2")
	if err := app.Save(proj2); err != nil {
		t.Fatal(err)
	}
	app2.Set("project", proj2.Id)
	app2.Set("server", server2.Id)
	app2.Set("name", "App2")
	app2.Set("slug", "atomic-app2")
	app2.Set("compose_manifest", "services:\n  web:\n    image: nginx\n")
	app2.Set("lifecycle_state", "ready")
	if err := app.Save(app2); err != nil {
		t.Fatal(err)
	}
	
	depCol, _ := app.FindCollectionByNameOrId("arsy_deployments")
	dep2 := core.NewRecord(depCol)
	dep2ID, _ := NewDeploymentID()
	dep2.Set("application", app2.Id)
	dep2.Set("deployment_id", dep2ID)
	dep2.Set("status", string(StateQueued))
	if err := app.Save(dep2); err != nil {
		t.Fatal(err)
	}

	claimed1, err := ClaimNextQueuedDeployment(app)
	if err != nil || claimed1 == nil || claimed1.Id != dep1.Id {
		t.Fatalf("first claim: got %v, want dep1", claimed1)
	}
	claimed2, err := ClaimNextQueuedDeployment(app)
	if err != nil || claimed2 == nil || claimed2.Id != dep2.Id {
		t.Fatalf("second claim: got %v, want dep2", claimed2)
	}
	if claimed1.GetString("status") != string(StatePreparing) || claimed2.GetString("status") != string(StatePreparing) {
		t.Fatal("both claimed deployments not in preparing state")
	}
	claimed3, err := ClaimNextQueuedDeployment(app)
	if err != ErrNoQueuedDeployments || claimed3 != nil {
		t.Fatalf("third claim: got (%v, %v), want (nil, ErrNoQueuedDeployments)", claimed3, err)
	}
}

