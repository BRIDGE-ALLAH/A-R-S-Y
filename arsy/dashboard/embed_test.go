package dashboard_test

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/dashboard"
)

func TestDistFSContainsBuiltConsole(t *testing.T) {
	index, err := fs.ReadFile(dashboard.DistFS, "dist/index.html")
	if err != nil {
		t.Fatalf("read embedded console index: %v", err)
	}
	if !strings.Contains(string(index), "/arsy/assets/") {
		t.Fatalf("console index does not reference the /arsy asset base: %s", index)
	}
	if _, err := fs.Stat(dashboard.DistFS, "dist/assets"); err != nil {
		t.Fatalf("built console assets are not embedded: %v", err)
	}
}
