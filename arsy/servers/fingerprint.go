package servers

import "golang.org/x/crypto/ssh"

// HostKeyFingerprint returns the canonical OpenSSH SHA-256 fingerprint.
func HostKeyFingerprint(publicKey ssh.PublicKey) string {
	if publicKey == nil {
		return ""
	}
	return ssh.FingerprintSHA256(publicKey)
}
