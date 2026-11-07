package deployments

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const maxComposeManifestBytes = 1 << 20

type sftpFile interface {
	Write([]byte) (int, error)
	Chmod(os.FileMode) error
	Sync() error
	Close() error
}

type sftpClient interface {
	MkdirAll(string) error
	OpenFile(string, int) (sftpFile, error)
	Rename(string, string) error
	Remove(string) error
	Close() error
}

type pkgSFTPClient struct {
	client *sftp.Client
}

func (c pkgSFTPClient) MkdirAll(name string) error {
	return c.client.MkdirAll(name)
}

func (c pkgSFTPClient) OpenFile(name string, flags int) (sftpFile, error) {
	return c.client.OpenFile(name, flags)
}

func (c pkgSFTPClient) Rename(oldPath, newPath string) error {
	return c.client.Rename(oldPath, newPath)
}

func (c pkgSFTPClient) Remove(name string) error {
	return c.client.Remove(name)
}

func (c pkgSFTPClient) Close() error {
	return c.client.Close()
}

// UploadComposeManifest validates the MVP-1 Compose subset and atomically
// uploads it over SFTP with owner-only file permissions.
func UploadComposeManifest(ctx context.Context, client *ssh.Client, applicationSlug, deploymentID string, content []byte) error {
	if ctx == nil {
		return errors.New("deployment context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if client == nil {
		return errors.New("SSH client is required")
	}
	remotePath, err := validateComposeUpload(applicationSlug, deploymentID, content)
	if err != nil {
		return err
	}
	sftpConnection, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("start SFTP subsystem: %w", err)
	}
	defer sftpConnection.Close()
	return uploadComposeManifest(ctx, pkgSFTPClient{client: sftpConnection}, remotePath, content)
}

func validateComposeUpload(applicationSlug, deploymentID string, content []byte) (string, error) {
	if len(content) == 0 || len(content) > maxComposeManifestBytes {
		return "", fmt.Errorf("Compose manifest must be between 1 byte and %d bytes", maxComposeManifestBytes)
	}
	if _, err := ValidateCompose(string(content)); err != nil {
		return "", fmt.Errorf("validate Compose manifest: %w", err)
	}
	return ManifestPath(applicationSlug, deploymentID)
}

func uploadComposeManifest(ctx context.Context, client sftpClient, remotePath string, content []byte) error {
	if ctx == nil {
		return errors.New("deployment context is required")
	}
	if client == nil {
		return errors.New("SFTP client is required")
	}
	if err := ctx.Err(); err != nil {
		_ = client.Close()
		return err
	}
	stopCancellation := context.AfterFunc(ctx, func() { _ = client.Close() })
	defer stopCancellation()

	if err := client.MkdirAll(path.Dir(remotePath)); err != nil {
		return fmt.Errorf("create remote release directory: %w", err)
	}
	tempID, err := NewDeploymentID()
	if err != nil {
		return err
	}
	tempPath := remotePath + ".tmp-" + tempID
	file, err := client.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL)
	if err != nil {
		return fmt.Errorf("create remote temporary manifest: %w", err)
	}
	renamed := false
	defer func() {
		if !renamed {
			_ = client.Remove(tempPath)
		}
	}()

	if err := file.Chmod(0600); err != nil {
		_ = file.Close()
		return fmt.Errorf("secure remote manifest permissions: %w", err)
	}
	for offset := 0; offset < len(content); {
		if err := ctx.Err(); err != nil {
			_ = client.Close()
			_ = file.Close()
			return err
		}
		end := offset + 32*1024
		if end > len(content) {
			end = len(content)
		}
		n, writeErr := file.Write(content[offset:end])
		if writeErr != nil {
			_ = file.Close()
			return fmt.Errorf("write remote manifest: %w", writeErr)
		}
		if n == 0 {
			_ = file.Close()
			return io.ErrShortWrite
		}
		offset += n
	}
	if err := ctx.Err(); err != nil {
		_ = client.Close()
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync remote manifest: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close remote manifest: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := client.Rename(tempPath, remotePath); err != nil {
		return fmt.Errorf("publish remote manifest: %w", err)
	}
	renamed = true
	return nil
}

// UploadComposeManifestWithTimeout applies a bounded SFTP transfer window.
func UploadComposeManifestWithTimeout(ctx context.Context, client *ssh.Client, applicationSlug, deploymentID string, content []byte, timeout time.Duration) error {
	if ctx == nil {
		return errors.New("deployment context is required")
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	transferCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return UploadComposeManifest(transferCtx, client, applicationSlug, deploymentID, content)
}
