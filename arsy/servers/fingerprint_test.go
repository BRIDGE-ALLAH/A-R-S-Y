package servers

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestHostKeyFingerprintUsesOpenSSHFormat(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	sshKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		t.Fatal(err)
	}
	got, err := HostKeyFingerprint(sshKey)
	if err != nil {
		t.Fatal(err)
	}
	want := ssh.FingerprintSHA256(sshKey)
	if got != want || len(got) < len("SHA256:") || got[:len("SHA256:")] != "SHA256:" {
		t.Fatalf("fingerprint = %q, want %q", got, want)
	}
}

func TestHostKeyFingerprintRejectsNilKey(t *testing.T) {
	if _, err := HostKeyFingerprint(nil); err == nil {
		t.Fatal("expected nil host key to be rejected")
	}
}
