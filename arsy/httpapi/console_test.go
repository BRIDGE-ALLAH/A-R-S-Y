package httpapi_test

import (
	"io/fs"
	"net/http"
	"testing"

	"github.com/bridge-allah/a-r-s-y/apis"
	"github.com/bridge-allah/a-r-s-y/arsy/dashboard"
	"github.com/bridge-allah/a-r-s-y/arsy/httpapi"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
	"github.com/bridge-allah/a-r-s-y/ui"
)

func TestConsoleRouteAndPocketBaseAdminRemainAvailable(t *testing.T) {
	for _, tc := range []struct {
		name   string
		path   string
		status int
		body   string
	}{
		{name: "console entry", path: "/arsy/", status: http.StatusOK, body: "/arsy/assets/"},
		{name: "console client route fallback", path: "/arsy/projects/demo", status: http.StatusOK, body: "/arsy/assets/"},
		{name: "missing console asset", path: "/arsy/assets/missing.js", status: http.StatusNotFound, body: "File not found"},
		{name: "PocketBase superuser dashboard canonical redirect", path: "/_", status: http.StatusTemporaryRedirect, body: "Temporary Redirect"},
		{name: "PocketBase API", path: "/api/health", status: http.StatusOK, body: "API is healthy"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scenario := tests.ApiScenario{
				Name:           tc.name,
				Method:         http.MethodGet,
				URL:            tc.path,
				ExpectedStatus: tc.status,
				ExpectedContent: func() []string {
					if tc.body == "" {
						return nil
					}
					return []string{tc.body}
				}(),
				ExpectedEvents: map[string]int{"*": 0},
				BeforeTestFunc: func(_ testing.TB, _ *tests.TestApp, e *core.ServeEvent) {
					if ui.DistDirFS != nil {
						e.Router.GET("/_/{path...}", apis.Static(ui.DistDirFS, false))
					}
				},
				TestAppFactory: func(t testing.TB) *tests.TestApp {
					app, err := tests.NewTestApp()
					if err != nil {
						t.Fatal(err)
					}
					httpapi.RegisterConsole(app, "/arsy")
					return app
				},
			}
			scenario.Test(t)
		})
	}
}

func TestConsoleServesEmbeddedAssets(t *testing.T) {
	assets, err := fs.ReadDir(dashboard.DistFS, "dist/assets")
	if err != nil {
		t.Fatal(err)
	}
	var script string
	for _, asset := range assets {
		if len(asset.Name()) > 3 && asset.Name()[len(asset.Name())-3:] == ".js" {
			script = asset.Name()
			break
		}
	}
	if script == "" {
		t.Fatal("built console JavaScript asset is missing")
	}
	scenario := tests.ApiScenario{
		Name:               "serve embedded console JavaScript",
		Method:             http.MethodGet,
		URL:                "/arsy/assets/" + script,
		ExpectedStatus:     http.StatusOK,
		NotExpectedContent: []string{"File not found"},
		ExpectedEvents:     map[string]int{"*": 0},
		TestAppFactory: func(t testing.TB) *tests.TestApp {
			app, err := tests.NewTestApp()
			if err != nil {
				t.Fatal(err)
			}
			httpapi.RegisterConsole(app, "/arsy")
			return app
		},
	}
	scenario.Test(t)
}
