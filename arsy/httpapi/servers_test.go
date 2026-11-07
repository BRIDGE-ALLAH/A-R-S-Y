package httpapi_test

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/security"
	"golang.org/x/crypto/ssh"
)

func TestServerManagementRoutesRequireSuperuser(t *testing.T) {
	app, mux := newTestMux(t, false)
	defer app.Cleanup()

	request := httptest.NewRequest(http.MethodGet, "/api/arsy/v1/servers", nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("list servers status = %d, want 401", response.Code)
	}
}

func TestServerCreateListGetAndUpdateNeverExposeCredentials(t *testing.T) {
	key, err := security.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("ARSY_SECRET_KEY", key)
	privateKey := testAPIPrivateKey(t)
	app, mux := newTestMux(t, true)
	defer app.Cleanup()

	body, err := json.Marshal(map[string]any{
		"name": "test server",
		"host": " Example.COM. ",
		"ssh_port": 2222,
		"ssh_username": "deploy",
		"private_key": string(privateKey),
	})
	if err != nil {
		t.Fatal(err)
	}
	create := httptest.NewRequest(http.MethodPost, "/api/arsy/v1/servers", bytes.NewReader(body))
	create.Header.Set("Content-Type", "application/json")
	created := httptest.NewRecorder()
	mux.ServeHTTP(created, create)
	if created.Code != http.StatusCreated {
		t.Fatalf("create server status = %d: %s", created.Code, created.Body.String())
	}
	if strings.Contains(created.Body.String(), "credential_ciphertext") || strings.Contains(created.Body.String(), string(privateKey)) {
		t.Fatal("create server response exposed SSH credentials")
	}
	var dto map[string]any
	if err := json.Unmarshal(created.Body.Bytes(), &dto); err != nil {
		t.Fatal(err)
	}
	id, _ := dto["id"].(string)
	if id == "" || dto["host"] != "example.com" {
		t.Fatalf("unexpected server response: %#v", dto)
	}
	record, err := app.FindRecordById("arsy_servers", id)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext := record.GetString("credential_ciphertext")
	if !strings.HasPrefix(ciphertext, "arsy:v1:") || strings.Contains(ciphertext, string(privateKey)) {
		t.Fatal("server record does not contain encrypted SSH credentials")
	}

	list := httptest.NewRecorder()
	mux.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/api/arsy/v1/servers", nil))
	if list.Code != http.StatusOK || strings.Contains(list.Body.String(), ciphertext) {
		t.Fatalf("list servers response = %d: %s", list.Code, list.Body.String())
	}

	get := httptest.NewRecorder()
	mux.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/arsy/v1/servers/"+id, nil))
	if get.Code != http.StatusOK || strings.Contains(get.Body.String(), ciphertext) {
		t.Fatalf("get server response = %d: %s", get.Code, get.Body.String())
	}

	updateBody := bytes.NewBufferString(`{"name":"renamed server"}`)
	update := httptest.NewRequest(http.MethodPatch, "/api/arsy/v1/servers/"+id, updateBody)
	updated := httptest.NewRecorder()
	mux.ServeHTTP(updated, update)
	if updated.Code != http.StatusOK || !strings.Contains(updated.Body.String(), "renamed server") {
		t.Fatalf("update server response = %d: %s", updated.Code, updated.Body.String())
	}
}

func TestServerCreateRejectsInvalidConnectionInput(t *testing.T) {
	key, err := security.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("ARSY_SECRET_KEY", key)
	app, mux := newTestMux(t, true)
	defer app.Cleanup()

	body := bytes.NewBufferString(`{"name":"bad","host":"bad host","ssh_port":70000,"ssh_username":"root","private_key":"not-a-key"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/arsy/v1/servers", body)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid server status = %d: %s", response.Code, response.Body.String())
	}
}

func testAPIPrivateKey(t *testing.T) []byte {
	t.Helper()
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	block, err := ssh.MarshalPrivateKey(privateKey, "arsy-api-test")
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(block)
}

