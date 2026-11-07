package servers_test

import (
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/servers"
)

func TestValidateConnectionInput(t *testing.T) {
	privateKey := testPrivateKey(t)
	tests := []struct {
		name      string
		input     servers.ConnectionInput
		wantHost  string
		wantError bool
	}{
		{name: "DNS hostname normalized", input: servers.ConnectionInput{Host: "  NODE.Example.COM ", Port: 22, Username: "deploy", PrivateKey: privateKey}, wantHost: "node.example.com"},
		{name: "IPv4 retained", input: servers.ConnectionInput{Host: "192.0.2.10", Port: 2222, Username: "deploy", PrivateKey: privateKey}, wantHost: "192.0.2.10"},
		{name: "IPv6 retained", input: servers.ConnectionInput{Host: "2001:db8::1", Port: 22, Username: "deploy", PrivateKey: privateKey}, wantHost: "2001:db8::1"},
		{name: "empty host", input: servers.ConnectionInput{Host: " ", Port: 22, Username: "deploy", PrivateKey: privateKey}, wantError: true},
		{name: "invalid host", input: servers.ConnectionInput{Host: "bad host", Port: 22, Username: "deploy", PrivateKey: privateKey}, wantError: true},
		{name: "invalid port low", input: servers.ConnectionInput{Host: "example.com", Port: 0, Username: "deploy", PrivateKey: privateKey}, wantError: true},
		{name: "invalid port high", input: servers.ConnectionInput{Host: "example.com", Port: 65536, Username: "deploy", PrivateKey: privateKey}, wantError: true},
		{name: "invalid username", input: servers.ConnectionInput{Host: "example.com", Port: 22, Username: "bad user", PrivateKey: privateKey}, wantError: true},
		{name: "malformed key", input: servers.ConnectionInput{Host: "example.com", Port: 22, Username: "deploy", PrivateKey: []byte("not a key")}, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotHost, err := servers.ValidateConnectionInput(test.input)
			if (err != nil) != test.wantError {
				t.Fatalf("ValidateConnectionInput() error = %v, wantError %v", err, test.wantError)
			}
			if err == nil && gotHost != test.wantHost {
				t.Fatalf("normalized host = %q, want %q", gotHost, test.wantHost)
			}
		})
	}
}
