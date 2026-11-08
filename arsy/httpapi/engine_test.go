package httpapi_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/engine/registry"
	"github.com/bridge-allah/a-r-s-y/arsy/httpapi"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
	"github.com/bridge-allah/a-r-s-y/tools/router"
	_ "modernc.org/sqlite"
)

const engineTestModel = "test.item"

type engineFixture struct {
	app  *tests.TestApp
	mux  http.Handler
	dirs config.BaaSDirectories
}

func newEngineFixture(t *testing.T, authenticated bool) *engineFixture {
	t.Helper()
	t.Setenv("ARSY_SECRET_KEY", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.Cleanup)

	dirs, err := config.EnsureBaaSDirectories(t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}

	var auth *core.Record
	if authenticated {
		collection, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
		if err != nil {
			t.Fatal(err)
		}
		auth = core.NewRecord(collection)
	}

	reg := registry.New()
	if err := reg.Register(registry.Descriptor{
		ID:          engineTestModel,
		Label:       "Item",
		TableName:   "eng_items",
		Version:     "1.0.0",
		OwnerModule: "test",
		Source:      registry.SourceStaticGo,
	}); err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter[*core.RequestEvent](func(w http.ResponseWriter, req *http.Request) (*core.RequestEvent, router.EventCleanupFunc) {
		event := &core.RequestEvent{App: app, Auth: auth}
		event.Request = req
		event.Response = w
		return event, nil
	})
	cfg := config.Config{
		DataDir:      t.TempDir(),
		APIPrefix:    "/api/arsy/v1",
		ConsolePath:  "/arsy",
		SecretKeyEnv: "ARSY_SECRET_KEY",
	}
	httpapi.RegisterWithEngine(app, cfg, "test-version", httpapi.EngineConfig{Registry: reg, TenantsRoot: dirs.TenantsRoot})

	if err := app.OnServe().Trigger(&core.ServeEvent{App: app, Router: r}, func(e *core.ServeEvent) error {
		return e.Next()
	}); err != nil {
		t.Fatal(err)
	}
	mux, err := r.BuildMux()
	if err != nil {
		t.Fatal(err)
	}
	return &engineFixture{app: app, mux: mux, dirs: dirs}
}

