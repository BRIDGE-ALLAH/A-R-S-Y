package servers

import (
	"context"
	"net"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func TestProbeHostKeyReturnsUntrustedFingerprint(t *testing.T) {
	listener, hostKey := startTestSSHServer(t)
	defer listener.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	fingerprint, err := ProbeHostKey(ctx, listener.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if fingerprint != ssh.FingerprintSHA256(hostKey) {
		t.Fatalf("fingerprint = %q, want %q", fingerprint, ssh.FingerprintSHA256(hostKey))
	}
}

func TestTestSSHConnectionRequiresPinBeforeAuthentication(t *testing.T) {
	listener, hostKey := startTestSSHServer(t)
	defer listener.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := TestSSHConnection(ctx, "127.0.0.1", listener.Addr().(*net.TCPAddr).Port, "deploy", testOpenSSHPrivateKey(t), "", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if result.Authenticated || result.Code != "host_key_confirmation_required" || result.Fingerprint != ssh.FingerprintSHA256(hostKey) {
		t.Fatalf("unexpected first-step result: %#v", result)
	}
}
