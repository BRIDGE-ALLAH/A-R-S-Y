package security

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	keyText, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateKey(keyText); err != nil {
		t.Fatal(err)
	}
	key, err := base64.RawStdEncoding.DecodeString(keyText)
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("private SSH key material")
	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(ciphertext, CiphertextPrefix+":") {
		t.Fatalf("ciphertext has no version marker: %q", ciphertext)
	}
	if strings.Contains(ciphertext, string(plaintext)) {
		t.Fatal("ciphertext contains the plaintext")
	}
	got, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(plaintext) {
		t.Fatalf("decrypted value = %q, want %q", got, plaintext)
	}
}

func TestDecryptRejectsTamperingAndWrongKey(t *testing.T) {
	keyText, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	key, _ := base64.RawStdEncoding.DecodeString(keyText)
	otherText, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	other, _ := base64.RawStdEncoding.DecodeString(otherText)

	ciphertext, err := Encrypt([]byte("sensitive"), key)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Decrypt(ciphertext, other); !errors.Is(err, ErrAuthenticationFailed) {
		t.Fatalf("wrong-key error = %v, want authentication failure", err)
	}

	parts := strings.Split(ciphertext, ":")
	sealed, err := base64.RawURLEncoding.DecodeString(parts[3])
	if err != nil {
		t.Fatal(err)
	}
	// Mutate an actual ciphertext byte. Changing only unused trailing base64
	// bits can decode to the original bytes and is not a ciphertext mutation.
	sealed[0] ^= 0x01
	parts[3] = base64.RawURLEncoding.EncodeToString(sealed)
	if _, err := Decrypt(strings.Join(parts, ":"), key); !errors.Is(err, ErrAuthenticationFailed) {
		t.Fatalf("tampered ciphertext error = %v, want authentication failure", err)
	}
}

func TestLoadKeyRejectsMissingAndMalformedValues(t *testing.T) {
	t.Setenv("ARSY_TEST_KEY", "")
	if _, err := LoadKey("ARSY_TEST_KEY"); !errors.Is(err, ErrMissingKey) {
		t.Fatalf("missing key error = %v", err)
	}
	t.Setenv("ARSY_TEST_KEY", "not-a-valid-key")
	if _, err := LoadKey("ARSY_TEST_KEY"); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("invalid key error = %v", err)
	}
	if err := ValidateKey("short"); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("validation error = %v", err)
	}
	if err := ValidateKey(""); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("empty key validation error = %v", err)
	}
}

func TestLoadKeyReadsValidEnvironmentKey(t *testing.T) {
	keyText, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("ARSY_TEST_KEY", keyText)
	key, err := LoadKey("ARSY_TEST_KEY")
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != KeySizeBytes {
		t.Fatalf("key length = %d, want %d", len(key), KeySizeBytes)
	}
}

func TestEncryptWithEnvRejectsMissingKey(t *testing.T) {
	t.Setenv("ARSY_TEST_KEY", "")
	if _, err := EncryptWithEnv([]byte("secret"), "ARSY_TEST_KEY"); !errors.Is(err, ErrMissingKey) {
		t.Fatalf("error = %v, want missing key", err)
	}
}
