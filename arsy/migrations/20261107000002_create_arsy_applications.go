package migrations

import (
	"database/sql"
	"errors"

	"github.com/bridge-allah/a-r-s-y/core"
)

const applicationsCollectionName = "applications"

func init() {
	core.SystemMigrations.Register(createApplicationsCollection, deleteApplicationsCollection)
}

func createApplicationsCollection(txApp core.App) error {
	_, err := txApp.FindCollectionByNameOrId(applicationsCollectionName)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	projects, err := txApp.FindCollectionByNameOrId(projectsCollectionName)
	if err != nil {
		return err
	}
	servers, err := txApp.FindCollectionByNameOrId(serversCollectionName)
	if err != nil {
		return err
	}

	collection := core.NewBaseCollection(applicationsCollectionName)
	collection.Fields.Add(&core.RelationField{Name: "project", Required: true, CollectionId: projects.Id, CascadeDelete: true})
	collection.Fields.Add(&core.RelationField{Name: "server", Required: true, CollectionId: servers.Id, CascadeDelete: false})
	collection.Fields.Add(&core.TextField{Name: "name", Required: true, Max: 120})
	collection.Fields.Add(&core.TextField{Name: "slug", Required: true, Max: 63, Pattern: "^[a-z0-9]+(-[a-z0-9]+)*$"})
	collection.Fields.Add(&core.EditorField{Name: "compose_manifest", Required: true, MaxSize: 1048576})
	collection.Fields.Add(&core.SelectField{Name: "lifecycle_state", Values: []string{"draft", "ready", "deploying", "deployed", "stopped", "failed"}})
	collection.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
	collection.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
	collection.AddIndex("idx_arsy_applications_project_slug", true, "project, slug", "")

	return txApp.Save(collection)
}

func deleteApplicationsCollection(txApp core.App) error {
	collection, err := txApp.FindCollectionByNameOrId(applicationsCollectionName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return txApp.Delete(collection)
}
