package servers

import (
	"errors"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestCheckHostKeyRequiresExplicitPin(t *testing.T) {
	key := testSSHHostPublicKey(t)
	fingerprint, trusted, err := CheckHostKey("", key)
	if err != nil {
		t.Fatal(err)
	}
	if fingerprint == "" || trusted {
		t.Fatalf("first-seen host key = (%q, %v), want fingerprint and untrusted", fingerprint, trusted)
	}
}

func TestCheckHostKeyAcceptsMatchingPin(t *testing.T) {
	key := testSSHHostPublicKey(t)
	expected := ssh.FingerprintSHA256(key)
	fingerprint, trusted, err := CheckHostKey(expected, key)
	if err != nil {
		t.Fatal(err)
	}
	if fingerprint != expected || !trusted {
		t.Fatalf("matching host key = (%q, %v), want (%q, true)", fingerprint, trusted, expected)
	}
}

func TestCheckHostKeyRejectsChangedPin(t *testing.T) {
	key := testSSHHostPublicKey(t)
	fingerprint, trusted, err := CheckHostKey("SHA256:wrong", key)
	if !errors.Is(err, ErrHostKeyMismatch) {
		t.Fatalf("error = %v, want host key mismatch", err)
	}
	if fingerprint == "" || trusted {
		t.Fatalf("mismatched host key = (%q, %v), want observed fingerprint and untrusted", fingerprint, trusted)
	}
}

func TestPinnedHostKeyCallbackRejectsUnpinnedAndChangedKeys(t *testing.T) {
	key := testSSHHostPublicKey(t)
	if err := PinnedHostKeyCallback("")("host", nil, key); !errors.Is(err, ErrHostKeyNotPinned) {
		t.Fatalf("un-pinned key error = %v, want ErrHostKeyNotPinned", err)
	}
	if err := PinnedHostKeyCallback("SHA256:wrong")("host", nil, key); !errors.Is(err, ErrHostKeyMismatch) {
		t.Fatalf("changed key error = %v, want ErrHostKeyMismatch", err)
	}
	if err := PinnedHostKeyCallback(ssh.FingerprintSHA256(key))("host", nil, key); err != nil {
		t.Fatalf("matching pinned key rejected: %v", err)
	}
}

func testSSHHostPublicKey(t *testing.T) ssh.PublicKey {
	t.Helper()
	privateKey := testOpenSSHPrivateKey(t)
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	return signer.PublicKey()
}
