package httpapi

import (
	"net/http"
	"testing"

	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestOnboardingStatusRoute(t *testing.T) {
	scenarios := []struct {
		name             string
		freshInstall     bool
		configuredSecret bool
		expected         []string
	}{
		{
			name:         "fresh install",
			freshInstall: true,
			expected: []string{
				`"superuser_exists":false`,
				`"secret_key_configured":false`,
				`"server_registered":false`,
			},
		},
		{
			name:             "configured control plane",
			configuredSecret: true,
			expected: []string{
				`"superuser_exists":true`,
				`"secret_key_configured":true`,
				`"server_registered":true`,
			},
		},
	}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			scenario := tests.ApiScenario{
				Name:               tc.name,
				Method:             http.MethodGet,
				URL:                "/api/arsy/v1/onboarding",
				ExpectedStatus:     http.StatusOK,
				ExpectedContent:    tc.expected,
				NotExpectedContent: []string{"sensitive-test-key", "arsy-test@example.invalid", "private_key"},
				TestAppFactory: func(t testing.TB) *tests.TestApp {
					app, err := tests.NewTestApp()
					if err != nil {
						t.Fatal(err)
					}
					if tc.freshInstall {
						if _, err := app.DB().NewQuery("DELETE FROM _superusers WHERE email != {:installer}").Bind(map[string]any{"installer": core.DefaultInstallerEmail}).Execute(); err != nil {
							t.Fatal(err)
						}
					}
					if tc.configuredSecret {
						servers := core.NewBaseCollection("servers")
						if err := app.Save(servers); err != nil {
							t.Fatal(err)
						}
						if err := app.Save(core.NewRecord(servers)); err != nil {
							t.Fatal(err)
						}
					}
					secretLookup := func(string) (string, bool) {
						if tc.configuredSecret {
							return "sensitive-test-key", true
						}
						return "", false
					}
					registerRoutes(app, "/api/arsy/v1", "test-version", "ARSY_SECRET_KEY", secretLookup, nil)
					return app
				},
			}
			scenario.Test(t)
		})
	}
}
