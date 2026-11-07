package servers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHClientConfig contains the trusted authentication and transport inputs for an SSH connection.
type SSHClientConfig struct {
	Host              string
	Port              int
	Username          string
	Signer            ssh.Signer
	PinnedFingerprint string
	ConnectionTimeout time.Duration
}

// NewSSHClient connects with public-key authentication and strict pinned host-key verification.
func NewSSHClient(ctx context.Context, config SSHClientConfig) (*ssh.Client, error) {
	if ctx == nil {
		return nil, errors.New("SSH context must not be nil")
	}
	if config.Signer == nil {
		return nil, errors.New("SSH private-key signer is required")
	}
	if config.PinnedFingerprint == "" {
		return nil, errors.New("SSH host-key fingerprint must be pinned")
	}
	if config.Port < 1 || config.Port > 65535 {
		return nil, errors.New("SSH port must be between 1 and 65535")
	}
	if config.Host == "" || config.Username == "" {
		return nil, errors.New("SSH host and username are required")
	}
	if config.ConnectionTimeout <= 0 {
		config.ConnectionTimeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, config.ConnectionTimeout)
	defer cancel()

	address := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	connection, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("connect to SSH server: %w", err)
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = connection.SetDeadline(deadline)
	}
	stopClose := context.AfterFunc(ctx, func() { _ = connection.Close() })
	clientConnection, channels, requests, err := ssh.NewClientConn(connection, address, &ssh.ClientConfig{
		User:            config.Username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(config.Signer)},
		HostKeyCallback: HostKeyCallback(config.PinnedFingerprint),
		Timeout:         config.ConnectionTimeout,
	})
	stopClose()
	if err != nil {
		_ = connection.Close()
		if ctx.Err() != nil {
			return nil, fmt.Errorf("SSH connection cancelled or timed out: %w", ctx.Err())
		}
		return nil, fmt.Errorf("establish authenticated SSH connection: %w", err)
	}
	if err := ctx.Err(); err != nil {
		_ = clientConnection.Close()
		return nil, fmt.Errorf("SSH connection cancelled or timed out: %w", err)
	}
	_ = connection.SetDeadline(time.Time{})
	return ssh.NewClient(clientConnection, channels, requests), nil
}
