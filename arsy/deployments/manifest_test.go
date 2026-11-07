package deployments

import (
	"testing"
)

func TestEncryptManifestRejectsEmpty(t *testing.T) {
	if _, err := EncryptManifest("", "TEST_ENV"); err == nil {
		t.Fatal("expected error for empty manifest")
	}
}

func TestDecryptManifestPassesThroughPlaintext(t *testing.T) {
	plain := "services:\n  web:\n    image: nginx\n"
	result, err := DecryptManifest(plain, "TEST_ENV")
	if err != nil || result != plain {
		t.Fatalf("DecryptManifest() passthrough failed: err=%v, got=%q", err, result)
	}
}

func TestDecryptManifestRejectsEmptyValue(t *testing.T) {
	if _, err := DecryptManifest("", "TEST_ENV"); err == nil {
		t.Fatal("expected error for empty ciphertext")
	}
}
