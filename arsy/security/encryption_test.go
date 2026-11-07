package security

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	want := "sensitive fixture"
	ciphertext, err := Encrypt([]byte(want), key)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(ciphertext, ciphertextPrefix) {
		t.Fatalf("ciphertext missing version prefix: %q", ciphertext)
	}
	if strings.Contains(ciphertext, want) {
		t.Fatal("ciphertext contains plaintext")
	}
	got, err := Decrypt(ciphertext, key)
	if err != nil || string(got) != want {
		t.Fatalf("decrypt result = %q, %v; want %q", got, err, want)
	}
}

func TestDecryptRejectsTamperingAndWrongKey(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	ciphertext, err := Encrypt([]byte("secret"), key)
	if err != nil {
		t.Fatal(err)
	}
	encoded := ciphertext[len(ciphertextPrefix):]
	last := encoded[len(encoded)-1]
	if last == 'A' {
		encoded = encoded[:len(encoded)-1] + "B"
	} else {
		encoded = encoded[:len(encoded)-1] + "A"
	}
	for name, candidate := range map[string]string{
		"tampered":  ciphertextPrefix + encoded,
		"wrong key": ciphertext,
	} {
		t.Run(name, func(t *testing.T) {
			decryptKey := key
			if name == "wrong key" {
				decryptKey = "abcdef0123456789abcdef0123456789"
			}
			if _, err := Decrypt(candidate, decryptKey); err == nil {
				t.Fatal("expected authentication failure")
			}
		})
	}
}

func TestEncryptionRejectsMalformedKeysAndCiphertexts(t *testing.T) {
	for _, key := range []string{"", "short", "0123456789abcdef0123456789abcdeé"} {
		if _, err := Encrypt([]byte("secret"), key); err == nil {
			t.Fatalf("expected malformed key rejection for %q", key)
		}
	}
	if _, err := Decrypt("legacy-or-malformed", "0123456789abcdef0123456789abcdef"); err == nil {
		t.Fatal("expected malformed ciphertext rejection")
	}
}

func TestGenerateKeyProducesValidUniqueKeys(t *testing.T) {
	first, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 64 || first == second {
		t.Fatalf("generated keys must be distinct 256-bit hex strings: lengths %d and %d", len(first), len(second))
	}
	if decoded, err := hex.DecodeString(first); err != nil || len(decoded) != 32 {
		t.Fatalf("generated key is not 32-byte hex: %v", err)
	}
	sealed, err := Encrypt([]byte("generated key fixture"), first)
	if err != nil {
		t.Fatalf("generated key rejected by encryption: %v", err)
	}
	if _, err := Decrypt(sealed, first); err != nil {
		t.Fatalf("generated key rejected by decryption: %v", err)
	}
}
