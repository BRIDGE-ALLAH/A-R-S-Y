package httpapi_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/servers"
	"github.com/bridge-allah/a-r-s-y/core"
	"golang.org/x/crypto/ssh"
)

func testOpenSSHKey(t *testing.T) []byte {
	t.Helper()
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	block, err := ssh.MarshalPrivateKey(privateKey, "arsy-stop-test")
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(block)
}

type stopFixture struct {
	project     *core.Record
	server      *core.Record
	application *core.Record
	deployment  *core.Record
}

func seedStopFixture(t *testing.T, app core.App, running bool) stopFixture {
	t.Helper()
	projCol, _ := app.FindCollectionByNameOrId("arsy_projects")
	srvCol, _ := app.FindCollectionByNameOrId("arsy_servers")
	appCol, _ := app.FindCollectionByNameOrId("arsy_applications")
	depCol, _ := app.FindCollectionByNameOrId("arsy_deployments")

	proj := core.NewRecord(projCol)
	proj.Set("name", "Stop Project")
	proj.Set("slug", "stop-project")
	if err := app.Save(proj); err != nil {
		t.Fatal(err)
	}
	srv := core.NewRecord(srvCol)
	srv.Set("name", "Stop Server")
	srv.Set("host", "127.0.0.1")
	srv.Set("ssh_port", 22)
	srv.Set("ssh_username", "deploy")
	srv.Set("host_key_fingerprint", "SHA256:pinned")
	if err := app.Save(srv); err != nil {
		t.Fatal(err)
	}
	appl := core.NewRecord(appCol)
	appl.Set("project", proj.Id)
	appl.Set("server", srv.Id)
	appl.Set("name", "Stop App")
	appl.Set("slug", "stop-app")
	appl.Set("manifest_ciphertext", "arsy:v1:placeholder")
	if running {
		appl.Set("lifecycle_state", "running")
	} else {
		appl.Set("lifecycle_state", "draft")
	}
	if err := app.Save(appl); err != nil {
		t.Fatal(err)
	}
	dep := core.NewRecord(depCol)
	dep.Set("application", appl.Id)
	dep.Set("deployment_id", "0123456789abcdef0123456789abcdef")
	dep.Set("status", "succeeded")
	if err := app.Save(dep); err != nil {
		t.Fatal(err)
	}
	return stopFixture{project: proj, server: srv, application: appl, deployment: dep}
}

func TestStopApplicationRejectsNonRunningApplication(t *testing.T) {
	app, mux := newTestMux(t, true)
	defer app.Cleanup()

	fixture := seedStopFixture(t, app, false)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/arsy/v1/applications/"+fixture.application.Id+"/stop", nil))
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for non-running application, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStopApplicationRejectsRevokedHostKey(t *testing.T) {
	app, mux := newTestMux(t, true)
	defer app.Cleanup()

	fixture := seedStopFixture(t, app, true)
	fixture.server.Set("host_key_revoked", true)
	if err := app.Save(fixture.server); err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/arsy/v1/applications/"+fixture.application.Id+"/stop", nil))
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for revoked host key, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStopApplicationRequiresPinnedHostKey(t *testing.T) {
	app, mux := newTestMux(t, true)
	defer app.Cleanup()

	fixture := seedStopFixture(t, app, true)
	fixture.server.Set("host_key_fingerprint", "")
	if err := app.Save(fixture.server); err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/arsy/v1/applications/"+fixture.application.Id+"/stop", nil))
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 without a pinned host key, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStopApplicationRequiresSucceededDeployment(t *testing.T) {
	app, mux := newTestMux(t, true)
	defer app.Cleanup()

	fixture := seedStopFixture(t, app, true)
	if err := app.Delete(fixture.deployment); err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/arsy/v1/applications/"+fixture.application.Id+"/stop", nil))
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 without a succeeded deployment, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStopApplicationRejectsActiveDeployment(t *testing.T) {
	app, mux := newTestMux(t, true)
	defer app.Cleanup()

	fixture := seedStopFixture(t, app, true)
	depCol, _ := app.FindCollectionByNameOrId("arsy_deployments")
	active := core.NewRecord(depCol)
	active.Set("application", fixture.application.Id)
	active.Set("deployment_id", "fedcba9876543210fedcba9876543210")
	active.Set("status", "deploying")
	if err := app.Save(active); err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/arsy/v1/applications/"+fixture.application.Id+"/stop", nil))
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 while a deployment is active, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStopApplicationHappyPathAgainstTestSSHServer(t *testing.T) {
	app, mux := newTestMux(t, true)
	defer app.Cleanup()

	sshServer, err := servers.NewTestSSHServer(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer sshServer.Close()

	fixture := seedStopFixture(t, app, true)
	fixture.server.Set("host", "127.0.0.1")
	fixture.server.Set("ssh_port", sshServer.Port)
	fixture.server.Set("host_key_fingerprint", sshServer.HostKeyFingerprint())
	if err := servers.StorePrivateKey(fixture.server, testOpenSSHKey(t), "ARSY_SECRET_KEY"); err != nil {
		t.Fatal(err)
	}
	if err := app.Save(fixture.server); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/arsy/v1/applications/"+fixture.application.Id+"/stop", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["lifecycle_state"] != "stopped" {
		t.Fatalf("expected stopped lifecycle, got %v", body["lifecycle_state"])
	}

	reloaded, err := app.FindRecordById("arsy_applications", fixture.application.Id)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.GetString("lifecycle_state") != "stopped" {
		t.Fatalf("persisted lifecycle = %q, want stopped", reloaded.GetString("lifecycle_state"))
	}

	audit, err := app.FindRecordsByFilter("arsy_audit_log", "event_type = {:t}", "-created", 1, 0, map[string]any{"t": "application_stopped"})
	if err != nil {
		t.Fatal(err)
	}
	if len(audit) != 1 {
		t.Fatalf("expected one application_stopped audit event, got %d", len(audit))
	}
}
