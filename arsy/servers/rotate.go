package servers

import (
	"errors"
	"fmt"
	"sync"

	"github.com/bridge-allah/a-r-s-y/arsy/security"
	"github.com/bridge-allah/a-r-s-y/core"
)

var (
	sshClientCache = make(map[string]struct{})
	sshCacheMu     sync.RWMutex
)

// RotateServerCredential replaces a server's SSH credential atomically.
// It encrypts the new key, writes it to the server record, and invalidates
// any cached connections for that server.
func RotateServerCredential(app core.App, serverID, newPrivateKeyPEM string, secretKeyEnv string) error {
	if app == nil {
		return errors.New("PocketBase app is required")
	}
	if serverID == "" {
		return errors.New("server ID is required")
	}
	if newPrivateKeyPEM == "" {
		return errors.New("new private key is required")
	}
	record, err := app.FindRecordById("arsy_servers", serverID)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}
	key, err := security.LoadKey(secretKeyEnv)
	if err != nil {
		return fmt.Errorf("load encryption key: %w", err)
	}
	newCiphertext, err := security.Encrypt(key, []byte(newPrivateKeyPEM))
	if err != nil {
		return fmt.Errorf("encrypt new credential: %w", err)
	}
	record.Set("credential_ciphertext", newCiphertext)
	if err := app.Save(record); err != nil {
		return fmt.Errorf("save rotated credential: %w", err)
	}
	sshCacheMu.Lock()
	delete(sshClientCache, serverID)
	sshCacheMu.Unlock()
	return nil
}
