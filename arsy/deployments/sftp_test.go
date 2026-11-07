package deployments

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path"
	"strings"
	"testing"
)

type fakeSFTPClient struct {
	directories []string
	files       map[string]*fakeSFTPFile
	renamed     map[string]string
	removeCalls []string
	mkdirErr    error
	openErr     error
	renameErr   error
	closed      bool
	onWrite     func()
}

func newFakeSFTPClient() *fakeSFTPClient {
	return &fakeSFTPClient{files: make(map[string]*fakeSFTPFile), renamed: make(map[string]string)}
}

func (c *fakeSFTPClient) MkdirAll(name string) error {
	if c.mkdirErr != nil {
		return c.mkdirErr
	}
	c.directories = append(c.directories, name)
	return nil
}

func (c *fakeSFTPClient) OpenFile(name string, _ int) (sftpFile, error) {
	if c.openErr != nil {
		return nil, c.openErr
	}
	file := &fakeSFTPFile{onWrite: c.onWrite}
	c.files[name] = file
	return file, nil
}

func (c *fakeSFTPClient) Rename(oldPath, newPath string) error {
	if c.renameErr != nil {
		return c.renameErr
	}
	file, ok := c.files[oldPath]
	if !ok {
		return errors.New("temporary file missing")
	}
	c.files[newPath] = file
	delete(c.files, oldPath)
	c.renamed[oldPath] = newPath
	return nil
}

func (c *fakeSFTPClient) Remove(name string) error {
	c.removeCalls = append(c.removeCalls, name)
	delete(c.files, name)
	return nil
}

func (c *fakeSFTPClient) Close() error {
	c.closed = true
	return nil
}

type fakeSFTPFile struct {
	buffer  bytes.Buffer
	mode    os.FileMode
	synced  bool
	closed  bool
	onWrite func()
}

func (f *fakeSFTPFile) Write(value []byte) (int, error) {
	n, err := f.buffer.Write(value)
	if f.onWrite != nil {
		f.onWrite()
		f.onWrite = nil
	}
	return n, err
}

func (f *fakeSFTPFile) Chmod(mode os.FileMode) error {
	f.mode = mode
	return nil
}

func (f *fakeSFTPFile) Sync() error {
	f.synced = true
	return nil
}

func (f *fakeSFTPFile) Close() error {
	f.closed = true
	return nil
}

func TestUploadComposeManifestWritesSecurelyAndRenamesAtomically(t *testing.T) {
	ctx := context.Background()
	client := newFakeSFTPClient()
	manifest := []byte("services:\n  web:\n    image: nginx:1.27\n")
	if err := uploadComposeManifest(ctx, client, "/opt/arsy/apps/demo/releases/0123456789abcdef0123456789abcdef/compose.yaml", manifest); err != nil {
		t.Fatal(err)
	}
	if len(client.directories) != 1 || client.directories[0] != "/opt/arsy/apps/demo/releases/0123456789abcdef0123456789abcdef" {
		t.Fatalf("created directories = %#v", client.directories)
	}
	if len(client.renamed) != 1 || len(client.removeCalls) != 0 {
		t.Fatalf("rename operations = %#v, cleanup = %#v", client.renamed, client.removeCalls)
	}
	for finalPath, file := range client.files {
		if finalPath != "/opt/arsy/apps/demo/releases/0123456789abcdef0123456789abcdef/compose.yaml" {
			t.Fatalf("unexpected final path: %q", finalPath)
		}
		if file.mode != 0600 || !file.synced || !file.closed || file.buffer.String() != string(manifest) {
			t.Fatalf("uploaded file state = %#v", file)
		}
	}
}

func TestValidateComposeUploadRejectsInvalidManifestAndPath(t *testing.T) {
	valid := []byte("services:\n  web:\n    image: nginx\n")
	if path, err := validateComposeUpload("demo", "0123456789abcdef0123456789abcdef", valid); err != nil || !strings.HasSuffix(path, "/compose.yaml") {
		t.Fatalf("valid Compose upload path = %q, %v", path, err)
	}
	if _, err := validateComposeUpload("demo", "0123456789abcdef0123456789abcdef", []byte("services: []")); err == nil {
		t.Fatal("invalid Compose manifest unexpectedly passed upload validation")
	}
	if _, err := validateComposeUpload("../escape", "0123456789abcdef0123456789abcdef", valid); err == nil {
		t.Fatal("unsafe application slug unexpectedly passed upload validation")
	}
}

func TestUploadComposeManifestCleansTemporaryFileOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client := newFakeSFTPClient()
	client.onWrite = cancel
	err := uploadComposeManifest(ctx, client, "/opt/arsy/apps/demo/releases/0123456789abcdef0123456789abcdef/compose.yaml", []byte(strings.Repeat("x", 64*1024)))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("upload error = %v, want context cancellation", err)
	}
	if !client.closed || len(client.removeCalls) != 1 {
		t.Fatalf("cancel cleanup: closed=%v remove=%#v", client.closed, client.removeCalls)
	}
	for name := range client.files {
		if strings.Contains(path.Base(name), ".tmp-") {
			t.Fatalf("temporary file was left behind: %q", name)
		}
	}
}

func TestUploadComposeManifestPropagatesRemotePermissionFailure(t *testing.T) {
	client := newFakeSFTPClient()
	client.mkdirErr = errors.New("permission denied")
	err := uploadComposeManifest(context.Background(), client, "/opt/arsy/apps/demo/releases/0123456789abcdef0123456789abcdef/compose.yaml", []byte("services:\n  web:\n    image: nginx\n"))
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("upload error = %v, want permission failure", err)
	}
	if len(client.files) != 0 {
		t.Fatal("file was opened after remote directory creation failed")
	}
}

func TestValidateComposeUploadEnforcesMaximumSizeAndDeploymentPath(t *testing.T) {
	valid := []byte("services:\n  web:\n    image: nginx\n")
	oversized := bytes.Repeat(valid, maxComposeManifestBytes/len(valid)+1)
	if _, err := validateComposeUpload("demo", "0123456789abcdef0123456789abcdef", oversized); err == nil {
		t.Fatal("oversized Compose manifest was accepted")
	}
	if _, err := validateComposeUpload("demo", "../../etc", valid); err == nil {
		t.Fatal("path traversal deployment ID accepted")
	}
}
