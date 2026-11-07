package deployments

import (
	"errors"
	"sync"
	"testing"
	"time"

	_ "github.com/bridge-allah/a-r-s-y/arsy/migrations"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestClaimNextQueuedDeploymentIsAtomic(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	deployment := createDeploymentFixture(t, app, "claim-test", StateQueued)

	var wg sync.WaitGroup
	wg.Add(2)
	results := make(chan *core.Record, 2)
	errorsOut := make(chan error, 2)
	for range 2 {
		go func() {
			defer wg.Done()
			record, claimErr := ClaimNextQueuedDeployment(app)
			results <- record
			errorsOut <- claimErr
		}()
	}
	wg.Wait()
	close(results)
	close(errorsOut)

	claimed := 0
	noJobs := 0
	for err := range errorsOut {
		switch {
		case err == nil:
			claimed++
		case errors.Is(err, ErrNoQueuedDeployments):
			noJobs++
		default:
			t.Fatalf("claim error = %v", err)
		}
	}
	if claimed != 1 || noJobs != 1 {
		t.Fatalf("successful claims=%d, empty claims=%d; want 1 each", claimed, noJobs)
	}
	record, err := app.FindRecordById("arsy_deployments", deployment.Id)
	if err != nil {
		t.Fatal(err)
	}
	if record.GetString("status") != string(StatePreparing) || record.GetString("started_at") == "" {
		t.Fatalf("claimed deployment state = %q, started_at=%q", record.GetString("status"), record.GetString("started_at"))
	}
}

func TestUpdateDeploymentStateUsesTransitionRules(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	deployment := createDeploymentFixture(t, app, "transition-test", StateQueued)

	if _, err := UpdateDeploymentState(app, deployment.Id, StateSucceeded, "not allowed"); err == nil {
		t.Fatal("queued deployment transitioned directly to succeeded")
	}
	claimed, err := ClaimNextQueuedDeployment(app)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := UpdateDeploymentState(app, claimed.Id, StateUploading, "manifest transfer started")
	if err != nil {
		t.Fatal(err)
	}
	if updated.GetString("status") != string(StateUploading) || updated.GetString("summary") != "manifest transfer started" {
		t.Fatalf("updated state = %q summary=%q", updated.GetString("status"), updated.GetString("summary"))
	}
}

func TestRecoverInterruptedDeploymentsMarksInProgressJobsFailed(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	deployment := createDeploymentFixture(t, app, "recovery-test", StateUploading)

	count, err := RecoverInterruptedDeployments(app, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("recovered deployments = %d, want 1", count)
	}
	recovered, err := app.FindRecordById("arsy_deployments", deployment.Id)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.GetString("status") != string(StateFailed) || recovered.GetString("finished_at") == "" || recovered.GetString("summary") == "" {
		t.Fatalf("recovered deployment state = %#v", recovered)
	}
}

func createDeploymentFixture(t *testing.T, app core.App, slug string, state State) *core.Record {
	t.Helper()
	projectCollection, err := app.FindCachedCollectionByNameOrId("arsy_projects")
	if err != nil {
		t.Fatal(err)
	}
	project := core.NewRecord(projectCollection)
	project.Set("name", "Job test project "+slug)
	project.Set("slug", slug)
	if err := app.Save(project); err != nil {
		t.Fatal(err)
	}

	serverCollection, err := app.FindCachedCollectionByNameOrId("arsy_servers")
	if err != nil {
		t.Fatal(err)
	}
	server := core.NewRecord(serverCollection)
	server.Set("name", "Job test server "+slug)
	server.Set("host", slug+".example.test")
	server.Set("ssh_port", 22)
	server.Set("ssh_username", "deploy")
	server.Set("readiness_state", "ready")
	if err := app.Save(server); err != nil {
		t.Fatal(err)
	}

	applicationCollection, err := app.FindCachedCollectionByNameOrId("arsy_applications")
	if err != nil {
		t.Fatal(err)
	}
	application := core.NewRecord(applicationCollection)
	application.Set("project", project.Id)
	application.Set("server", server.Id)
	application.Set("name", "Job test application "+slug)
	application.Set("slug", slug)
	application.Set("compose_manifest", "services:\n  web:\n    image: nginx\n")
	application.Set("lifecycle_state", "ready")
	if err := app.Save(application); err != nil {
		t.Fatal(err)
	}

	deploymentID, err := NewDeploymentID()
	if err != nil {
		t.Fatal(err)
	}
	deploymentCollection, err := app.FindCachedCollectionByNameOrId("arsy_deployments")
	if err != nil {
		t.Fatal(err)
	}
	deployment := core.NewRecord(deploymentCollection)
	deployment.Set("application", application.Id)
	deployment.Set("deployment_id", deploymentID)
	deployment.Set("status", string(state))
	if err := app.Save(deployment); err != nil {
		t.Fatal(err)
	}
	return deployment
}
