package servers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

const defaultSSHTimeout = 10 * time.Second

// NewSSHClient connects using an unencrypted private key and an administrator-
// pinned host fingerprint. It never accepts an unknown host key.
func NewSSHClient(ctx context.Context, address, username string, privateKey []byte, pinnedFingerprint string, timeout time.Duration) (*ssh.Client, error) {
	if ctx == nil {
		return nil, errors.New("SSH context is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(address) == "" {
		return nil, errors.New("SSH address is required")
	}
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("SSH username is required")
	}
	if strings.TrimSpace(pinnedFingerprint) == "" {
		return nil, ErrHostKeyNotPinned
	}
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("parse SSH private key: %w", err)
	}
	if timeout <= 0 {
		timeout = defaultSSHTimeout
	}
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("connect to SSH server: %w", err)
	}
	stopCancellation := context.AfterFunc(ctx, func() {
		_ = conn.Close()
	})
	defer stopCancellation()

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: PinnedHostKeyCallback(pinnedFingerprint),
		Timeout: timeout,
	}
	connection, channels, requests, err := ssh.NewClientConn(conn, address, config)
	if err != nil {
		_ = conn.Close()
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("SSH handshake failed: %w", err)
	}
	if err := ctx.Err(); err != nil {
		_ = connection.Close()
		return nil, err
	}
	return ssh.NewClient(connection, channels, requests), nil
}
