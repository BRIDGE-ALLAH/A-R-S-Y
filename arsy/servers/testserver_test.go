package servers

import (
	"context"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func TestTestSSHServerStartsAndStops(t *testing.T) {
	ctx := context.Background()
	server, err := NewTestSSHServer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	addr := server.Addr()
	if addr == "" {
		t.Fatal("TestSSHServer.Addr() returned empty")
	}
	if server.Port <= 0 {
		t.Fatal("TestSSHServer.Port not allocated")
	}
	if err := server.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestTestSSHServerHostKeyFingerprint(t *testing.T) {
	ctx := context.Background()
	server, err := NewTestSSHServer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	fp := server.HostKeyFingerprint()
	if fp == "" {
		t.Fatal("HostKeyFingerprint() returned empty")
	}
	if fp != ssh.FingerprintSHA256(server.HostSigner.PublicKey()) {
		t.Fatalf("fingerprint mismatch: got %q", fp)
	}
}

func TestTestSSHServerAcceptsSSHAuthAndExec(t *testing.T) {
	ctx := context.Background()
	server, err := NewTestSSHServer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	privateKey := testOpenSSHPrivateKey(t)
	fingerprint := server.HostKeyFingerprint()

	sshCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client, err := NewSSHClient(sshCtx, server.Addr(), "deploy", privateKey, fingerprint, 2*time.Second)
	if err != nil {
		t.Fatalf("NewSSHClient() error = %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput("echo ok")
	if err != nil {
		t.Fatalf("CombinedOutput() error = %v", err)
	}
	if string(output) != "ok\n" {
		t.Fatalf("output = %q, want %q", string(output), "ok\n")
	}
}

func TestTestSSHServerReadinessCommands(t *testing.T) {
	ctx := context.Background()
	server, err := NewTestSSHServer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	privateKey := testOpenSSHPrivateKey(t)
	fingerprint := server.HostKeyFingerprint()

	sshCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client, err := NewSSHClient(sshCtx, server.Addr(), "deploy", privateKey, fingerprint, 2*time.Second)
	if err != nil {
		t.Fatalf("NewSSHClient() error = %v", err)
	}
	defer client.Close()

	tests := []struct {
		cmd  string
		want string
	}{
		{"uname -s && uname -m", "Linux\nx86_64\n"},
		{"docker version --format '{{.Server.Version}}'", "Docker version 24.0.7, build afdd53b\n"},
		{"docker compose version --short", "Docker Compose version v2.23.0\n"},
	}

	for _, tt := range tests {
		session, err := client.NewSession()
		if err != nil {
			t.Fatalf("NewSession(%q) error = %v", tt.cmd, err)
		}
		output, err := session.CombinedOutput(tt.cmd)
		session.Close()
		if err != nil {
			t.Fatalf("CombinedOutput(%q) error = %v", tt.cmd, err)
		}
		if string(output) != tt.want {
			t.Fatalf("output for %q = %q, want %q", tt.cmd, string(output), tt.want)
		}
	}
}
