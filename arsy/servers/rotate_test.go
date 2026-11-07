package servers

import (
	"testing"

	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestRotateServerCredentialValidatesInput(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	if err := RotateServerCredential(nil, "", "", ""); err == nil {
		t.Fatal("expected error for nil app")
	}
}

func TestRotateServerCredentialRejectsUnknownServer(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	err = RotateServerCredential(app, "nonexistent", "-----BEGIN OPENSSH PRIVATE KEY-----\ntest\n-----END OPENSSH PRIVATE KEY-----", "ARSY_SECRET_KEY")
	if err == nil || err.Error()[:len("server not found")] != "server not found" {
		t.Fatalf("expected 'server not found' error, got: %v", err)
	}
}

func TestRotateServerCredentialRejectsEmptyKey(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	serverCol, _ := app.FindCollectionByNameOrId("arsy_servers")
	server := core.NewRecord(serverCol)
	server.Set("name", "RotateServer")
	server.Set("host", "rotate.example.test")
	server.Set("ssh_port", 22)
	server.Set("ssh_username", "deploy")
	if err := app.Save(server); err != nil {
		t.Fatal(err)
	}
	if err := RotateServerCredential(app, server.Id, "", "ARSY_SECRET_KEY"); err == nil || err.Error() != "new private key is required" {
		t.Fatalf("expected 'new private key is required', got: %v", err)
	}
}
