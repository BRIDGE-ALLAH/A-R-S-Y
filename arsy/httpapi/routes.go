package httpapi

import (
	"net/http"
	"os"
	"strings"

	"github.com/bridge-allah/a-r-s-y/core"
)

// RegisterRoutes registers the A-R-S-Y API endpoints under one versioned route group.
func RegisterRoutes(app core.App, apiPrefix, version, secretKeyEnv string) {
	registerRoutes(app, apiPrefix, version, secretKeyEnv, os.LookupEnv, func() error {
		_, err := app.DB().NewQuery("SELECT 1").Execute()
		return err
	})
}

func registerRoutes(app core.App, apiPrefix, version, secretKeyEnv string, secretLookup func(string) (string, bool), readinessCheck func() error) {
	prefix := strings.TrimRight(apiPrefix, "/")
	if secretLookup == nil {
		secretLookup = os.LookupEnv
	}
	if readinessCheck == nil {
		readinessCheck = func() error {
			_, err := app.DB().NewQuery("SELECT 1").Execute()
			return err
		}
	}
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		group := e.Router.Group(prefix)
		group.GET("/health", func(re *core.RequestEvent) error {
			return re.JSON(http.StatusOK, struct {
				Product string `json:"product"`
				Version string `json:"version"`
				Status  string `json:"status"`
			}{Product: "A-R-S-Y", Version: version, Status: "ok"})
		})
		group.GET("/ready", func(re *core.RequestEvent) error {
			status := http.StatusOK
			result := "ready"
			if readinessCheck() != nil {
				status = http.StatusServiceUnavailable
				result = "not_ready"
			}
			return re.JSON(status, struct {
				Status string `json:"status"`
			}{Status: result})
		})
		group.GET("/onboarding", func(re *core.RequestEvent) error {
			status, err := readOnboardingStatus(app, secretKeyEnv, secretLookup)
			if err != nil {
				return re.InternalServerError("Failed to read onboarding status", err)
			}
			return re.JSON(http.StatusOK, status)
		})
		return e.Next()
	})
}
