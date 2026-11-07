package httpapi

import (
	"net/http"

	"github.com/bridge-allah/a-r-s-y/apis"
	"github.com/bridge-allah/a-r-s-y/arsy/config"
	"github.com/bridge-allah/a-r-s-y/arsy/console"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/hook"
)

// Register attaches the A-R-S-Y API and console to the PocketBase HTTP server
// without changing the existing PocketBase routes.
func Register(app core.App, cfg config.Config, version string) {
	app.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Id: "arsy_http_api",
		Func: func(e *core.ServeEvent) error {
			e.Router.GET(cfg.APIPrefix+"/health", func(re *core.RequestEvent) error {
				return re.JSON(http.StatusOK, map[string]any{
					"product": "A-R-S-Y",
					"version": version,
					"status":  "alive",
				})
			})
			e.Router.GET(cfg.APIPrefix+"/ready", func(re *core.RequestEvent) error {
				if !re.HasSuperuserAuth() {
					return re.JSON(http.StatusUnauthorized, map[string]any{"status": "unauthorized"})
				}

				statusCode, status := readinessStatus(re.App.IsBootstrapped(), func() error {
					_, err := re.App.DB().NewQuery("SELECT 1").Execute()
					return err
				})
				return re.JSON(statusCode, map[string]any{"status": status})
			})

			e.Router.GET(cfg.ConsolePath, func(re *core.RequestEvent) error {
				return re.Redirect(http.StatusPermanentRedirect, cfg.ConsolePath+"/")
			})
			e.Router.GET(cfg.ConsolePath+"/{path...}", apis.Static(console.FS(), true))
			return e.Next()
		},
	})
}
