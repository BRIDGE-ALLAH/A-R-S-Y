package servers_test

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/servers"
	"golang.org/x/crypto/ssh"
)

func TestRequirePinnedHostKey(t *testing.T) {
	first := fixtureHostKey(t, 0x42)
	second := fixtureHostKey(t, 0x24)
	fingerprint := servers.HostKeyFingerprint(first)

	err := servers.RequirePinnedHostKey("", first)
	var unpinned *servers.UnpinnedHostKeyError
	if !errors.As(err, &unpinned) || unpinned.Fingerprint != fingerprint {
		t.Fatalf("first-seen host key error = %#v, want candidate fingerprint %q", err, fingerprint)
	}
	if err := servers.RequirePinnedHostKey(fingerprint, first); err != nil {
		t.Fatalf("matching pinned host key rejected: %v", err)
	}
	if err := servers.RequirePinnedHostKey(fingerprint, second); err == nil {
		t.Fatal("changed host key was accepted")
	}
	if err := servers.RequirePinnedHostKey(fingerprint, nil); err == nil {
		t.Fatal("missing host key was accepted")
	}
}

func TestHostKeyCallbackRequiresPin(t *testing.T) {
	key := fixtureHostKey(t, 0x42)
	callback := servers.HostKeyCallback(servers.HostKeyFingerprint(key))
	if err := callback("example.com:22", nil, key); err != nil {
		t.Fatalf("matching callback failed: %v", err)
	}
	if err := servers.HostKeyCallback("")("example.com:22", nil, key); err == nil {
		t.Fatal("callback accepted an unpinned host key")
	}
}

func fixtureHostKey(t *testing.T, fill byte) ssh.PublicKey {
	t.Helper()
	seed := bytes.Repeat([]byte{fill}, ed25519.SeedSize)
	key, err := ssh.NewPublicKey(ed25519.NewKeyFromSeed(seed).Public())
	if err != nil {
		t.Fatal(err)
	}
	return key
}
