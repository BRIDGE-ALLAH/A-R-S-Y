package httpapi

import (
	"net/http"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/hook"
)

// Register attaches the A-R-S-Y health and readiness endpoints to the
// PocketBase HTTP server without changing the existing PocketBase routes.
func Register(app core.App, cfg config.Config, version string) {
	app.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Id: "arsy_http_api",
		Func: func(e *core.ServeEvent) error {
			e.Router.GET(cfg.APIPrefix+"/health", func(re *core.RequestEvent) error {
				return re.JSON(http.StatusOK, map[string]any{
					"product": "A-R-S-Y",
					"version": version,
					"status": "alive",
				})
			})
			e.Router.GET(cfg.APIPrefix+"/ready", func(re *core.RequestEvent) error {
				if !re.HasSuperuserAuth() {
					return re.JSON(http.StatusUnauthorized, map[string]any{"status": "unauthorized"})
				}
				if !re.App.IsBootstrapped() {
					return re.JSON(http.StatusServiceUnavailable, map[string]any{"status": "not_ready"})
				}
				return re.JSON(http.StatusOK, map[string]any{"status": "ready"})
			})
			return e.Next()
		},
	})
}
