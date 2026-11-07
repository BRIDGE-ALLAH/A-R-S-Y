package httpapi

import (
	"errors"
	"net/http"
	"testing"

	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestReadinessRoute(t *testing.T) {
	scenarios := []struct {
		name           string
		check          func() error
		status         int
		responseStatus string
	}{
		{name: "ready", status: http.StatusOK, responseStatus: "ready"},
		{name: "not ready", check: func() error { return errors.New("database unavailable") }, status: http.StatusServiceUnavailable, responseStatus: "not_ready"},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			apiScenario := tests.ApiScenario{
				Name:           scenario.name,
				Method:         http.MethodGet,
				URL:            "/api/arsy/v1/ready",
				ExpectedStatus: scenario.status,
				ExpectedContent: []string{
					`"status":"` + scenario.responseStatus + `"`,
				},
				NotExpectedContent: []string{"database unavailable", "secret", "path"},
				ExpectedEvents:     map[string]int{"*": 0},
				TestAppFactory: func(t testing.TB) *tests.TestApp {
					app, err := tests.NewTestApp()
					if err != nil {
						t.Fatal(err)
					}
					registerRoutes(app, "/api/arsy/v1", "test-version", "ARSY_SECRET_KEY", nil, scenario.check)
					return app
				},
			}
			apiScenario.Test(t)
		})
	}
}
