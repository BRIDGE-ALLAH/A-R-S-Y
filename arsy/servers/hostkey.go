package servers

import (
	"errors"
	"net"
	"strings"

	"golang.org/x/crypto/ssh"
)

var (
	ErrHostKeyNotPinned = errors.New("SSH host key has not been pinned")
	ErrHostKeyMismatch  = errors.New("SSH host key fingerprint does not match the pinned value")
)

// CheckHostKey reports the observed fingerprint and whether it matches the
// administrator-pinned fingerprint. An empty expected value is a first-seen
// key and is never treated as trusted.
func CheckHostKey(expected string, key ssh.PublicKey) (fingerprint string, trusted bool, err error) {
	fingerprint, err = HostKeyFingerprint(key)
	if err != nil {
		return "", false, err
	}
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return fingerprint, false, nil
	}
	if expected != fingerprint {
		return fingerprint, false, ErrHostKeyMismatch
	}
	return fingerprint, true, nil
}

// PinnedHostKeyCallback returns an SSH callback that accepts only the exact
// fingerprint previously confirmed by an administrator.
func PinnedHostKeyCallback(expected string) ssh.HostKeyCallback {
	return func(_ string, _ net.Addr, key ssh.PublicKey) error {
		_, trusted, err := CheckHostKey(expected, key)
		if err != nil {
			return err
		}
		if !trusted {
			return ErrHostKeyNotPinned
		}
		return nil
	}
}
