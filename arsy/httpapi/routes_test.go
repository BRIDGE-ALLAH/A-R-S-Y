package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/httpapi"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
	"github.com/bridge-allah/a-r-s-y/tools/router"
)

func newTestMux(t *testing.T) (*tests.TestApp, http.Handler) {
	t.Helper()
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter[*core.RequestEvent](func(w http.ResponseWriter, req *http.Request) (*core.RequestEvent, router.EventCleanupFunc) {
		event := &core.RequestEvent{App: app}
		event.Request = req
		event.Response = w
		return event, nil
	})
	cfg := config.Config{
		DataDir: t.TempDir(),
		APIPrefix: "/api/arsy/v1",
		ConsolePath: "/arsy",
		SecretKeyEnv: "ARSY_SECRET_KEY",
	}
	httpapi.Register(app, cfg, "test-version")

	serveEvent := &core.ServeEvent{App: app, Router: r}
	if err := app.OnServe().Trigger(serveEvent, func(e *core.ServeEvent) error {
		return e.Next()
	}); err != nil {
		app.Cleanup()
		t.Fatal(err)
	}

	mux, err := r.BuildMux()
	if err != nil {
		app.Cleanup()
		t.Fatal(err)
	}
	return app, mux
}

func TestHealthEndpointIsMinimal(t *testing.T) {
	app, mux := newTestMux(t)
	defer app.Cleanup()

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/arsy/v1/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 3 || body["product"] != "A-R-S-Y" || body["version"] != "test-version" || body["status"] != "alive" {
		t.Fatalf("unexpected health response: %#v", body)
	}
}

func TestReadinessEndpointRequiresSuperuser(t *testing.T) {
	app, mux := newTestMux(t)
	defer app.Cleanup()

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/arsy/v1/ready", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestConsoleRouteServesEmbeddedDashboard(t *testing.T) {
	app, mux := newTestMux(t)
	defer app.Cleanup()

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/arsy/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "A-R-S-Y Console") {
		t.Fatalf("console response does not contain the dashboard title: %s", rec.Body.String())
	}
}
