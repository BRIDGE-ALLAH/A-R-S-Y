package migrations

import (
	"testing"

	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestApplicationsCollectionMigration(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	if err := app.RunSystemMigrations(); err != nil {
		t.Fatalf("apply system migrations: %v", err)
	}

	projects, err := app.FindCollectionByNameOrId(projectsCollectionName)
	if err != nil {
		t.Fatal(err)
	}
	servers, err := app.FindCollectionByNameOrId(serversCollectionName)
	if err != nil {
		t.Fatal(err)
	}
	collection, err := app.FindCollectionByNameOrId(applicationsCollectionName)
	if err != nil {
		t.Fatalf("find applications collection: %v", err)
	}
	if collection.ListRule != nil || collection.ViewRule != nil || collection.CreateRule != nil || collection.UpdateRule != nil || collection.DeleteRule != nil {
		t.Fatal("applications collection must remain superuser-only")
	}
	projectField, ok := collection.Fields.GetByName("project").(*core.RelationField)
	if !ok || !projectField.Required || projectField.CollectionId != projects.Id {
		t.Fatal("application must require a project relation")
	}
	serverField, ok := collection.Fields.GetByName("server").(*core.RelationField)
	if !ok || !serverField.Required || serverField.CollectionId != servers.Id || serverField.MaxSelect > 1 {
		t.Fatal("application must map to exactly one server")
	}
	slug, ok := collection.Fields.GetByName("slug").(*core.TextField)
	if !ok || !slug.Required || slug.Pattern != "^[a-z0-9]+(-[a-z0-9]+)*$" {
		t.Fatal("application slug must be required and validated")
	}
	manifest, ok := collection.Fields.GetByName("compose_manifest").(*core.EditorField)
	if !ok || !manifest.Required || manifest.MaxSize != 1048576 {
		t.Fatal("Compose manifest must be required and bounded to 1 MiB")
	}
	if collection.Fields.GetByName("lifecycle_state") == nil {
		t.Fatal("applications collection is missing lifecycle state")
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
	application := core.NewRecord(collection)
	application.Set("project", project.Id)
	application.Set("server", server.Id)
	application.Set("name", "Web")
	application.Set("slug", "web")
	application.Set("compose_manifest", "services:\n  web:\n    image: nginx:alpine\n")
	application.Set("lifecycle_state", "draft")
	if err := app.Save(application); err != nil {
		t.Fatalf("save valid application: %v", err)
	}

	missingRelation := core.NewRecord(collection)
	missingRelation.Set("name", "Invalid")
	missingRelation.Set("slug", "invalid")
	missingRelation.Set("compose_manifest", "services: {}")
	if err := app.Save(missingRelation); err == nil {
		t.Fatal("expected missing project/server relation to be rejected")
	}
}
