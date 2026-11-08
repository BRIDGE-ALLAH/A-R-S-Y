package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/engine/bootstrap"
	"github.com/bridge-allah/a-r-s-y/arsy/httpapi"
	"github.com/bridge-allah/a-r-s-y/arsy/tenancy"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
	"github.com/bridge-allah/a-r-s-y/tools/router"
)

func createTenantRecord(t *testing.T, app *tests.TestApp, slug, status string) *core.Record {
	t.Helper()
	collection, err := app.FindCollectionByNameOrId("baas_tenants")
	if err != nil {
		t.Fatal(err)
	}
	record := core.NewRecord(collection)
	record.Set("slug", slug)
	record.Set("display_name", slug)
	record.Set("status", status)
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
	return record
}

func createMembershipRecord(t *testing.T, app *tests.TestApp, tenantID, principal, role, status string) {
	t.Helper()
	collection, err := app.FindCollectionByNameOrId("baas_tenant_memberships")
	if err != nil {
		t.Fatal(err)
	}
	record := core.NewRecord(collection)
	record.Set("tenant", tenantID)
	record.Set("principal", principal)
	record.Set("role", role)
	record.Set("status", status)
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
}

func assertNoSensitiveMarkers(t *testing.T, body string) {
	t.Helper()
	for _, bad := range []string{`pb_data`, `.db`, `ARSY_SECRET_KEY`, `credential_ciphertext`, `manifest_ciphertext`, `private_key`, `DSN`, `TenantsRoot`, `\\arsy\\tenants`, `tenancy provision:`, `Z:\`} {
		if strings.Contains(body, bad) {
			t.Fatalf("response leaked sensitive/internal marker %q: %s", bad, body)
		}
	}
}

func newTenantMuxWithAuth(t *testing.T, auth *core.Record) (*tests.TestApp, http.Handler) {
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

func mustSuperuserRecord(t *testing.T, app *tests.TestApp) *core.Record {
	t.Helper()
	col, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		t.Fatal(err)
	}
	return core.NewRecord(col)
}

func TestTenantAdminContractsSuperuser(t *testing.T) {
	app, mux := newTenantMuxWithAuth(t, nil)
	su := mustSuperuserRecord(t, app)
	app, mux = newTenantMuxWithAuth(t, su)

	create := doJSON(t, mux, http.MethodPost, "/tenants", map[string]any{"slug": "acme", "display_name": "Acme"}, nil)
	if create.Code != http.StatusCreated {
		t.Fatalf("create=%d body=%s", create.Code, create.Body.String())
	}
	assertNoSensitiveMarkers(t, create.Body.String())

	get := doJSON(t, mux, http.MethodGet, "/tenants/acme", nil, nil)
	if get.Code != http.StatusOK || !strings.Contains(get.Body.String(), `"slug":"acme"`) {
		t.Fatalf("get=%d body=%s", get.Code, get.Body.String())
	}
	assertNoSensitiveMarkers(t, get.Body.String())

	list := doJSON(t, mux, http.MethodGet, "/tenants?limit=1&offset=0", nil, nil)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"limit":1`) {
		t.Fatalf("list=%d body=%s", list.Code, list.Body.String())
	}

	update := doJSON(t, mux, http.MethodPatch, "/tenants/acme", map[string]any{"display_name": "Acme Updated", "plan": "pro", "status": "suspended"}, nil)
	if update.Code != http.StatusOK || !strings.Contains(update.Body.String(), `"status":"suspended"`) {
		t.Fatalf("update=%d body=%s", update.Code, update.Body.String())
	}

	resume := doJSON(t, mux, http.MethodPatch, "/tenants/acme", map[string]any{"status": "active"}, nil)
	if resume.Code != http.StatusOK {
		t.Fatalf("resume=%d body=%s", resume.Code, resume.Body.String())
	}

	badTransition := doJSON(t, mux, http.MethodPatch, "/tenants/acme", map[string]any{"status": "provisioning"}, nil)
	if badTransition.Code != http.StatusConflict {
		t.Fatalf("bad transition=%d body=%s", badTransition.Code, badTransition.Body.String())
	}
	assertNoSensitiveMarkers(t, badTransition.Body.String())
}

