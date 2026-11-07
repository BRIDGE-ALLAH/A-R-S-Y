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

var errHostKeyProbeComplete = errors.New("SSH host key probe complete")

type ConnectionTestResult struct {
	Fingerprint   string
	Authenticated bool
	Code          string
	Message       string
}

// ProbeHostKey obtains the SSH host key without authenticating or trusting it.
// The returned fingerprint must be explicitly confirmed by an administrator.
func ProbeHostKey(ctx context.Context, address string, timeout time.Duration) (string, error) {
	if ctx == nil {
		return "", errors.New("SSH context is required")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if timeout <= 0 {
		timeout = defaultSSHTimeout
	}
	conn, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, "tcp", address)
	if err != nil {
		return "", fmt.Errorf("connect to SSH host for fingerprint probe: %w", err)
	}
	stopCancellation := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stopCancellation()
	var fingerprint string
	config := &ssh.ClientConfig{
		User:            "arsy-host-key-probe",
		Timeout:         timeout,
		HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
			fingerprint, err = HostKeyFingerprint(key)
			if err != nil {
				return err
			}
			return errHostKeyProbeComplete
		},
	}
	_, _, _, handshakeErr := ssh.NewClientConn(conn, address, config)
	_ = conn.Close()
	if ctx.Err() != nil {
		return fingerprint, ctx.Err()
	}
	if errors.Is(handshakeErr, errHostKeyProbeComplete) {
		return fingerprint, nil
	}
	if handshakeErr != nil {
		return fingerprint, fmt.Errorf("probe SSH host key: %w", handshakeErr)
	}
	return fingerprint, errors.New("SSH host key probe did not complete")
}

// TestSSHConnection performs a two-step trust flow: first it reports an
// untrusted fingerprint; after pinning, it authenticates with the SSH key.
func TestSSHConnection(ctx context.Context, host string, port int, username string, privateKey []byte, pinnedFingerprint string, timeout time.Duration) (ConnectionTestResult, error) {
	address := net.JoinHostPort(host, strconv.Itoa(port))
	if pinnedFingerprint == "" {
		fingerprint, err := ProbeHostKey(ctx, address, timeout)
		if err != nil {
			return ConnectionTestResult{Code: "probe_failed", Message: "Unable to retrieve the SSH host key"}, err
		}
		return ConnectionTestResult{
			Fingerprint: fingerprint,
			Code:        "host_key_confirmation_required",
			Message:     "Confirm this fingerprint before authenticating",
		}, nil
	}
	client, err := NewSSHClient(ctx, address, username, privateKey, pinnedFingerprint, timeout)
	if err != nil {
		if errors.Is(err, ErrHostKeyMismatch) {
			return ConnectionTestResult{Code: "host_key_mismatch", Message: "The SSH host key changed; verify the host before updating its pin"}, err
		}
		return ConnectionTestResult{Code: "connection_failed", Message: "SSH authentication or connection failed"}, err
	}
	_ = client.Close()
	return ConnectionTestResult{
		Fingerprint:   pinnedFingerprint,
		Authenticated: true,
		Code:          "connected",
		Message:       "SSH authentication succeeded and the host key matched",
	}, nil
}
