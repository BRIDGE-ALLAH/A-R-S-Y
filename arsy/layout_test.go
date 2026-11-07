package arsy

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEnsureDataLayout(t *testing.T) {
	root := t.TempDir()
	layout, err := EnsureDataLayout(root)
	if err != nil {
		t.Fatalf("EnsureDataLayout() error = %v", err)
	}
	for _, path := range []string{layout.AppDataDir, layout.EncryptedTemporaryDir, layout.DeploymentManifestsDir} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("expected directory %q: %v", path, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %q to be a directory", path)
		}
		if runtime.GOOS != "windows" && info.Mode().Perm() != 0o700 {
			t.Errorf("permissions for %q = %04o, want 0700", path, info.Mode().Perm())
		}
	}
	if filepath.Dir(layout.AppDataDir) != filepath.Join(root, "arsy") {
		t.Fatalf("A-R-S-Y directories escaped their namespace: %#v", layout)
	}
}

func TestEnsureDataLayoutRejectsEmptyPath(t *testing.T) {
	if _, err := EnsureDataLayout(""); err == nil {
		t.Fatal("expected an empty data directory to fail")
	}
}
