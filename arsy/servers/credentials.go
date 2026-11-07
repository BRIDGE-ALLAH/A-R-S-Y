package servers

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/bridge-allah/a-r-s-y/arsy/security"
	"github.com/bridge-allah/a-r-s-y/core"
	"golang.org/x/crypto/ssh"
)

var ErrEncryptedPrivateKeyUnsupported = errors.New("encrypted OpenSSH private keys are not supported yet")

// ValidatePrivateKey accepts unencrypted OpenSSH private-key material only.
func ValidatePrivateKey(privateKey []byte) error {
	privateKey = bytes.TrimSpace(privateKey)
	if !bytes.HasPrefix(privateKey, []byte("-----BEGIN OPENSSH PRIVATE KEY-----")) {
		return errors.New("unsupported SSH private-key format: expected an OpenSSH private key")
	}
	_, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		var passphraseErr *ssh.PassphraseMissingError
		if errors.As(err, &passphraseErr) {
			return ErrEncryptedPrivateKeyUnsupported
		}
		return fmt.Errorf("invalid OpenSSH private key: %w", err)
	}
	return nil
}

// StorePrivateKey encrypts key material before placing it on the server record.
// The caller must persist the record only after this method succeeds.
func StorePrivateKey(record *core.Record, privateKey []byte, secretKeyEnv string) error {
	if record == nil {
		return errors.New("server record is required")
	}
	if err := ValidatePrivateKey(privateKey); err != nil {
		return err
	}
	ciphertext, err := security.EncryptWithEnv(privateKey, secretKeyEnv)
	if err != nil {
		return err
	}
	record.Set("credential_ciphertext", ciphertext)
	return nil
}

// LoadPrivateKey decrypts a stored key for immediate SSH client construction.
// Callers must not cache or return the plaintext beyond that operation.
func LoadPrivateKey(record *core.Record, secretKeyEnv string) ([]byte, error) {
	if record == nil {
		return nil, errors.New("server record is required")
	}
	ciphertext := record.GetString("credential_ciphertext")
	if ciphertext == "" {
		return nil, errors.New("server has no stored SSH credential")
	}
	privateKey, err := security.DecryptWithEnv(ciphertext, secretKeyEnv)
	if err != nil {
		return nil, err
	}
	if err := ValidatePrivateKey(privateKey); err != nil {
		return nil, err
	}
	return privateKey, nil
}
