package migrations_test

import (
	"testing"

	_ "github.com/bridge-allah/a-r-s-y/arsy/migrations"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestDeploymentsCollectionHasHiddenManifestSnapshotField(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	collection, err := app.FindCollectionByNameOrId("arsy_deployments")
	if err != nil {
		t.Fatal(err)
	}
	field, ok := collection.Fields.AsMap()["manifest_ciphertext"].(*core.TextField)
	if !ok {
		t.Fatal("arsy_deployments.manifest_ciphertext must be a text field")
	}
	if !field.Hidden {
		t.Fatal("arsy_deployments.manifest_ciphertext must be hidden")
	}
}
