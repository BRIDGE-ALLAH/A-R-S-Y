package arsy

import (
	"fmt"
	"os"
	"path/filepath"
)

// DataLayout contains the stable A-R-S-Y paths beneath the PocketBase data directory.
type DataLayout struct {
	AppDataDir             string
	EncryptedTemporaryDir  string
	DeploymentManifestsDir string
}

// EnsureDataLayout creates the A-R-S-Y directories without moving PocketBase data.
func EnsureDataLayout(dataDir string) (DataLayout, error) {
	if dataDir == "" {
		return DataLayout{}, fmt.Errorf("A-R-S-Y data directory must not be empty")
	}

	arsyDir := filepath.Join(dataDir, "arsy")
	layout := DataLayout{
		AppDataDir:             filepath.Join(arsyDir, "app"),
		EncryptedTemporaryDir:  filepath.Join(arsyDir, "tmp", "encrypted"),
		DeploymentManifestsDir: filepath.Join(arsyDir, "deployments", "manifests"),
	}
	for _, path := range []string{layout.AppDataDir, layout.EncryptedTemporaryDir, layout.DeploymentManifestsDir} {
		if err := os.MkdirAll(path, 0o700); err != nil {
			return DataLayout{}, fmt.Errorf("create A-R-S-Y data directory: %w", err)
		}
		if err := os.Chmod(path, 0o700); err != nil {
			return DataLayout{}, fmt.Errorf("restrict A-R-S-Y data directory permissions: %w", err)
		}
	}

	return layout, nil
}
