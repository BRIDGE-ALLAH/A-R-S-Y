package servers

import (
	"strings"
	"testing"
)

func TestValidateConnectionInput(t *testing.T) {
	validKey := testOpenSSHPrivateKey(t)
	tests := []struct {
		name  string
		input ConnectionInput
		wantHost string
		wantErr string
	}{
		{name: "valid hostname", input: ConnectionInput{Host: " Example.COM. ", Port: 22, Username: "deploy", PrivateKey: validKey}, wantHost: "example.com"},
		{name: "valid IPv4", input: ConnectionInput{Host: "192.0.2.10", Port: 2222, Username: "root", PrivateKey: validKey}, wantHost: "192.0.2.10"},
		{name: "valid IPv6", input: ConnectionInput{Host: "2001:db8::1", Port: 22, Username: "root", PrivateKey: validKey}, wantHost: "2001:db8::1"},
		{name: "empty host", input: ConnectionInput{Port: 22, Username: "root", PrivateKey: validKey}, wantErr: "host is required"},
		{name: "invalid hostname", input: ConnectionInput{Host: "bad host", Port: 22, Username: "root", PrivateKey: validKey}, wantErr: "invalid characters"},
		{name: "invalid port low", input: ConnectionInput{Host: "host.test", Port: 0, Username: "root", PrivateKey: validKey}, wantErr: "port must be"},
		{name: "invalid port high", input: ConnectionInput{Host: "host.test", Port: 65536, Username: "root", PrivateKey: validKey}, wantErr: "port must be"},
		{name: "invalid username", input: ConnectionInput{Host: "host.test", Port: 22, Username: "root user", PrivateKey: validKey}, wantErr: "username"},
		{name: "invalid key", input: ConnectionInput{Host: "host.test", Port: 22, Username: "root", PrivateKey: []byte("not a key")}, wantErr: "private key"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateConnectionInput(tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Host != tt.wantHost {
				t.Fatalf("host = %q, want %q", got.Host, tt.wantHost)
			}
		})
	}
}
