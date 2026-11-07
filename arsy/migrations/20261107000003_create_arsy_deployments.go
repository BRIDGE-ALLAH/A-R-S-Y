package migrations

import (
	"database/sql"
	"errors"

	"github.com/bridge-allah/a-r-s-y/core"
)

const deploymentsCollectionName = "deployments"

func init() {
	core.SystemMigrations.Register(createDeploymentsCollection, deleteDeploymentsCollection)
}

func createDeploymentsCollection(txApp core.App) error {
	_, err := txApp.FindCollectionByNameOrId(deploymentsCollectionName)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	applications, err := txApp.FindCollectionByNameOrId(applicationsCollectionName)
	if err != nil {
		return err
	}

	collection := core.NewBaseCollection(deploymentsCollectionName)
	collection.Fields.Add(&core.RelationField{Name: "application", Required: true, CollectionId: applications.Id, CascadeDelete: true})
	collection.Fields.Add(&core.TextField{Name: "deployment_id", Required: true, Max: 64, Pattern: "^[a-zA-Z0-9][a-zA-Z0-9_-]*$"})
	collection.Fields.Add(&core.SelectField{Name: "status", Required: true, Values: []string{"queued", "running", "succeeded", "failed", "cancelled"}})
	collection.Fields.Add(&core.DateField{Name: "started_at"})
	collection.Fields.Add(&core.DateField{Name: "ended_at"})
	collection.Fields.Add(&core.TextField{Name: "manifest_checksum", Max: 64, Pattern: "^[a-f0-9]{64}$"})
	collection.Fields.Add(&core.TextField{Name: "safe_summary", Max: 2000})
	collection.Fields.Add(&core.TextField{Name: "previous_deployment", Max: 64, Pattern: "^[a-zA-Z0-9][a-zA-Z0-9_-]*$"})
	collection.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
	collection.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
	collection.AddIndex("idx_arsy_deployments_application_id", false, "application, deployment_id", "")

	return txApp.Save(collection)
}

func deleteDeploymentsCollection(txApp core.App) error {
	collection, err := txApp.FindCollectionByNameOrId(deploymentsCollectionName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return txApp.Delete(collection)
}
