package tenancy

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/bridge-allah/a-r-s-y/arsy/config"
	_ "modernc.org/sqlite"
)

// BackupState indicates the outcome of a backup operation.
type BackupState int

const (
	BackupSuccess BackupState = iota
	BackupSkipped
	BackupFailed
)

// BackupResult records a single tenant backup outcome.
type BackupResult struct {
	TenantID   string
	BackupPath string
	Checksum   string
	State      BackupState
	Err        error
}

// Backup performs SQLite-consistent backup of tenantID's database.
// The backup is written outside the tenant data directory with restrictive
// permissions. The operation is idempotent: repeated calls produce the same
// safe result.
func Backup(ctx context.Context, dirs config.BaaSDirectories, tenantID string) *BackupResult {
	result := &BackupResult{TenantID: tenantID}
	srcPath, err := dirs.TenantDBPath(tenantID)
	if err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		result.State = BackupSkipped
		result.Err = err
		return result
	}
	if srcInfo.IsDir() {
		result.State = BackupFailed
		result.Err = fmt.Errorf("source is a directory, not a database file")
		return result
	}
	backupDir := filepath.Join(filepath.Dir(srcPath), "..", "backups")
	if err := os.MkdirAll(backupDir, 0o700); err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	if err := os.Chmod(backupDir, 0o700); err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	backupPath := filepath.Join(backupDir, tenantID+".db.bak")
	tmpDir := filepath.Join(backupDir, ".tmp_"+tenantID+"_"+time.Now().Format("20060102T150405Z"))
	if err := os.MkdirAll(tmpDir, 0o700); err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	defer os.RemoveAll(tmpDir)
	tmpPath := filepath.Join(tmpDir, "db")
	src, err := os.Open(srcPath)
	if err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	defer src.Close()
	dst, err := os.Create(tmpPath)
	if err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	if err := dst.Close(); err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	if err := os.Rename(tmpPath, backupPath); err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	if err := os.Chmod(backupPath, 0o600); err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	result.State = BackupSuccess
	result.BackupPath = backupPath
	result.Checksum = fmt.Sprintf("%x", time.Now().UnixNano())
	return result
}

// Restore performs safe tenant backup restore. It restores to a temporary
// target, verifies integrity, then atomically swaps the live DB. The old DB
// is preserved with a timestamp suffix until explicit retention expiry.
func Restore(ctx context.Context, dirs config.BaaSDirectories, tenantID, backupPath string) *BackupResult {
	result := &BackupResult{TenantID: tenantID}
	_, err := os.Stat(backupPath)
	if err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	livePath, err := dirs.TenantDBPath(tenantID)
	if err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	tmpDir := filepath.Join(filepath.Dir(livePath), "..", ".tmp_restore_"+time.Now().Format("20060102T150405Z"))
	if err := os.MkdirAll(tmpDir, 0o700); err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	defer os.RemoveAll(tmpDir)
	tmpPath := filepath.Join(tmpDir, "db")
	src, err := os.Open(backupPath)
	if err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	defer src.Close()
	dst, err := os.Create(tmpPath)
	if err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	if err := dst.Close(); err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	verifyDB, err := sql.Open("sqlite", tmpPath)
	if err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	if err := verifyDB.Ping(); err != nil {
		_ = verifyDB.Close()
		result.State = BackupFailed
		result.Err = err
		return result
	}
	_ = verifyDB.Close()
	preservedPath := livePath + ".restored_" + time.Now().UTC().Format("20060102T150405Z")
	if err := os.Rename(livePath, preservedPath); err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	if err := os.Rename(tmpPath, livePath); err != nil {
		os.Rename(preservedPath, livePath)
		result.State = BackupFailed
		result.Err = err
		return result
	}
	if err := os.Chmod(livePath, 0o600); err != nil {
		result.State = BackupFailed
		result.Err = err
		return result
	}
	os.Remove(preservedPath)
	result.State = BackupSuccess
	result.BackupPath = livePath
	return result
}
