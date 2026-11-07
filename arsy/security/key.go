package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// GenerateKey returns a cryptographically random 256-bit key as 64 hex characters.
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("generate A-R-S-Y secret key: %w", err)
	}
	return hex.EncodeToString(key), nil
}