func (f *engineFixture) createTenant(t *testing.T, slug, status string) string {
	t.Helper()
	collection, err := f.app.FindCollectionByNameOrId("baas_tenants")
	if err != nil {
		t.Fatal(err)
	}
	record := core.NewRecord(collection)
	record.Set("slug", slug)
	record.Set("display_name", slug)
	record.Set("status", status)
	if err := f.app.Save(record); err != nil {
		t.Fatal(err)
	}
	if status != "active" {
		return record.Id
	}
	dbPath, err := f.dirs.TenantDBPath(record.Id)
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE eng_items (id TEXT PRIMARY KEY NOT NULL, name TEXT, secret TEXT)`); err != nil {
		t.Fatal(err)
	}
	return record.Id
}

func (f *engineFixture) do(t *testing.T, method, path, tenant string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, "/api/arsy/v1"+path, reader)
	req.Header.Set("Content-Type", "application/json")
	if tenant != "" {
		req.Header.Set("X-Arsy-Tenant", tenant)
	}
	rec := httptest.NewRecorder()
	f.mux.ServeHTTP(rec, req)
	return rec
}

func decodeEngineBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body %q: %v", rec.Body.String(), err)
	}
	return body
}

func assertEngineStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d: %s", rec.Code, want, rec.Body.String())
	}
}

func TestEngineRoutesRequireSuperuser(t *testing.T) {
	f := newEngineFixture(t, false)
	f.createTenant(t, "acme", "active")

	assertEngineStatus(t, f.do(t, http.MethodGet, "/engine/models", "", nil), http.StatusUnauthorized)
	assertEngineStatus(t, f.do(t, http.MethodPost, "/engine/"+engineTestModel+"/records", "acme", map[string]any{
		"values":         map[string]any{"id": "r1", "name": "x"},
		"allowed_fields": []string{"id", "name"},
	}), http.StatusUnauthorized)
}

func TestEngineListsModels(t *testing.T) {
	f := newEngineFixture(t, true)

	rec := f.do(t, http.MethodGet, "/engine/models", "", nil)
	assertEngineStatus(t, rec, http.StatusOK)
	if !bytes.Contains(rec.Body.Bytes(), []byte(engineTestModel)) {
		t.Fatalf("models response missing %s: %s", engineTestModel, rec.Body.String())
	}
}

func TestEngineCRUDAndQueryWithinTenant(t *testing.T) {
	f := newEngineFixture(t, true)
	f.createTenant(t, "acme", "active")
	base := "/engine/" + engineTestModel + "/records"

	assertEngineStatus(t, f.do(t, http.MethodPost, base, "acme", map[string]any{
		"values":         map[string]any{"id": "r1", "name": "first"},
		"allowed_fields": []string{"id", "name"},
	}), http.StatusCreated)

	get := f.do(t, http.MethodGet, base+"/r1", "acme", nil)
	assertEngineStatus(t, get, http.StatusOK)

	assertEngineStatus(t, f.do(t, http.MethodPatch, base+"/r1", "acme", map[string]any{
		"values":         map[string]any{"name": "renamed"},
		"allowed_fields": []string{"name"},
	}), http.StatusOK)

	query := f.do(t, http.MethodPost, "/engine/"+engineTestModel+"/query", "acme", map[string]any{
		"allowed_fields": []string{"id", "name"},
		"query":          map[string]any{"limit": 10},
	})
	assertEngineStatus(t, query, http.StatusOK)
	if !bytes.Contains(query.Body.Bytes(), []byte("renamed")) {
		t.Fatalf("query did not return updated row: %s", query.Body.String())
	}

	assertEngineStatus(t, f.do(t, http.MethodDelete, base+"/r1", "acme", nil), http.StatusOK)
	assertEngineStatus(t, f.do(t, http.MethodGet, base+"/r1", "acme", nil), http.StatusNotFound)
}

func TestEngineTenantIsolation(t *testing.T) {
	f := newEngineFixture(t, true)
	f.createTenant(t, "tenant-a", "active")
	f.createTenant(t, "tenant-b", "active")
	base := "/engine/" + engineTestModel + "/records"

	assertEngineStatus(t, f.do(t, http.MethodPost, base, "tenant-a", map[string]any{
		"values":         map[string]any{"id": "secret-row", "name": "only-a"},
		"allowed_fields": []string{"id", "name"},
	}), http.StatusCreated)

	assertEngineStatus(t, f.do(t, http.MethodGet, base+"/secret-row", "tenant-a", nil), http.StatusOK)
	assertEngineStatus(t, f.do(t, http.MethodGet, base+"/secret-row", "tenant-b", nil), http.StatusNotFound)

	query := f.do(t, http.MethodPost, "/engine/"+engineTestModel+"/query", "tenant-b", map[string]any{
		"allowed_fields": []string{"id", "name"},
		"query":          map[string]any{"limit": 10},
	})
	assertEngineStatus(t, query, http.StatusOK)
	if bytes.Contains(query.Body.Bytes(), []byte("only-a")) {
		t.Fatalf("tenant-b query leaked tenant-a row: %s", query.Body.String())
	}
}

func TestEngineRejectsInvalidRequests(t *testing.T) {
	f := newEngineFixture(t, true)
	f.createTenant(t, "acme", "active")
	base := "/engine/" + engineTestModel + "/records"

	assertEngineStatus(t, f.do(t, http.MethodPost, base, "", map[string]any{
		"values":         map[string]any{"id": "r1"},
		"allowed_fields": []string{"id"},
	}), http.StatusBadRequest)

	assertEngineStatus(t, f.do(t, http.MethodPost, base, "acme", map[string]any{
		"values": map[string]any{"id": "r1", "name": "x"},
	}), http.StatusBadRequest)

	assertEngineStatus(t, f.do(t, http.MethodPost, base, "acme", map[string]any{
		"values":         map[string]any{"id": "r1", "secret": "x"},
		"allowed_fields": []string{"id", "name"},
	}), http.StatusBadRequest)

	assertEngineStatus(t, f.do(t, http.MethodPost, "/engine/unknown.model/records", "acme", map[string]any{
		"values":         map[string]any{"id": "r1"},
		"allowed_fields": []string{"id"},
	}), http.StatusNotFound)

	assertEngineStatus(t, f.do(t, http.MethodPost, "/engine/"+engineTestModel+"/query", "acme", map[string]any{
		"query": map[string]any{"limit": 10},
	}), http.StatusBadRequest)
}

func TestEngineTenantStatusMapping(t *testing.T) {
	f := newEngineFixture(t, true)
	f.createTenant(t, "paused", "suspended")
	f.createTenant(t, "closing", "deleting")
	base := "/engine/" + engineTestModel + "/records"
	body := map[string]any{
		"values":         map[string]any{"id": "r1", "name": "x"},
		"allowed_fields": []string{"id", "name"},
	}

	assertEngineStatus(t, f.do(t, http.MethodPost, base, "paused", body), http.StatusConflict)
	assertEngineStatus(t, f.do(t, http.MethodPost, base, "closing", body), http.StatusConflict)
	assertEngineStatus(t, f.do(t, http.MethodPost, base, "missing-tenant", body), http.StatusNotFound)
}
