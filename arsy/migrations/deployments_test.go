package migrations

import (
	"testing"

	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestDeploymentsCollectionMigration(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	if err := app.RunSystemMigrations(); err != nil {
		t.Fatalf("apply system migrations: %v", err)
	}
	applications, err := app.FindCollectionByNameOrId(applicationsCollectionName)
	if err != nil {
		t.Fatal(err)
	}
	projects, err := app.FindCollectionByNameOrId(projectsCollectionName)
	if err != nil {
		t.Fatal(err)
	}
	servers, err := app.FindCollectionByNameOrId(serversCollectionName)
	if err != nil {
		t.Fatal(err)
	}
	collection, err := app.FindCollectionByNameOrId(deploymentsCollectionName)
	if err != nil {
		t.Fatalf("find deployments collection: %v", err)
	}
	if collection.ListRule != nil || collection.ViewRule != nil || collection.CreateRule != nil || collection.UpdateRule != nil || collection.DeleteRule != nil {
		t.Fatal("deployments collection must remain superuser-only")
	}
	applicationField, ok := collection.Fields.GetByName("application").(*core.RelationField)
	if !ok || !applicationField.Required || applicationField.CollectionId != applications.Id {
		t.Fatal("deployment must require an application relation")
	}
	status, ok := collection.Fields.GetByName("status").(*core.SelectField)
	if !ok || !status.Required || len(status.Values) != 5 {
		t.Fatal("deployment must define all required lifecycle statuses")
	}
	for _, required := range []string{"deployment_id", "started_at", "ended_at", "manifest_checksum", "safe_summary", "previous_deployment"} {
		if collection.Fields.GetByName(required) == nil {
			t.Fatalf("deployments collection is missing %s field", required)
		}
	}

	project := core.NewRecord(projects)
	project.Set("name", "Main project")
	project.Set("slug", "main-project")
	if err := app.Save(project); err != nil {
		t.Fatalf("save project fixture: %v", err)
	}
	server := core.NewRecord(servers)
	server.Set("name", "Production")
	server.Set("host", "example.com")
	server.Set("ssh_port", 22)
	server.Set("ssh_username", "deploy")
	if err := app.Save(server); err != nil {
		t.Fatalf("save server fixture: %v", err)
	}
	appRecord := core.NewRecord(applications)
	appRecord.Set("project", project.Id)
	appRecord.Set("server", server.Id)
	appRecord.Set("name", "Web")
	appRecord.Set("slug", "web")
	appRecord.Set("compose_manifest", "services: {}")
	if err := app.Save(appRecord); err != nil {
		t.Fatalf("save application fixture: %v", err)
	}

	for _, state := range []string{"queued", "running", "succeeded", "failed", "cancelled"} {
		record := core.NewRecord(collection)
		record.Set("application", appRecord.Id)
		record.Set("deployment_id", "deploy-"+state)
		record.Set("status", state)
		record.Set("safe_summary", "Validated deployment summary")
		if err := app.Save(record); err != nil {
			t.Fatalf("save %s deployment record: %v", state, err)
		}
	}
	invalid := core.NewRecord(collection)
	invalid.Set("application", appRecord.Id)
	invalid.Set("deployment_id", "invalid")
	invalid.Set("status", "arbitrary-command-output")
	if err := app.Save(invalid); err == nil {
		t.Fatal("expected unsupported deployment status to be rejected")
	}
}
