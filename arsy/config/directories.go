package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Directories describes the private A-R-S-Y runtime directories.
type Directories struct {
	Root      string
	Temp      string
	Manifests string
}

// EnsureDirectories creates the A-R-S-Y directory layout. Private directories
// are created with owner-only permissions.
func EnsureDirectories(root string) (Directories, error) {
	if root == "" {
		return Directories{}, fmt.Errorf("A-R-S-Y data directory is empty")
	}
	dirs := Directories{
		Root: root,
		Temp: filepath.Join(root, "tmp"),
		Manifests: filepath.Join(root, "manifests"),
	}
	for _, dir := range []string{dirs.Root, dirs.Temp, dirs.Manifests} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return Directories{}, fmt.Errorf("create A-R-S-Y directory %q: %w", dir, err)
		}
		if err := os.Chmod(dir, 0o700); err != nil {
			return Directories{}, fmt.Errorf("secure A-R-S-Y directory %q: %w", dir, err)
		}
	}
	return dirs, nil
}
