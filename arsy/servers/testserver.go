package servers

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// TestSSHServer is a deterministic SSH test server for integration tests.
// It supports host-key handshake, public-key auth, and basic command execution.
// All operations are network-local (127.0.0.1) and deterministic.
type TestSSHServer struct {
	Listener   net.Listener
	Port       int
	HostSigner ssh.Signer
	Address    string
	mu         sync.Mutex
	closed     bool
	done       chan struct{}
}

// NewTestSSHServer creates and starts a deterministic test SSH server.
// It accepts public-key authentication (any key) and handles exec requests
// with canned responses for common readiness/deployment commands.
func NewTestSSHServer(ctx context.Context) (*TestSSHServer, error) {
	_, hostPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate host key: %w", err)
	}
	hostSigner, err := ssh.NewSignerFromKey(hostPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("create host signer: %w", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	ts := &TestSSHServer{
		Listener:   listener,
		Port:       listener.Addr().(*net.TCPAddr).Port,
		HostSigner: hostSigner,
		Address:    listener.Addr().String(),
		done:       make(chan struct{}),
	}

	serverConfig := &ssh.ServerConfig{
		PublicKeyCallback: func(_ ssh.ConnMetadata, _ ssh.PublicKey) (*ssh.Permissions, error) {
			return &ssh.Permissions{}, nil
		},
	}
	serverConfig.AddHostKey(hostSigner)

	go ts.acceptLoop(serverConfig)

	return ts, nil
}

func (ts *TestSSHServer) acceptLoop(config *ssh.ServerConfig) {
	defer close(ts.done)
	for {
		conn, err := ts.Listener.Accept()
		if err != nil {
			return
		}
		go ts.handleConn(conn, config)
	}
}

func (ts *TestSSHServer) handleConn(conn net.Conn, config *ssh.ServerConfig) {
	serverConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		_ = conn.Close()
		return
	}
	defer serverConn.Close()
	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}
		go ts.handleSession(channel, requests)
	}
}

func (ts *TestSSHServer) handleSession(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()
	for req := range requests {
		switch req.Type {
		case "exec":
			cmd := parseExecCommand(req.Payload)
			_ = req.Reply(true, nil)
			output, exitCode := cannedCommandResponse(cmd)
			_, _ = io.WriteString(channel, output)
			_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{uint32(exitCode)}))
			return
		default:
			_ = req.Reply(false, nil)
		}
	}
}

func parseExecCommand(payload []byte) string {
	if len(payload) < 4 {
		return ""
	}
	cmdLen := int(payload[0])<<24 | int(payload[1])<<16 | int(payload[2])<<8 | int(payload[3])
	if cmdLen <= 0 || 4+cmdLen > len(payload) {
		return ""
	}
	return string(payload[4 : 4+cmdLen])
}

// cannedCommandResponse returns deterministic responses for known commands.
func cannedCommandResponse(cmd string) (string, int) {
	// Match by the actual shell commands emitted by commandForOperation.
	switch {
	case strings.Contains(cmd, "uname -s") && strings.Contains(cmd, "uname -m"):
		return "Linux\nx86_64\n", 0
	case strings.Contains(cmd, "docker version"):
		return "Docker version 24.0.7, build afdd53b\n", 0
	case strings.Contains(cmd, "docker compose version"):
		return "Docker Compose version v2.23.0\n", 0
	case cmd == "echo ok":
		return "ok\n", 0
	case cmd == "exit 1" || cmd == "false":
		return "", 1
	case strings.Contains(cmd, "docker compose"):
		// Deterministic success for compose operations (validate/deploy/inspect/stop).
		if strings.Contains(cmd, "ps --format json") {
			return `[{"Name":"test_web_1","State":"running","Status":"Up"}]` + "\n", 0
		}
		return "", 0
	default:
		return "", 0
	}
}

// Addr returns the network address the test server is listening on.
func (ts *TestSSHServer) Addr() string {
	if ts.Listener == nil {
		return ""
	}
	return ts.Listener.Addr().String()
}

// HostKeyFingerprint returns the SHA256 fingerprint of the test server's host key.
func (ts *TestSSHServer) HostKeyFingerprint() string {
	return ssh.FingerprintSHA256(ts.HostSigner.PublicKey())
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
		_ = ts.Listener.Close()
	}
	select {
	case <-ts.done:
	case <-time.After(2 * time.Second):
	}
	return nil
}
