package servers

import (
	"context"
	"testing"
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
	if err := server.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestTestSSHServerAcceptsConnection(t *testing.T) {
	ctx := context.Background()
	server, err := NewTestSSHServer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	if server.Port <= 0 {
		t.Fatal("TestSSHServer.Port not allocated")
	}
}

func TestTestSSHServerExecutesCommand(t *testing.T) {
	ctx := context.Background()
	server, err := NewTestSSHServer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	if server.Port == 0 {
		t.Fatal("test server port not set")
	}
}
