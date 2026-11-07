package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	KeySizeBytes    = 32
	CiphertextPrefix = "arsy:v1"
)

var (
	ErrMissingKey             = errors.New("A-R-S-Y secret key is not configured")
	ErrInvalidKey             = errors.New("A-R-S-Y secret key must be base64-encoded 32-byte material")
	ErrInvalidCiphertext      = errors.New("invalid A-R-S-Y ciphertext format")
	ErrAuthenticationFailed  = errors.New("A-R-S-Y ciphertext authentication failed")
)

// GenerateKey returns a cryptographically random 32-byte key encoded as
// unpadded standard base64. It never writes the key to disk or logs it.
func GenerateKey() (string, error) {
	key := make([]byte, KeySizeBytes)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("generate A-R-S-Y secret key: %w", err)
	}
	return base64.RawStdEncoding.EncodeToString(key), nil
}

// ValidateKey validates the externally supplied base64 key without retaining it.
func ValidateKey(encoded string) error {
	_, err := decodeKey(encoded)
	return err
}

// LoadKey reads and validates a key from the named environment variable.
func LoadKey(envName string) ([]byte, error) {
	if strings.TrimSpace(envName) == "" {
		return nil, ErrMissingKey
	}
	value := os.Getenv(envName)
	if value == "" {
		return nil, ErrMissingKey
	}
	return decodeKey(value)
}

func decodeKey(encoded string) ([]byte, error) {
	key, err := base64.RawStdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		key, err = base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	}
	if err != nil || len(key) != KeySizeBytes {
		return nil, ErrInvalidKey
	}
	return key, nil
}

// Encrypt seals plaintext with AES-256-GCM and returns a versioned ciphertext
// containing the random nonce. The caller remains responsible for key storage.
func Encrypt(plaintext, key []byte) (string, error) {
	if len(key) != KeySizeBytes {
		return "", ErrInvalidKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("initialize A-R-S-Y encryption: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("initialize A-R-S-Y authenticated encryption: %w", err)
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate A-R-S-Y encryption nonce: %w", err)
	}
	sealed := aead.Seal(nil, nonce, plaintext, []byte(CiphertextPrefix))
	return strings.Join([]string{
		CiphertextPrefix,
		base64.RawURLEncoding.EncodeToString(nonce),
		base64.RawURLEncoding.EncodeToString(sealed),
	}, ":"), nil
}

// Decrypt opens a versioned AES-256-GCM ciphertext and rejects malformed,
// unsupported, or tampered values.
func Decrypt(value string, key []byte) ([]byte, error) {
	if len(key) != KeySizeBytes {
		return nil, ErrInvalidKey
	}
	parts := strings.Split(value, ":")
	if len(parts) != 4 || parts[0] != "arsy" || parts[1] != "v1" {
		return nil, ErrInvalidCiphertext
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("initialize A-R-S-Y decryption: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("initialize A-R-S-Y authenticated decryption: %w", err)
	}
	nonce, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(nonce) != aead.NonceSize() {
		return nil, ErrInvalidCiphertext
	}
	sealed, err := base64.RawURLEncoding.DecodeString(parts[3])
	if err != nil {
		return nil, ErrInvalidCiphertext
	}
	plaintext, err := aead.Open(nil, nonce, sealed, []byte(CiphertextPrefix))
	if err != nil {
		return nil, ErrAuthenticationFailed
	}
	return plaintext, nil
}

// EncryptWithEnv encrypts using the key stored in the named environment variable.
func EncryptWithEnv(plaintext []byte, envName string) (string, error) {
	key, err := LoadKey(envName)
	if err != nil {
		return "", err
	}
	return Encrypt(plaintext, key)
}

// DecryptWithEnv decrypts using the key stored in the named environment variable.
func DecryptWithEnv(value, envName string) ([]byte, error) {
	key, err := LoadKey(envName)
	if err != nil {
		return nil, err
	}
	return Decrypt(value, key)
}
