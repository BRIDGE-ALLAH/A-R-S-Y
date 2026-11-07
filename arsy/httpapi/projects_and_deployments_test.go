package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bridge-allah/a-r-s-y/core"
)

func TestProjectsAndApplicationsCRUD(t *testing.T) {
	app, mux := newTestMux(t, true)
	defer app.Cleanup()

	// 1. List projects initially empty
	req := httptest.NewRequest(http.MethodGet, "/api/arsy/v1/projects", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// 2. Create project
	body := []byte(`{"name":"Project Alpha","slug":"project-alpha","description":"Test project"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/arsy/v1/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var createdProject map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &createdProject); err != nil {
		t.Fatal(err)
	}
	projectID := createdProject["id"].(string)

	// 3. Create a server record directly in database for testing apps
	serversCol, err := app.FindCollectionByNameOrId("arsy_servers")
	if err != nil {
		t.Fatal(err)
	}
	serverRec := core.NewRecord(serversCol)
	serverRec.Set("name", "Host 1")
	serverRec.Set("host", "127.0.0.1")
	serverRec.Set("ssh_port", 22)
	serverRec.Set("ssh_username", "root")
	if err := app.Save(serverRec); err != nil {
		t.Fatal(err)
	}

	// 4. Create application
	appBody, _ := json.Marshal(map[string]any{
		"project_id":       projectID,
		"server_id":        serverRec.Id,
		"name":             "App One",
		"slug":             "app-one",
		"compose_manifest": "services:\n  web:\n    image: nginx",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/arsy/v1/applications", bytes.NewReader(appBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	// 5. List applications
	req = httptest.NewRequest(http.MethodGet, "/api/arsy/v1/applications", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestDeploymentsOperationsEndpoints(t *testing.T) {
	app, mux := newTestMux(t, true)
	defer app.Cleanup()

	// Create test deployment
	depCol, err := app.FindCollectionByNameOrId("arsy_deployments")
	if err != nil {
		t.Fatal(err)
	}

	// Create dummy application record
	appCol, _ := app.FindCollectionByNameOrId("arsy_applications")
	projCol, _ := app.FindCollectionByNameOrId("arsy_projects")
	srvCol, _ := app.FindCollectionByNameOrId("arsy_servers")

	proj := core.NewRecord(projCol)
	proj.Set("name", "P1")
	proj.Set("slug", "p1")
	app.Save(proj)

	srv := core.NewRecord(srvCol)
	srv.Set("name", "S1")
	srv.Set("host", "127.0.0.1")
	srv.Set("ssh_port", 2222)
	srv.Set("ssh_username", "user")
	app.Save(srv)

	appl := core.NewRecord(appCol)
	appl.Set("project", proj.Id)
	appl.Set("server", srv.Id)
	appl.Set("name", "A1")
	appl.Set("slug", "a1")
	appl.Set("compose_manifest", "services:\n  app:\n    image: redis")
	app.Save(appl)

	dep := core.NewRecord(depCol)
	dep.Set("application", appl.Id)
	dep.Set("deployment_id", "test-dep-001")
	dep.Set("status", "deploying")
	dep.Set("summary", "Sample log output")
	if err := app.Save(dep); err != nil {
		t.Fatal(err)
	}

	// 1. Get deployment
	req := httptest.NewRequest(http.MethodGet, "/api/arsy/v1/deployments/test-dep-001", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// 2. Get deployment logs
	req = httptest.NewRequest(http.MethodGet, "/api/arsy/v1/deployments/test-dep-001/logs", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// 3. Stop deployment
	req = httptest.NewRequest(http.MethodPost, "/api/arsy/v1/deployments/test-dep-001/stop", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Verify status updated to cancelled
	depReloaded, _ := app.FindFirstRecordByData("arsy_deployments", "deployment_id", "test-dep-001")
	if depReloaded.GetString("status") != "cancelled" {
		t.Fatalf("expected cancelled status, got %s", depReloaded.GetString("status"))
	}
}
