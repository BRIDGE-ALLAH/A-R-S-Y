package httpapi

import (
	"io/fs"
	"net/http/httptest"
	"testing"

	"github.com/bridge-allah/a-r-s-y/apis"
	"github.com/bridge-allah/a-r-s-y/arsy/dashboard"
	"github.com/bridge-allah/a-r-s-y/core"
)

func TestStaticAtServesEmbeddedAsset(t *testing.T) {
	assets, err := fs.ReadDir(dashboard.DistFS, "dist/assets")
	if err != nil {
		t.Fatal(err)
	}
	var script string
	for _, asset := range assets {
		if len(asset.Name()) > 3 && asset.Name()[len(asset.Name())-3:] == ".js" {
			script = asset.Name()
			break
		}
	}
	if script == "" {
		t.Fatal("built console JavaScript asset is missing")
	}
	consoleFS := apis.MustSubFS(dashboard.DistFS, "dist/assets")
	req := httptest.NewRequest("GET", "/arsy/assets/"+script, nil)
	recorder := httptest.NewRecorder()
	event := new(core.RequestEvent)
	event.Request = req
	event.Response = recorder
	if err := staticAt(consoleFS, "/arsy/assets", false)(event); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != 200 {
		t.Fatalf("unexpected response status: %d", recorder.Code)
	}
}
