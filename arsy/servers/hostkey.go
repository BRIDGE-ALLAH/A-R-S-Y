package servers

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"
)

// UnpinnedHostKeyError reports the candidate fingerprint without trusting it.
type UnpinnedHostKeyError struct {
	Fingerprint string
}

func (e *UnpinnedHostKeyError) Error() string {
	return fmt.Sprintf("SSH host key is not pinned; review and confirm fingerprint %s", e.Fingerprint)
}

// RequirePinnedHostKey rejects missing or changed pins and never persists trust.
func RequirePinnedHostKey(pinned string, observed ssh.PublicKey) error {
	if observed == nil {
		return errors.New("SSH server did not provide a host key")
	}
	fingerprint := HostKeyFingerprint(observed)
	if pinned == "" {
		return &UnpinnedHostKeyError{Fingerprint: fingerprint}
	}
	if subtle.ConstantTimeCompare([]byte(pinned), []byte(fingerprint)) != 1 {
		return fmt.Errorf("SSH host key fingerprint changed: expected %s, observed %s", pinned, fingerprint)
	}
	return nil
}

// HostKeyCallback returns a strict callback that requires the stored pin.
func HostKeyCallback(pinned string) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		if err := RequirePinnedHostKey(pinned, key); err != nil {
			return fmt.Errorf("verify SSH host key for %s: %w", hostname, err)
		}
		return nil
	}
}
