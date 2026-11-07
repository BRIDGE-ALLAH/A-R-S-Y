package servers

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"net"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func TestNewSSHClientConnectsOnlyWithPinnedHostKey(t *testing.T) {
	listener, hostKey := startTestSSHServer(t)
	defer listener.Close()

	clientKey := testOpenSSHPrivateKey(t)
	fingerprint := ssh.FingerprintSHA256(hostKey)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client, err := NewSSHClient(ctx, listener.Addr().String(), "deploy", clientKey, fingerprint, time.Second)
	if err != nil {
		t.Fatalf("NewSSHClient() error = %v", err)
	}
	_ = client.Close()
}

func TestNewSSHClientRejectsMissingOrChangedPin(t *testing.T) {
	listener, hostKey := startTestSSHServer(t)
	defer listener.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	key := testOpenSSHPrivateKey(t)

	if _, err := NewSSHClient(ctx, listener.Addr().String(), "deploy", key, "", time.Second); err != ErrHostKeyNotPinned {
		t.Fatalf("missing pin error = %v, want ErrHostKeyNotPinned", err)
	}
	if _, err := NewSSHClient(ctx, listener.Addr().String(), "deploy", key, "SHA256:wrong", time.Second); err == nil {
		t.Fatal("SSH client accepted a changed host key")
	}

	_ = hostKey
}

func startTestSSHServer(t *testing.T) (net.Listener, ssh.PublicKey) {
	t.Helper()
	_, hostPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	hostSigner, err := ssh.NewSignerFromKey(hostPrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	serverConfig := &ssh.ServerConfig{
		PublicKeyCallback: func(_ ssh.ConnMetadata, _ ssh.PublicKey) (*ssh.Permissions, error) {
			return nil, nil
		},
	}
	serverConfig.AddHostKey(hostSigner)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		serverConn, channels, requests, err := ssh.NewServerConn(conn, serverConfig)
		if err != nil {
			_ = conn.Close()
			return
		}
		go ssh.DiscardRequests(requests)
		go func() {
			for channel := range channels {
				_ = channel.Reject(ssh.UnknownChannelType, "not used in test")
			}
		}()
		_ = serverConn.Wait()
		_ = serverConn.Close()
	}()
	return listener, hostSigner.PublicKey()
}
