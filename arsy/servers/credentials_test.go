package servers

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"encoding/pem"
	"strings"
	"testing"

	_ "github.com/bridge-allah/a-r-s-y/arsy/migrations"
	"github.com/bridge-allah/a-r-s-y/arsy/security"
	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
	"golang.org/x/crypto/ssh"
)

func testOpenSSHPrivateKey(t *testing.T) []byte {
	t.Helper()
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	block, err := ssh.MarshalPrivateKey(privateKey, "arsy-test")
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(block)
}

func TestStorePrivateKeyPersistsCiphertextOnly(t *testing.T) {
	key, err := security.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("ARSY_SECRET_KEY", key)

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	collection, err := app.FindCollectionByNameOrId("arsy_servers")
	if err != nil {
		t.Fatal(err)
	}
	record := core.NewRecord(collection)
	record.Set("name", "test host")
	record.Set("host", "127.0.0.1")
	record.Set("ssh_port", 22)
	record.Set("ssh_username", "root")

	privateKey := testOpenSSHPrivateKey(t)
	if err := StorePrivateKey(record, privateKey, "ARSY_SECRET_KEY"); err != nil {
		t.Fatal(err)
	}
	if err := app.Save(record); err != nil {
		t.Fatal(err)
	}

	stored, err := app.FindRecordById("arsy_servers", record.Id)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext := stored.GetString("credential_ciphertext")
	if !strings.HasPrefix(ciphertext, "arsy:v1:") {
		t.Fatalf("stored credential is not versioned ciphertext: %q", ciphertext)
	}
	if bytes.Contains([]byte(ciphertext), privateKey) {
		t.Fatal("database record contains the original private key")
	}

	decrypted, err := LoadPrivateKey(stored, "ARSY_SECRET_KEY")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decrypted, privateKey) {
		t.Fatal("decrypted private key does not match original")
	}

	dto := ServerResponseFromRecord(stored)
	response, err := json.Marshal(dto)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"credential_ciphertext", ciphertext, string(privateKey)} {
		if strings.Contains(string(response), forbidden) {
			t.Fatalf("server DTO exposed secret data %q", forbidden)
		}
	}
}

func TestServerResponseIsAllowListed(t *testing.T) {
	record := core.NewRecord(core.NewBaseCollection("test"))
	record.Id = "server123456789"
	record.Set("name", "production")
	record.Set("host", "example.test")
	record.Set("credential_ciphertext", "arsy:v1:private-ciphertext")
	record.Set("ssh_username", "deploy")

	response, err := json.Marshal(ServerResponseFromRecord(record))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"credential_ciphertext", "private-ciphertext"} {
		if strings.Contains(string(response), forbidden) {
			t.Fatalf("response contains forbidden field/value %q", forbidden)
		}
	}
}

func TestRedactSecrets(t *testing.T) {
	got := RedactSecrets("ssh failed using key abc123 and password p455", "abc123", "p455")
	if strings.Contains(got, "abc123") || strings.Contains(got, "p455") {
		t.Fatalf("secret was not redacted: %s", got)
	}
}
