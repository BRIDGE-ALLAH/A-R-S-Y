package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/engine/bootstrap"
	"github.com/bridge-allah/a-r-s-y/arsy/httpapi"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
	"github.com/bridge-allah/a-r-s-y/tools/router"
)

func newTenantAPIFixture(t *testing.T, authenticated bool) (*tests.TestApp, http.Handler) {
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
	reg, err := bootstrap.DefaultRegistry()
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
	r := router.NewRouter[*core.RequestEvent](func(w http.ResponseWriter, req *http.Request) (*core.RequestEvent, router.EventCleanupFunc) {
		event := &core.RequestEvent{App: app, Auth: auth}
		event.Request = req
		event.Response = w
		return event, nil
	})
	cfg := config.Config{DataDir: t.TempDir(), APIPrefix: "/api/arsy/v1", ConsolePath: "/arsy", SecretKeyEnv: "ARSY_SECRET_KEY"}
	httpapi.RegisterWithEngine(app, cfg, "test-version", httpapi.EngineConfig{Registry: reg, TenantsRoot: dirs.TenantsRoot})
	if err := app.OnServe().Trigger(&core.ServeEvent{App: app, Router: r}, func(e *core.ServeEvent) error { return e.Next() }); err != nil {
		t.Fatal(err)
	}
	mux, err := r.BuildMux()
	if err != nil {
		t.Fatal(err)
	}
	return app, mux
}

func doJSON(t *testing.T, mux http.Handler, method, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var raw []byte
	if body != nil {
		var err error
		raw, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, "/api/arsy/v1"+path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestTenantRoutesRequireSuperuser(t *testing.T) {
	_, mux := newTenantAPIFixture(t, false)
	rec := doJSON(t, mux, http.MethodGet, "/tenants", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestTenantCreateListProvisionAndEngineCRUD(t *testing.T) {
	_, mux := newTenantAPIFixture(t, true)

	create := doJSON(t, mux, http.MethodPost, "/tenants", map[string]any{"slug": "acme-co", "display_name": "Acme Co"}, nil)
	if create.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}

	list := doJSON(t, mux, http.MethodGet, "/tenants", nil, nil)
	if list.Code != http.StatusOK || !bytes.Contains(list.Body.Bytes(), []byte("acme-co")) {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}

	provision := doJSON(t, mux, http.MethodPost, "/tenants/acme-co/provision", nil, nil)
	if provision.Code != http.StatusOK {
		t.Fatalf("provision status=%d body=%s", provision.Code, provision.Body.String())
	}

	h := map[string]string{"X-Arsy-Tenant": "acme-co"}
	createRec := doJSON(t, mux, http.MethodPost, "/engine/base.company/records", map[string]any{
		"values":         map[string]any{"id": "c1", "name": "Acme", "code": "ACME"},
		"allowed_fields": []string{"id", "name", "code"},
	}, h)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("engine create status=%d body=%s", createRec.Code, createRec.Body.String())
	}
	get := doJSON(t, mux, http.MethodGet, "/engine/base.company/records/c1", nil, h)
	if get.Code != http.StatusOK || !bytes.Contains(get.Body.Bytes(), []byte("Acme")) {
		t.Fatalf("engine get status=%d body=%s", get.Code, get.Body.String())
	}
}

func TestTenantCreateRejectsInvalidSlug(t *testing.T) {
	_, mux := newTenantAPIFixture(t, true)
	rec := doJSON(t, mux, http.MethodPost, "/tenants", map[string]any{"slug": "Bad Slug", "display_name": "Bad"}, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
