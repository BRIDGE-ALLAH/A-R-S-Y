package httpapi_test

import (
	"net/http"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/httpapi"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestHealthRoute(t *testing.T) {
	scenario := tests.ApiScenario{
		Name:           "public A-R-S-Y liveness response",
		Method:         http.MethodGet,
		URL:            "/api/arsy/v1/health",
		ExpectedStatus: http.StatusOK,
		ExpectedContent: []string{
			`"product":"A-R-S-Y"`,
			`"version":"test-version"`,
			`"status":"ok"`,
		},
		NotExpectedContent: []string{`"data"`, `"secret"`},
		ExpectedEvents:     map[string]int{"*": 0},
		TestAppFactory: func(t testing.TB) *tests.TestApp {
			app, err := tests.NewTestApp()
			if err != nil {
				t.Fatal(err)
			}
			httpapi.RegisterRoutes(app, "/api/arsy/v1", "test-version", "ARSY_SECRET_KEY")
			return app
		},
	}
	scenario.Test(t)
}
