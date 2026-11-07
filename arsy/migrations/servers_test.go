package migrations

import (
	"testing"

	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestServersCollectionMigration(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	if err := app.RunSystemMigrations(); err != nil {
		t.Fatalf("apply system migrations: %v", err)
	}
	collection, err := app.FindCollectionByNameOrId(serversCollectionName)
	if err != nil {
		t.Fatalf("find servers collection: %v", err)
	}
	if collection.ListRule != nil || collection.ViewRule != nil || collection.CreateRule != nil || collection.UpdateRule != nil || collection.DeleteRule != nil {
		t.Fatal("servers collection must remain superuser-only")
	}

	name, ok := collection.Fields.GetByName("name").(*core.TextField)
	if !ok || !name.Required {
		t.Fatal("server name must be required")
	}
	port, ok := collection.Fields.GetByName("ssh_port").(*core.NumberField)
	if !ok || !port.Required || !port.OnlyInt || port.Min == nil || *port.Min != 1 || port.Max == nil || *port.Max != 65535 {
		t.Fatal("SSH port must be a required integer from 1 to 65535")
	}
	username, ok := collection.Fields.GetByName("ssh_username").(*core.TextField)
	if !ok || !username.Required || username.Pattern != "^[a-z_][a-z0-9_-]*[$]?$" {
		t.Fatal("SSH username must have the expected validation")
	}
	for _, fieldName := range []string{"host", "os_name", "os_version", "host_key_fingerprint", "readiness_state"} {
		if collection.Fields.GetByName(fieldName) == nil {
			t.Fatalf("servers collection is missing %s field", fieldName)
		}
	}
	privateKey, ok := collection.Fields.GetByName("private_key_ciphertext").(*core.TextField)
	if !ok || !privateKey.Hidden {
		t.Fatal("private-key ciphertext must be stored in a hidden field")
	}
	for _, fieldName := range []string{"created", "updated"} {
		if _, ok := collection.Fields.GetByName(fieldName).(*core.AutodateField); !ok {
			t.Fatalf("servers collection is missing %s timestamp", fieldName)
		}
	}

	valid := core.NewRecord(collection)
	valid.Set("name", "Production")
	valid.Set("host", "example.com")
	valid.Set("ssh_port", 22)
	valid.Set("ssh_username", "deploy")
	valid.Set("readiness_state", "pending")
	valid.Set("private_key_ciphertext", "ciphertext-fixture")
	if err := app.Save(valid); err != nil {
		t.Fatalf("save valid server: %v", err)
	}
	if got := valid.PublicExport(); got["private_key_ciphertext"] != nil {
		t.Fatal("server public export exposed private-key ciphertext")
	}
	if valid.GetString("created") == "" || valid.GetString("updated") == "" {
		t.Fatal("server timestamps were not populated")
	}

	invalidPort := core.NewRecord(collection)
	invalidPort.Set("name", "Invalid")
	invalidPort.Set("host", "example.com")
	invalidPort.Set("ssh_port", 70000)
	invalidPort.Set("ssh_username", "deploy")
	if err := app.Save(invalidPort); err == nil {
		t.Fatal("expected out-of-range SSH port to be rejected")
	}
}
