package servers

import (
	"errors"
	"fmt"

	"github.com/bridge-allah/a-r-s-y/arsy/security"
	"github.com/bridge-allah/a-r-s-y/core"
	"golang.org/x/crypto/ssh"
)

const privateKeyField = "private_key_ciphertext"

// StorePrivateKey validates an unencrypted OpenSSH private key and persists only its ciphertext.
func StorePrivateKey(app core.App, serverID string, privateKey []byte, secretKey string) error {
	if _, err := ssh.ParsePrivateKey(privateKey); err != nil {
		var passphraseError *ssh.PassphraseMissingError
		if errors.As(err, &passphraseError) {
			return errors.New("encrypted SSH private keys are not supported; provide an unencrypted key")
		}
		return errors.New("invalid SSH private key")
	}
	ciphertext, err := security.Encrypt(privateKey, secretKey)
	if err != nil {
		return fmt.Errorf("encrypt SSH private key: %w", err)
	}
	server, err := app.FindRecordById("servers", serverID)
	if err != nil {
		return fmt.Errorf("find server: %w", err)
	}
	server.Set(privateKeyField, ciphertext)
	if err := app.Save(server); err != nil {
		return fmt.Errorf("save encrypted SSH private key: %w", err)
	}
	return nil
}

// WithPrivateKey decrypts a server key only for the duration of the callback.
func WithPrivateKey(app core.App, serverID string, secretKey string, use func(ssh.Signer) error) error {
	if use == nil {
		return errors.New("SSH private key callback must not be nil")
	}
	server, err := app.FindRecordById("servers", serverID)
	if err != nil {
		return fmt.Errorf("find server: %w", err)
	}
	ciphertext := server.GetString(privateKeyField)
	if ciphertext == "" {
		return errors.New("server has no SSH private key")
	}
	privateKey, err := security.Decrypt(ciphertext, secretKey)
	if err != nil {
		return fmt.Errorf("decrypt SSH private key: %w", err)
	}
	defer clear(privateKey)
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return errors.New("stored SSH private key is invalid or unsupported")
	}
	return use(signer)
}
