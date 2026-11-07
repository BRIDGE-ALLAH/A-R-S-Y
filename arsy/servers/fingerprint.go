package servers

import (
	"errors"

	"golang.org/x/crypto/ssh"
)

// HostKeyFingerprint returns the canonical OpenSSH SHA-256 fingerprint.
func HostKeyFingerprint(key ssh.PublicKey) (string, error) {
	if key == nil {
		return "", errors.New("SSH host key is required")
	}
	return ssh.FingerprintSHA256(key), nil
}
