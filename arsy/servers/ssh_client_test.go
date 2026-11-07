package servers_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/bridge-allah/a-r-s-y/arsy/servers"
	"golang.org/x/crypto/ssh"
)

func TestNewSSHClientRequiresTrustedPinnedHost(t *testing.T) {
	serverSigner, err := ssh.ParsePrivateKey(testPrivateKey(t))
	if err != nil {
		t.Fatal(err)
	}
	address := startSSHTestServer(t, serverSigner, true)
	host, port := splitTestAddress(t, address)

	client, err := servers.NewSSHClient(context.Background(), servers.SSHClientConfig{
		Host: host, Port: port, Username: "deploy", Signer: serverSigner,
		PinnedFingerprint: servers.HostKeyFingerprint(serverSigner.PublicKey()),
		ConnectionTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("connect to matching pinned host: %v", err)
	}
	_ = client.Close()

	otherSigner, err := ssh.ParsePrivateKey(testPrivateKey(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := servers.NewSSHClient(context.Background(), servers.SSHClientConfig{
		Host: host, Port: port, Username: "deploy", Signer: serverSigner,
		PinnedFingerprint: servers.HostKeyFingerprint(otherSigner.PublicKey()),
		ConnectionTimeout: time.Second,
	}); err == nil {
		t.Fatal("accepted a server with a mismatched pinned fingerprint")
	}
}

func TestNewSSHClientRequiresPinAndHonorsCancellation(t *testing.T) {
	signer, err := ssh.ParsePrivateKey(testPrivateKey(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := servers.NewSSHClient(context.Background(), servers.SSHClientConfig{
		Host: "127.0.0.1", Port: 22, Username: "deploy", Signer: signer,
	}); err == nil {
		t.Fatal("accepted an SSH connection without a host-key pin")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := servers.NewSSHClient(ctx, servers.SSHClientConfig{
		Host: "127.0.0.1", Port: 22, Username: "deploy", Signer: signer,
		PinnedFingerprint: servers.HostKeyFingerprint(signer.PublicKey()),
	}); err == nil {
		t.Fatal("accepted a cancelled SSH connection")
	}
}

func startSSHTestServer(t *testing.T, hostSigner ssh.Signer, allowUser bool) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	serverConfig := &ssh.ServerConfig{PublicKeyCallback: func(metadata ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		if !allowUser || metadata.User() != "deploy" {
			return nil, ssh.ErrNoAuth
		}
		return &ssh.Permissions{}, nil
	}}
	serverConfig.AddHostKey(hostSigner)
	go func() {
		connection, err := listener.Accept()
		if err != nil {
			return
		}
		defer connection.Close()
		serverConnection, channels, requests, err := ssh.NewServerConn(connection, serverConfig)
		if err != nil {
			return
		}
		defer serverConnection.Close()
		go ssh.DiscardRequests(requests)
		for channel := range channels {
			_ = channel.Reject(ssh.Prohibited, "test server does not open channels")
		}
	}()
	return listener.Addr().String()
}

func splitTestAddress(t *testing.T, address string) (string, int) {
	t.Helper()
	host, portText, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatal(err)
	}
	var port int
	if _, err := fmt.Sscanf(portText, "%d", &port); err != nil {
		t.Fatal(err)
	}
	return host, port
}
