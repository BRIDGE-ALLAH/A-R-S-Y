package httpapi

import (
	"github.com/bridge-allah/a-r-s-y/apis"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/router"
)

// ProtectedGroup creates an A-R-S-Y route group restricted to PocketBase superusers.
func ProtectedGroup(parent *router.RouterGroup[*core.RequestEvent], path string) *router.RouterGroup[*core.RequestEvent] {
	return parent.Group(path).Bind(apis.RequireSuperuserAuth())
}
