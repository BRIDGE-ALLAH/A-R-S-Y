package httpapi_test

import (
	"strings"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/httpapi"
	"github.com/bridge-allah/a-r-s-y/core"
)

func TestServerResponseOmitsCredentialFields(t *testing.T) {
	collection := core.NewBaseCollection("servers")
	record := core.NewRecord(collection)
	record.Id = "server12345678901"
	record.Set("name", "Test")
	record.Set("host", "example.com")
	record.Set("ssh_port", 22)
	record.Set("ssh_username", "deploy")
	record.Set("private_key_ciphertext", "ciphertext-secret-fixture")
	record.Set("secret", "raw-secret-fixture")

	encoded, err := httpapi.MarshalServerResponse(record)
	if err != nil {
		t.Fatal(err)
	}
	response := string(encoded)
	for _, forbidden := range []string{"private_key_ciphertext", "ciphertext-secret-fixture", "raw-secret-fixture", "secret"} {
		if strings.Contains(response, forbidden) {
			t.Fatalf("server response leaked forbidden field/value %q: %s", forbidden, response)
		}
	}
	for _, expected := range []string{`"id":"server12345678901"`, `"name":"Test"`, `"ssh_port":22`} {
		if !strings.Contains(response, expected) {
			t.Fatalf("server response omitted %s: %s", expected, response)
		}
	}
}
