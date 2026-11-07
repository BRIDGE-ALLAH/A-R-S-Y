package servers_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	_ "github.com/bridge-allah/a-r-s-y/arsy/migrations"
	"github.com/bridge-allah/a-r-s-y/arsy/servers"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
	"golang.org/x/crypto/ssh"
)

const testSecretKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestStorePrivateKeyEncryptsAndLoadsOnlyThroughCallback(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	if err := app.RunSystemMigrations(); err != nil {
		t.Fatal(err)
	}
	// Ensure this package's migrations are registered by its import before the test runs.
	serverCollection, err := app.FindCollectionByNameOrId("servers")
	if err != nil {
		t.Fatal(err)
	}
	record := core.NewRecord(serverCollection)
	record.Set("name", "Test host")
	record.Set("host", "example.com")
	record.Set("ssh_port", 22)
	record.Set("ssh_username", "deploy")
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}

	privateKey := testPrivateKey(t)
	if err := servers.StorePrivateKey(app, record.Id, privateKey, testSecretKey); err != nil {
		t.Fatalf("store private key: %v", err)
	}
	stored, err := app.FindRecordById("servers", record.Id)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext := stored.GetString("private_key_ciphertext")
	if ciphertext == "" || strings.Contains(ciphertext, string(privateKey)) || strings.Contains(ciphertext, "BEGIN RSA PRIVATE KEY") {
		t.Fatal("database record did not contain ciphertext-only key material")
	}
	if stored.PublicExport()["private_key_ciphertext"] != nil {
		t.Fatal("public export exposed private-key ciphertext")
	}
	if err := servers.WithPrivateKey(app, record.Id, testSecretKey, func(signer ssh.Signer) error {
		if signer == nil {
			t.Fatal("callback received a nil signer")
		}
		return nil
	}); err != nil {
		t.Fatalf("use stored private key: %v", err)
	}
	if err := servers.WithPrivateKey(app, record.Id, "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789", func(ssh.Signer) error { return nil }); err == nil {
		t.Fatal("expected wrong secret key to fail")
	}
}

func TestStorePrivateKeyRejectsMalformedInput(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()
	if err := app.RunSystemMigrations(); err != nil {
		t.Fatal(err)
	}
	serverCollection, err := app.FindCollectionByNameOrId("servers")
	if err != nil {
		t.Fatal(err)
	}
	record := core.NewRecord(serverCollection)
	record.Set("name", "Test host")
	record.Set("host", "example.com")
	record.Set("ssh_port", 22)
	record.Set("ssh_username", "deploy")
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}
	if err := servers.StorePrivateKey(app, record.Id, []byte("malformed"), testSecretKey); err == nil {
		t.Fatal("expected malformed SSH private key to be rejected")
	}
	encryptedBlock, err := x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", mustRSAKey(t), []byte("fixture-passphrase"), x509.PEMCipherAES256)
	if err != nil {
		t.Fatal(err)
	}
	if err := servers.StorePrivateKey(app, record.Id, pem.EncodeToMemory(encryptedBlock), testSecretKey); err == nil || !strings.Contains(err.Error(), "encrypted SSH private keys are not supported") {
		t.Fatalf("expected clear encrypted-key rejection, got %v", err)
	}
}

func testPrivateKey(t *testing.T) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: mustRSAKey(t)})
}

func mustRSAKey(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return x509.MarshalPKCS1PrivateKey(key)
}
