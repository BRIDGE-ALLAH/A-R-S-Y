package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

const ciphertextPrefix = "arsy:v1:"

var ErrInvalidKey = errors.New("A-R-S-Y secret key must be exactly 32 bytes")

// Encrypt protects plaintext with AES-256-GCM and encodes a versioned ciphertext.
func Encrypt(plaintext []byte, key string) (string, error) {
	block, err := newCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("initialize AES-GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate encryption nonce: %w", err)
	}
	sealed := gcm.Seal(nil, nonce, plaintext, []byte(ciphertextPrefix))
	encoded := base64.RawStdEncoding.EncodeToString(append(nonce, sealed...))
	return ciphertextPrefix + encoded, nil
}

// Decrypt authenticates and opens a versioned A-R-S-Y ciphertext.
func Decrypt(ciphertext string, key string) ([]byte, error) {
	block, err := newCipher(key)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(ciphertext, ciphertextPrefix) {
		return nil, errors.New("unsupported A-R-S-Y ciphertext format")
	}
	encoded, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(ciphertext, ciphertextPrefix))
	if err != nil {
		return nil, errors.New("invalid A-R-S-Y ciphertext encoding")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("initialize AES-GCM: %w", err)
	}
	if len(encoded) < gcm.NonceSize()+gcm.Overhead() {
		return nil, errors.New("invalid A-R-S-Y ciphertext length")
	}
	nonce, sealed := encoded[:gcm.NonceSize()], encoded[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, sealed, []byte(ciphertextPrefix))
	if err != nil {
		return nil, errors.New("A-R-S-Y ciphertext authentication failed")
	}
	return plaintext, nil
}

func newCipher(key string) (cipher.Block, error) {
	keyBytes := []byte(key)
	if len(keyBytes) == 64 {
		decoded := make([]byte, 32)
		if _, err := hex.Decode(decoded, keyBytes); err == nil {
			keyBytes = decoded
		}
	}
	if len(keyBytes) != 32 {
		return nil, ErrInvalidKey
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, ErrInvalidKey
	}
	return block, nil
}
