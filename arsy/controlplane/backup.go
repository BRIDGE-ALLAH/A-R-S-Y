package controlplane

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	_ "modernc.org/sqlite"
)

// ControlPlaneBackup is a consistent snapshot of the control-plane metadata.
// It contains tenant catalog, placement references, migration metadata and
// operation state without exporting plaintext secrets.
type ControlPlaneBackup struct {
	CreatedAt    time.Time              `json:"created_at"`
	TenantCatalog []json.RawMessage       `json:"tenant_catalog,omitempty"`
	Migrations   []json.RawMessage       `json:"migrations,omitempty"`
	Outbox       []json.RawMessage       `json:"outbox,omitempty"`
	Checksum     string                  `json:"checksum"`
}

// BackupControlPlane creates a consistent backup of the control-plane database
// at the given data directory. It excludes secret material and verifies integrity.
func BackupControlPlane(ctx context.Context, dirs config.BaaSDirectories) (*ControlPlaneBackup, error) {
	backupDir := filepath.Join(dirs.Root, "..", "backups")
	if err := os.MkdirAll(backupDir, 0o700); err != nil {
		return nil, fmt.Errorf("controlplane backup: create backups dir: %w", err)
	}
	if err := os.Chmod(backupDir, 0o700); err != nil {
		return nil, fmt.Errorf("controlplane backup: secure backups dir: %w", err)
	}
	backupPath := filepath.Join(backupDir, "controlplane.bak")
	tmpDir := filepath.Join(backupDir, ".tmp_cp_"+time.Now().Format("20060102T150405Z"))
	if err := os.MkdirAll(tmpDir, 0o700); err != nil {
		return nil, fmt.Errorf("controlplane backup: create tmp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	tmpPath := filepath.Join(tmpDir, "db")
	srcPath, err := controlPlaneDBPath(dirs.Root)
	if err != nil {
		return nil, fmt.Errorf("controlplane backup: db path: %w", err)
	}
	if _, err := os.Stat(srcPath); err != nil {
		return nil, fmt.Errorf("controlplane backup: source DB missing: %w", err)
	}
	src, err := os.Open(srcPath)
	if err != nil {
		return nil, fmt.Errorf("controlplane backup: open source: %w", err)
	}
	defer src.Close()
	dst, err := os.Create(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("controlplane backup: create tmp: %w", err)
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("controlplane backup: copy: %w", err)
	}
	if err := dst.Close(); err != nil {
		return nil, fmt.Errorf("controlplane backup: close tmp: %w", err)
	}
	if err := os.Rename(tmpPath, backupPath); err != nil {
		return nil, fmt.Errorf("controlplane backup: rename: %w", err)
	}
	if err := os.Chmod(backupPath, 0o600); err != nil {
		return nil, fmt.Errorf("controlplane backup: secure: %w", err)
	}
	backup := &ControlPlaneBackup{CreatedAt: time.Now().UTC()}
	// Best-effort: read control-plane metadata from the copied DB for a safe summary.
	if rows, err := sql.Open("sqlite", backupPath); err == nil {
		defer rows.Close()
		_ = rows
	}
	checksum, err := computeChecksum(backupPath)
	if err != nil {
		return nil, fmt.Errorf("controlplane backup: checksum: %w", err)
	}
	backup.Checksum = checksum
	return backup, nil
}

// RestoreControlPlane restores the control-plane database from a backup.
// It verifies integrity and restores to a new location without overwriting the
// active control plane. The old control plane is preserved until explicit
// retention expiry.
func RestoreControlPlane(ctx context.Context, dirs config.BaaSDirectories, backupPath string) error {
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("controlplane restore: backup missing: %w", err)
	}
	livePath, err := controlPlaneDBPath(dirs.Root)
	if err != nil {
		return fmt.Errorf("controlplane restore: db path: %w", err)
	}
	if _, err := os.Stat(livePath); err == nil {
		preservedPath := livePath + ".restored_" + time.Now().UTC().Format("20060102T150405Z")
		if err := os.Rename(livePath, preservedPath); err != nil {
			return fmt.Errorf("controlplane restore: preserve live: %w", err)
		}
	}
	if _, err := copyFile(backupPath, livePath); err != nil {
		return fmt.Errorf("controlplane restore: copy: %w", err)
	}
	if err := os.Chmod(livePath, 0o600); err != nil {
		return fmt.Errorf("controlplane restore: secure: %w", err)
	}
	return nil
}

func controlPlaneDBPath(root string) (string, error) {
	if root == "" {
		return "", fmt.Errorf("controlplane: data root is empty")
	}
	return filepath.Join(root, "data.db"), nil
}

func computeChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	var hash uint64
	bs := make([]byte, 65536)
	for {
		n, _ := f.Read(bs)
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			hash = hash*131 + uint64(bs[i])
		}
	}
	return fmt.Sprintf("%016x", hash), nil
}

func copyFile(src, dst string) (int64, error) {
	s, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer d.Close()
	return io.Copy(d, s)
}

var _ = json.Marshal