func TestTenantAdminRequiresAuthAndMembership(t *testing.T) {
	app, mux := newTenantMuxWithAuth(t, nil)
	createTenantRecord(t, app, "acme2", "active")
	unauth := doJSON(t, mux, http.MethodGet, "/tenants/acme2", nil, nil)
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("unauth=%d body=%s", unauth.Code, unauth.Body.String())
	}

	// member without admin role cannot update
	member := core.NewRecord(&core.Collection{Name: "users", Type: core.CollectionTypeAuth})
	member.Id = "member000000001"
	app, mux = newTenantMuxWithAuth(t, member)
	tenant := createTenantRecord(t, app, "acme3", "active")
	createMembershipRecord(t, app, tenant.Id, member.Id, string(tenancy.RoleViewer), string(tenancy.MembershipActive))
	forbidden := doJSON(t, mux, http.MethodPatch, "/tenants/acme3", map[string]any{"display_name": "x"}, nil)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("forbidden=%d body=%s", forbidden.Code, forbidden.Body.String())
	}
}

func TestMembershipAdminContracts(t *testing.T) {
	app, mux := newTenantMuxWithAuth(t, nil)
	su := mustSuperuserRecord(t, app)
	app, mux = newTenantMuxWithAuth(t, su)
	createTenantRecord(t, app, "tenant-a", "active")

	add := doJSON(t, mux, http.MethodPost, "/tenants/tenant-a/memberships", map[string]any{"principal": "membera00000001", "role": "viewer"}, nil)
	if add.Code != http.StatusCreated {
		t.Fatalf("add=%d body=%s", add.Code, add.Body.String())
	}
	assertNoSensitiveMarkers(t, add.Body.String())

	list := doJSON(t, mux, http.MethodGet, "/tenants/tenant-a/memberships?limit=10&offset=0", nil, nil)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"items"`) || !strings.Contains(list.Body.String(), `"limit":10`) {
		t.Fatalf("list=%d body=%s", list.Code, list.Body.String())
	}

	upd := doJSON(t, mux, http.MethodPatch, "/tenants/tenant-a/memberships/membera00000001", map[string]any{"role": "developer", "status": "inactive"}, nil)
	if upd.Code != http.StatusOK || !strings.Contains(upd.Body.String(), `"status":"inactive"`) {
		t.Fatalf("upd=%d body=%s", upd.Code, upd.Body.String())
	}

	del := doJSON(t, mux, http.MethodDelete, "/tenants/tenant-a/memberships/membera00000001", nil, nil)
	if del.Code != http.StatusOK {
		t.Fatalf("del=%d body=%s", del.Code, del.Body.String())
	}

	badRole := doJSON(t, mux, http.MethodPost, "/tenants/tenant-a/memberships", map[string]any{"principal": "memberb00000001", "role": "ownerish"}, nil)
	if badRole.Code != http.StatusBadRequest {
		t.Fatalf("bad role=%d body=%s", badRole.Code, badRole.Body.String())
	}

	notFound := doJSON(t, mux, http.MethodDelete, "/tenants/tenant-a/memberships/missing00000001", nil, nil)
	if notFound.Code != http.StatusNotFound {
		t.Fatalf("missing=%d body=%s", notFound.Code, notFound.Body.String())
	}
}

func TestCrossTenantMembershipIsolation(t *testing.T) {
	// member of tenant A must not access tenant B admin endpoints
	member := core.NewRecord(&core.Collection{Name: "users", Type: core.CollectionTypeAuth})
	member.Id = "member0000000aa"
	app, mux := newTenantMuxWithAuth(t, member)
	ta := createTenantRecord(t, app, "a", "active")
	createTenantRecord(t, app, "b", "active")
	createMembershipRecord(t, app, ta.Id, member.Id, string(tenancy.RoleAdmin), string(tenancy.MembershipActive))

	forbidden := doJSON(t, mux, http.MethodGet, "/tenants/b/memberships", nil, nil)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("cross-tenant memberships=%d body=%s", forbidden.Code, forbidden.Body.String())
	}
	assertNoSensitiveMarkers(t, forbidden.Body.String())

	forbidden2 := doJSON(t, mux, http.MethodPatch, "/tenants/b", map[string]any{"display_name": "x"}, nil)
	if forbidden2.Code != http.StatusForbidden {
		t.Fatalf("cross-tenant update=%d body=%s", forbidden2.Code, forbidden2.Body.String())
	}
}

var _ = httptest.NewRecorder
var _ = bytes.Buffer{}
var _ = json.Marshal
