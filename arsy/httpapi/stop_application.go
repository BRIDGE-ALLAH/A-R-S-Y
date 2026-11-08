package httpapi

import (
	"errors"
	"net/http"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/deployments"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/router"
)

func stopApplication(cfg config.Config) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		id := e.Request.PathValue("id")
		if id == "" {
			return e.JSON(http.StatusBadRequest, map[string]any{"status": "error", "message": "application ID required"})
		}
		state, err := deployments.StopApplication(e.Request.Context(), e.App, id, cfg.SecretKeyEnv)
		if err != nil {
			var stopErr *deployments.StopApplicationError
			if errors.As(err, &stopErr) {
				return e.JSON(stopApplicationStatus(stopErr.Code), map[string]any{"status": "error", "message": stopErr.Message})
			}
			return e.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": "unable to stop the application"})
		}
		return e.JSON(http.StatusOK, map[string]any{
			"status":          "ok",
			"application_id":  id,
			"lifecycle_state": state,
		})
	}
}

func stopApplicationStatus(code string) int {
	switch code {
	case "not_found":
		return http.StatusNotFound
	case "not_running", "conflict":
		return http.StatusConflict
	case "service_unavailable":
		return http.StatusServiceUnavailable
	case "bad_gateway":
		return http.StatusBadGateway
	case "invalid_request":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func registerStopApplicationRoute(api *router.RouterGroup[*core.RequestEvent], cfg config.Config) {
	api.POST("/applications/{id}/stop", requireSuperuser(stopApplication(cfg)))
}
