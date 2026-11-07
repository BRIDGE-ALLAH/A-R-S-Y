package servers_test

import (
	"bytes"
	"crypto/ed25519"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/servers"
	"golang.org/x/crypto/ssh"
)

func TestHostKeyFingerprintKnownFixture(t *testing.T) {
	seed := bytes.Repeat([]byte{0x42}, ed25519.SeedSize)
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey, err := ssh.NewPublicKey(privateKey.Public())
	if err != nil {
		t.Fatal(err)
	}
	const want = "SHA256:ZsrOVCtcb1bouzun0GIHz5vL5oCjVhVIQ3jfIBIgZ8g"
	if got := servers.HostKeyFingerprint(publicKey); got != want {
		t.Fatalf("HostKeyFingerprint() = %q, want %q", got, want)
	}
}
