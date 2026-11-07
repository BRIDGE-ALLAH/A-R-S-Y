package httpapi

import (
	"errors"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	pathpkg "path"
	"strings"

	"github.com/bridge-allah/a-r-s-y/apis"
	"github.com/bridge-allah/a-r-s-y/arsy/dashboard"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tools/router"
)

// RegisterConsole serves the embedded A-R-S-Y single-page application.
func RegisterConsole(app core.App, route string) {
	consoleFS := apis.MustSubFS(dashboard.DistFS, "dist")
	consoleAssetFS := apis.MustSubFS(consoleFS, "assets")
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		e.Router.GET(route, func(re *core.RequestEvent) error {
			return re.Redirect(http.StatusMovedPermanently, route+"/")
		})
		e.Router.GET(route+"/assets/{path...}", staticAt(consoleAssetFS, route+"/assets", false))
		e.Router.GET(route+"/{path...}", staticAt(consoleFS, route, true))
		return e.Next()
	})
}

func staticAt(fsys fs.FS, route string, fallback bool) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		path := strings.TrimPrefix(e.Request.URL.Path, strings.TrimSuffix(route, "/")+"/")
		path, err := url.PathUnescape(path)
		if err != nil {
			return errors.New("invalid console asset path")
		}
		path = pathpkg.Clean(path)
		if path == "." {
			path = ""
		}
		e.Request.SetPathValue(apis.StaticWildcardParam, path)
		return serveConsoleAsset(e, fsys, path, fallback)
	}
}

func serveConsoleAsset(e *core.RequestEvent, fsys fs.FS, filename string, fallback bool) error {
	if filename == "" || filename == "." {
		filename = "index.html"
	}
	if !fs.ValidPath(filename) {
		return router.ErrFileNotFound
	}
	content, err := fs.ReadFile(fsys, filename)
	if err != nil && fallback {
		filename = "index.html"
		content, err = fs.ReadFile(fsys, filename)
	}
	if err != nil {
		return router.ErrFileNotFound
	}
	contentType := mime.TypeByExtension(pathpkg.Ext(filename))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return e.Blob(http.StatusOK, contentType, content)
}
