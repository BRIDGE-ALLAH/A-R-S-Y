package httpapi

import (
	"net/http"

	"github.com/bridge-allah/a-r-s-y/core"
)

// requireSuperuser wraps an infrastructure handler with PocketBase's existing
// superuser authentication check. It deliberately does not introduce a
// separate A-R-S-Y credential system.
func requireSuperuser(next func(*core.RequestEvent) error) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if e.Auth == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"status": "unauthorized"})
		}
		if !e.HasSuperuserAuth() {
			return e.JSON(http.StatusForbidden, map[string]any{"status": "forbidden"})
		}
		return next(e)
	}
}
