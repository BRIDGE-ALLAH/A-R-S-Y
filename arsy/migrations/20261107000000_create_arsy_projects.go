package migrations

import (
	"database/sql"
	"errors"

	"github.com/bridge-allah/a-r-s-y/core"
)

const projectsCollectionName = "projects"

func init() {
	core.SystemMigrations.Register(createProjectsCollection, deleteProjectsCollection)
}

func createProjectsCollection(txApp core.App) error {
	_, err := txApp.FindCollectionByNameOrId(projectsCollectionName)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	collection := core.NewBaseCollection(projectsCollectionName)
	collection.Fields.Add(&core.TextField{
		Name:     "name",
		Required: true,
		Max:      120,
	})
	collection.Fields.Add(&core.TextField{
		Name:     "slug",
		Required: true,
		Max:      63,
		Pattern:  "^[a-z0-9]+(-[a-z0-9]+)*$",
	})
	collection.Fields.Add(&core.TextField{
		Name: "description",
		Max:  2000,
	})
	collection.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
	collection.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
	collection.AddIndex("idx_arsy_projects_slug", true, "slug", "")

	return txApp.Save(collection)
}

func deleteProjectsCollection(txApp core.App) error {
	collection, err := txApp.FindCollectionByNameOrId(projectsCollectionName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return txApp.Delete(collection)
}
