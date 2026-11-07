package deployments

import (
	"errors"
	"strings"

	"github.com/bridge-allah/a-r-s-y/arsy/security"
)

// EncryptManifest encrypts a Compose manifest using the configured secret key.
func EncryptManifest(plaintext, secretKeyEnv string) (string, error) {
	if strings.TrimSpace(plaintext) == "" {
		return "", errors.New("manifest must not be empty")
	}
	return security.EncryptWithEnv([]byte(plaintext), secretKeyEnv)
}

// DecryptManifest decrypts a Compose manifest ciphertext.
// Falls back to plaintext if value does not look like ciphertext (backward-compat).
func DecryptManifest(value, secretKeyEnv string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", errors.New("manifest ciphertext is empty")
	}
	if !strings.HasPrefix(value, security.CiphertextPrefix) {
		return value, nil
	}
	plaintext, err := security.DecryptWithEnv(value, secretKeyEnv)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
