package migrations

import (
	"strings"
	"testing"

	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestProjectsCollectionMigration(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	if err := app.RunSystemMigrations(); err != nil {
		t.Fatalf("apply system migrations: %v", err)
	}
	collection, err := app.FindCollectionByNameOrId(projectsCollectionName)
	if err != nil {
		t.Fatalf("find projects collection: %v", err)
	}

	nameField, ok := collection.Fields.GetByName("name").(*core.TextField)
	if !ok || !nameField.Required || nameField.Max != 120 {
		t.Fatalf("unexpected projects name field: %#v", collection.Fields.GetByName("name"))
	}
	slugField, ok := collection.Fields.GetByName("slug").(*core.TextField)
	if !ok || !slugField.Required || slugField.Pattern != "^[a-z0-9]+(-[a-z0-9]+)*$" {
		t.Fatalf("unexpected projects slug field: %#v", collection.Fields.GetByName("slug"))
	}
	if collection.Fields.GetByName("description") == nil {
		t.Fatal("projects collection is missing description field")
	}
	for _, fieldName := range []string{"created", "updated"} {
		field, ok := collection.Fields.GetByName(fieldName).(*core.AutodateField)
		if !ok || !field.OnCreate || (fieldName == "updated" && !field.OnUpdate) {
			t.Fatalf("projects collection is missing configured %s timestamp", fieldName)
		}
	}
	if collection.ListRule != nil || collection.ViewRule != nil || collection.CreateRule != nil ||
		collection.UpdateRule != nil || collection.DeleteRule != nil {
		t.Fatal("projects collection must remain superuser-only")
	}
	if !strings.Contains(strings.Join(collection.Indexes, "\n"), "UNIQUE INDEX `idx_arsy_projects_slug`") {
		t.Fatalf("projects collection is missing its unique slug index: %v", collection.Indexes)
	}

	invalid := core.NewRecord(collection)
	invalid.Set("name", "Invalid project")
	invalid.Set("slug", "Invalid_Slug")
	if err := app.Save(invalid); err == nil {
		t.Fatal("expected invalid project slug to be rejected")
	}

	first := core.NewRecord(collection)
	first.Set("name", "Main project")
	first.Set("slug", "main-project")
	if err := app.Save(first); err != nil {
		t.Fatalf("save valid project: %v", err)
	}
	if first.GetString("created") == "" || first.GetString("updated") == "" {
		t.Fatal("project timestamps were not populated")
	}
	duplicate := core.NewRecord(collection)
	duplicate.Set("name", "Duplicate project")
	duplicate.Set("slug", "main-project")
	if err := app.Save(duplicate); err == nil {
		t.Fatal("expected duplicate project slug to be rejected")
	}
}
