package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
)

var deploymentIDPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)

func createDeployApplicationFixture(t *testing.T, app *tests.TestApp, slug, composeManifest string) *core.Record {
	t.Helper()
	projectCollection, err := app.FindCollectionByNameOrId("arsy_projects")
	if err != nil {
		t.Fatal(err)
	}
	project := core.NewRecord(projectCollection)
	project.Set("name", "Deploy project "+slug)
	project.Set("slug", slug)
	if err := app.Save(project); err != nil {
		t.Fatal(err)
	}
	serverCollection, err := app.FindCollectionByNameOrId("arsy_servers")
	if err != nil {
		t.Fatal(err)
	}
	server := core.NewRecord(serverCollection)
	server.Set("name", "Deploy server "+slug)
	server.Set("host", slug+".example.test")
	server.Set("ssh_port", 22)
	server.Set("ssh_username", "deploy")
	if err := app.Save(server); err != nil {
		t.Fatal(err)
	}
	applicationCollection, err := app.FindCollectionByNameOrId("arsy_applications")
	if err != nil {
		t.Fatal(err)
	}
	application := core.NewRecord(applicationCollection)
	application.Set("project", project.Id)
	application.Set("server", server.Id)
	application.Set("name", "Deploy application "+slug)
	application.Set("slug", slug)
	application.Set("compose_manifest", composeManifest)
	application.Set("lifecycle_state", "ready")
	if err := app.Save(application); err != nil {
		t.Fatal(err)
	}
	return application
}

func TestCreateDeploymentQueuesApplication(t *testing.T) {
	app, mux := newTestMux(t, true)
	defer app.Cleanup()
	application := createDeployApplicationFixture(t, app, "deploy-app", "services:\n  web:\n    image: nginx\n")

	body := []byte(`{"application_id":"` + application.Id + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/arsy/v1/deployments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	deploymentID, _ := created["deployment_id"].(string)
	if !deploymentIDPattern.MatchString(deploymentID) {
		t.Fatalf("deployment_id = %q, want 32-char hex", deploymentID)
	}
	if created["status"] != "queued" {
		t.Fatalf("status = %v, want queued", created["status"])
	}
	if _, ok := created["previous_deployment_id"]; ok {
		t.Fatal("first deployment must not reference a previous deployment")
	}
	records, err := app.FindRecordsByFilter("arsy_deployments", "deployment_id = {:id}", "", 1, 0, map[string]any{"id": deploymentID})
	if err != nil || len(records) != 1 {
		t.Fatalf("queued deployment not persisted: err=%v records=%d", err, len(records))
	}

	body = []byte(`{"application_id":"` + application.Id + `"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/arsy/v1/deployments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on second deploy, got %d: %s", rec.Code, rec.Body.String())
	}
	var second map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &second); err != nil {
		t.Fatal(err)
	}
	// Second deployment: no succeeded predecessor, so no previous_deployment_id should be set
	if _, ok := second["previous_deployment_id"]; ok && second["previous_deployment_id"] != "" {
		t.Fatalf("second deployment has previous_deployment_id, but none succeeded: %v", second["previous_deployment_id"])
	}
	// Mark the first deployment as succeeded so the third can reference it
	firstDep, _ := app.FindFirstRecordByData("arsy_deployments", "deployment_id", deploymentID)
	firstDep.Set("status", "succeeded")
	_ = app.Save(firstDep)

	body = []byte(`{"application_id":"` + application.Id + `"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/arsy/v1/deployments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on third deploy, got %d: %s", rec.Code, rec.Body.String())
	}
	var third map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &third); err != nil {
		t.Fatal(err)
	}
	// Third deployment should reference the first (succeeded one)
	if third["previous_deployment_id"] != deploymentID {
		t.Fatalf("third previous should ref first succeeded, = %v, want %q", third["previous_deployment_id"], deploymentID)
	}
}

func TestCreateDeploymentRejectsBadRequests(t *testing.T) {
	app, mux := newTestMux(t, true)
	defer app.Cleanup()
	application := createDeployApplicationFixture(t, app, "deploy-bad", "services: {}")

	cases := []struct {
		name string
		body string
		want int
	}{
		{"missing application id", `{}`, http.StatusBadRequest},
		{"unknown application", `{"application_id":"nope123"}`, http.StatusNotFound},
		{"invalid stored manifest", `{"application_id":"` + application.Id + `"}`, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/arsy/v1/deployments", bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("expected %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCreateDeploymentRequiresSuperuser(t *testing.T) {
	app, mux := newTestMux(t, false)
	defer app.Cleanup()
	req := httptest.NewRequest(http.MethodPost, "/api/arsy/v1/deployments", bytes.NewReader([]byte(`{"application_id":"x"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestGetDeploymentProgressMapsStates(t *testing.T) {
	app, mux := newTestMux(t, true)
	defer app.Cleanup()
	application := createDeployApplicationFixture(t, app, "deploy-progress", "services:\n  web:\n    image: nginx\n")
	deploymentCollection, err := app.FindCollectionByNameOrId("arsy_deployments")
	if err != nil {
		t.Fatal(err)
	}

	progressStates := map[string]int{
		"queued":     0,
		"preparing":  10,
		"uploading":  30,
		"validating": 50,
		"deploying":  70,
		"verifying":  90,
		"succeeded":  100,
		"failed":     0,
		"cancelled":  0,
	}
	for state := range progressStates {
		record := core.NewRecord(deploymentCollection)
		record.Set("application", application.Id)
		record.Set("deployment_id", "progress-deployment-"+state)
		record.Set("status", state)
		if err := app.Save(record); err != nil {
			t.Fatal(err)
		}
	}
	for state, wantPercent := range progressStates {
		req := httptest.NewRequest(http.MethodGet, "/api/arsy/v1/deployments/progress-deployment-"+state+"/progress", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("state %s: expected 200, got %d: %s", state, rec.Code, rec.Body.String())
		}
		var progress map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &progress); err != nil {
			t.Fatal(err)
		}
		if progress["percent"].(float64) != float64(wantPercent) {
			t.Fatalf("state %s: percent = %v, want %d", state, progress["percent"], wantPercent)
		}
		if progress["state"] != state {
			t.Fatalf("state = %v, want %s", progress["state"], state)
		}
		step, _ := progress["step"].(string)
		if step == "" {
			t.Fatalf("state %s: step must not be empty", state)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/arsy/v1/deployments/does-not-exist/progress", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown deployment, got %d", rec.Code)
	}
}
