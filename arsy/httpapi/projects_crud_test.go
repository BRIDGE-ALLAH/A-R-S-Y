package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bridge-allah/a-r-s-y/core"
)

func TestProjectGetUpdateDeleteAndGuards(t *testing.T) {
	app, mux := newTestMux(t, true)
	defer app.Cleanup()

	body := []byte(`{"name":"CRUD Project","slug":"crud-project","description":"v1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/arsy/v1/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	id := created["id"].(string)

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/arsy/v1/projects/"+id, nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "CRUD Project") {
		t.Fatalf("get: %d %s", rec.Code, rec.Body.String())
	}

	patch := []byte(`{"name":"CRUD Project v2","description":"v2"}`)
	req = httptest.NewRequest(http.MethodPatch, "/api/arsy/v1/projects/"+id, bytes.NewReader(patch))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch: %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "CRUD Project v2") || !strings.Contains(rec.Body.String(), `"slug":"crud-project"`) {
		t.Fatalf("patch response: %s", rec.Body.String())
	}

	// create application to block delete
	serversCol, _ := app.FindCollectionByNameOrId("arsy_servers")
	serverRec := core.NewRecord(serversCol)
	serverRec.Set("name", "Host CRUD")
	serverRec.Set("host", "127.0.0.2")
	serverRec.Set("ssh_port", 22)
	serverRec.Set("ssh_username", "root")
	if err := app.Save(serverRec); err != nil {
		t.Fatal(err)
	}
	appBody, _ := json.Marshal(map[string]any{
		"project_id": id, "server_id": serverRec.Id, "name": "App CRUD", "slug": "app-crud",
		"compose_manifest": "services:\n  web:\n    image: nginx",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/arsy/v1/applications", bytes.NewReader(appBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create app: %d %s", rec.Code, rec.Body.String())
	}
	var createdApp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &createdApp)

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/arsy/v1/projects/"+id, nil))
	if rec.Code != http.StatusConflict {
		t.Fatalf("delete project with apps should 409, got %d %s", rec.Code, rec.Body.String())
	}

	// delete app then project
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/arsy/v1/applications/"+createdApp["id"].(string), nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("delete app: %d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/arsy/v1/projects/"+id, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("delete project: %d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/arsy/v1/projects/"+id, nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get deleted: %d", rec.Code)
	}
}

func TestApplicationGetUpdateRedactsManifest(t *testing.T) {
	app, mux := newTestMux(t, true)
	defer app.Cleanup()

	projCol, _ := app.FindCollectionByNameOrId("arsy_projects")
	proj := core.NewRecord(projCol)
	proj.Set("name", "P"); proj.Set("slug", "p-redact"); app.Save(proj)
	srvCol, _ := app.FindCollectionByNameOrId("arsy_servers")
	srv := core.NewRecord(srvCol)
	srv.Set("name", "S"); srv.Set("host", "127.0.0.3"); srv.Set("ssh_port", 22); srv.Set("ssh_username", "u"); app.Save(srv)

	appBody, _ := json.Marshal(map[string]any{
		"project_id": proj.Id, "server_id": srv.Id, "name": "A", "slug": "a-redact",
		"compose_manifest": "services:\n  web:\n    image: nginx",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/arsy/v1/applications", bytes.NewReader(appBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}
	var created map[string]any
	json.Unmarshal(rec.Body.Bytes(), &created)
	id := created["id"].(string)
	for _, forbidden := range []string{"manifest_ciphertext", "compose_manifest", "topsecret"} {
		if strings.Contains(rec.Body.String(), forbidden) {
			t.Fatalf("create leaked %q: %s", forbidden, rec.Body.String())
		}
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/arsy/v1/applications/"+id, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("get: %d %s", rec.Code, rec.Body.String())
	}
	for _, forbidden := range []string{"manifest_ciphertext", "compose_manifest", "topsecret"} {
		if strings.Contains(rec.Body.String(), forbidden) {
			t.Fatalf("get leaked %q: %s", forbidden, rec.Body.String())
		}
	}

	patch, _ := json.Marshal(map[string]any{"name": "A2", "compose_manifest": "services:\n  web:\n    image: redis"})
	req = httptest.NewRequest(http.MethodPatch, "/api/arsy/v1/applications/"+id, bytes.NewReader(patch))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "A2") {
		t.Fatalf("patch: %d %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "redis") || strings.Contains(rec.Body.String(), "manifest_ciphertext") {
		t.Fatalf("patch leaked manifest: %s", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/arsy/v1/applications/"+id, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("delete: %d %s", rec.Code, rec.Body.String())
	}
}

func TestApplicationDeleteBlockedByActiveDeployment(t *testing.T) {
	app, mux := newTestMux(t, true)
	defer app.Cleanup()

	projCol, _ := app.FindCollectionByNameOrId("arsy_projects")
	proj := core.NewRecord(projCol); proj.Set("name", "P"); proj.Set("slug", "p-active"); app.Save(proj)
	srvCol, _ := app.FindCollectionByNameOrId("arsy_servers")
	srv := core.NewRecord(srvCol); srv.Set("name", "S"); srv.Set("host", "127.0.0.4"); srv.Set("ssh_port", 22); srv.Set("ssh_username", "u"); app.Save(srv)
	appCol, _ := app.FindCollectionByNameOrId("arsy_applications")
	appl := core.NewRecord(appCol)
	appl.Set("project", proj.Id); appl.Set("server", srv.Id); appl.Set("name", "A"); appl.Set("slug", "a-active"); app.Save(appl)
	depCol, _ := app.FindCollectionByNameOrId("arsy_deployments")
	dep := core.NewRecord(depCol)
	dep.Set("application", appl.Id); dep.Set("deployment_id", "active00000000000000000000000001"); dep.Set("status", "deploying"); app.Save(dep)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/arsy/v1/applications/"+appl.Id, nil))
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d %s", rec.Code, rec.Body.String())
	}
}
