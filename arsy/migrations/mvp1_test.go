package migrations_test

import (
	"strings"
	"testing"

	_ "github.com/bridge-allah/a-r-s-y/arsy/migrations"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestMVP1CollectionsAreCreatedWithPrivateRules(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	for _, name := range []string{
		"arsy_projects",
		"arsy_servers",
		"arsy_applications",
		"arsy_deployments",
		"arsy_deployment_logs",
	} {
		collection, err := app.FindCollectionByNameOrId(name)
		if err != nil {
			t.Fatalf("collection %q was not created: %v", name, err)
		}
		if collection.ListRule != nil || collection.ViewRule != nil || collection.CreateRule != nil || collection.UpdateRule != nil || collection.DeleteRule != nil {
			t.Fatalf("collection %q must be superuser-only", name)
		}
	}

	projects, err := app.FindCollectionByNameOrId("arsy_projects")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := projects.Fields.AsMap()["slug"].(*core.TextField); !ok {
		t.Fatal("projects.slug must be a text field")
	}

	servers, err := app.FindCollectionByNameOrId("arsy_servers")
	if err != nil {
		t.Fatal(err)
	}
	credential, ok := servers.Fields.AsMap()["credential_ciphertext"].(*core.TextField)
	if !ok || !credential.Hidden {
		t.Fatal("server credential ciphertext must be a hidden text field")
	}
	pendingFingerprint, ok := servers.Fields.AsMap()["pending_host_key_fingerprint"].(*core.TextField)
	if !ok || !pendingFingerprint.Hidden {
		t.Fatal("pending SSH host-key fingerprint must be hidden")
	}

	logs, err := app.FindCollectionByNameOrId("arsy_deployment_logs")
	if err != nil {
		t.Fatal(err)
	}
	logDeployment, ok := logs.Fields.AsMap()["deployment"].(*core.RelationField)
	if !ok || logDeployment.CollectionId == "" {
		t.Fatal("deployment logs must relate to deployments")
	}

	applications, err := app.FindCollectionByNameOrId("arsy_applications")
	if err != nil {
		t.Fatal(err)
	}
	projectRelation, ok := applications.Fields.AsMap()["project"].(*core.RelationField)
	if !ok || projectRelation.CollectionId != projects.Id {
		t.Fatal("application project relation is invalid")
	}
	serverRelation, ok := applications.Fields.AsMap()["server"].(*core.RelationField)
	if !ok || serverRelation.CollectionId != servers.Id {
		t.Fatal("application server relation is invalid")
	}
	if !containsMigrationIndex(applications.Indexes, "idx_arsy_applications_server") {
		t.Fatal("applications.server must be unique for the MVP-1 one-application-per-server constraint")
	}
}

func containsMigrationIndex(indexes []string, name string) bool {
	for _, index := range indexes {
		if strings.Contains(index, name) {
			return true
		}
	}
	return false
}
