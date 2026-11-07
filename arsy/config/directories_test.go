package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEnsureDirectories(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permissions not enforced on Windows")
	}
	root := filepath.Join(t.TempDir(), "arsy")
	dirs, err := EnsureDirectories(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{dirs.Root, dirs.Temp, dirs.Manifests} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatal(err)
		}
		if !info.IsDir() {
			t.Fatalf("%s is not a directory", dir)
		}
		if info.Mode().Perm() != 0o700 {
			t.Fatalf("%s permissions = %o, want 700", dir, info.Mode().Perm())
		}
	}
}
