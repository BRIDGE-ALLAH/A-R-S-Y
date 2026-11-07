package httpapi_test

import (
	"net/http"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/httpapi"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestProtectedGroupRequiresPocketBaseSuperuser(t *testing.T) {
	scenarios := []struct {
		name     string
		token    string
		status   int
		response string
	}{
		{name: "guest", status: http.StatusUnauthorized, response: `"status":401`},
		{
			name:     "regular user",
			token:    "eyJhbGciOiJIUzI1NiJ9.eyJpZCI6IjRxMXhsY2xtZmxva3UzMyIsInR5cGUiOiJhdXRoIiwiY29sbGVjdGlvbklkIjoiX3BiX3VzZXJzX2F1dGhfIiwiZXhwIjoyNTI0NjA0NDYxLCJyZWZyZXNoYWJsZSI6dHJ1ZX0.ZT3F0Z3iM-xbGgSG3LEKiEzHrPHr8t8IuHLZGGNuxLo",
			status:   http.StatusForbidden,
			response: `"status":403`,
		},
		{
			name:     "superuser",
			token:    "eyJhbGciOiJIUzI1NiJ9.eyJpZCI6InN5d2JoZWNuaDQ2cmhtMCIsInR5cGUiOiJhdXRoIiwiY29sbGVjdGlvbklkIjoicGJjXzMxNDI2MzU4MjMiLCJleHAiOjI1MjQ2MDQ0NjEsInJlZnJlc2hhYmxlIjp0cnVlfQ.UXgO3j-0BumcugrFjbd7j0M4MQvbrLggLlcu_YNGjoY",
			status:   http.StatusOK,
			response: "protected",
		},
	}

	for _, tc := range scenarios {
		for _, path := range []string{"/servers", "/deployments"} {
			t.Run(tc.name+path, func(t *testing.T) {
				scenario := tests.ApiScenario{
					Name:           tc.name + path,
					Method:         http.MethodGet,
					URL:            "/api/arsy/v1" + path,
					ExpectedStatus: tc.status,
					Headers:        map[string]string{},
					ExpectedEvents: map[string]int{"*": 0},
					TestAppFactory: func(t testing.TB) *tests.TestApp {
						app, err := tests.NewTestApp()
						if err != nil {
							t.Fatal(err)
						}
						return app
					},
					BeforeTestFunc: func(_ testing.TB, _ *tests.TestApp, e *core.ServeEvent) {
						group := httpapi.ProtectedGroup(e.Router.Group("/api/arsy/v1"), path)
						group.GET("", func(re *core.RequestEvent) error {
							return re.String(http.StatusOK, "protected")
						})
					},
				}
				if tc.token != "" {
					scenario.Headers["Authorization"] = tc.token
				}
				if tc.response != "" {
					scenario.ExpectedContent = []string{tc.response}
				}
				scenario.Test(t)
			})
		}
	}
}
