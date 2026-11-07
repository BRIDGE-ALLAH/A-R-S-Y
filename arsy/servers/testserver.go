package servers

import (
	"context"
	"fmt"
	"net"
	"sync"
)

// TestSSHServer is a mock SSH server for testing without external infrastructure.
// Note: This is a placeholder for MVP-1; full SSH server implementation deferred.
type TestSSHServer struct {
	Listener net.Listener
	Port     int
	mu       sync.Mutex
	closed   bool
}

// NewTestSSHServer creates a mock test SSH server.
// For MVP-1, this is a skeletal implementation; full SSH behavior deferred.
func NewTestSSHServer(ctx context.Context) (*TestSSHServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	ts := &TestSSHServer{
		Listener: listener,
		Port:     listener.Addr().(*net.TCPAddr).Port,
	}

	return ts, nil
}

// Addr returns the network address the test server is listening on.
func (ts *TestSSHServer) Addr() string {
	if ts.Listener == nil {
		return ""
	}
	return ts.Listener.Addr().String()
}

// Close gracefully shuts down the test SSH server.
func (ts *TestSSHServer) Close() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.closed {
		return nil
	}
	ts.closed = true
	if ts.Listener != nil {
		return ts.Listener.Close()
	}
	return nil
}
