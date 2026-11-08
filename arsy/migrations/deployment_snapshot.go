package migrations

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/bridge-allah/a-r-s-y/core"
)

// deploymentManifestSnapshotField stores the encrypted Compose manifest that a
// specific deployment actually uploaded. It enables historical rollback: the
// application manifest may change over time, but each deployment keeps the exact
// revision it deployed.
const deploymentManifestSnapshotField = "manifest_ciphertext"

func init() {
	core.SystemMigrations.Register(upDeploymentSnapshot, downDeploymentSnapshot, "20261108050000_arsy_deployment_snapshot.go")
}

func upDeploymentSnapshot(app core.App) error {
	collection, err := app.FindCollectionByNameOrId(deploymentsCollectionName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("A-R-S-Y deployments collection is missing; run the base migration first")
		}
		return fmt.Errorf("find A-R-S-Y deployments collection: %w", err)
	}

	if _, ok := collection.Fields.AsMap()[deploymentManifestSnapshotField]; ok {
		return nil
	}

	collection.Fields.Add(&core.TextField{
		Name:   deploymentManifestSnapshotField,
		Hidden: true,
		Max:    100000,
	})
	if err := app.Save(collection); err != nil {
		return fmt.Errorf("add deployment manifest snapshot field: %w", err)
	}
	return nil
}

func downDeploymentSnapshot(app core.App) error {
	collection, err := app.FindCollectionByNameOrId(deploymentsCollectionName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("find A-R-S-Y deployments collection: %w", err)
	}
	if _, ok := collection.Fields.AsMap()[deploymentManifestSnapshotField]; !ok {
		return nil
	}
	collection.Fields.RemoveByName(deploymentManifestSnapshotField)
	if err := app.Save(collection); err != nil {
		return fmt.Errorf("remove deployment manifest snapshot field: %w", err)
	}
	return nil
}
