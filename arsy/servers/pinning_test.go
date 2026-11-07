package servers

import (
	"testing"

	"github.com/bridge-allah/a-r-s-y/core"
)

func TestConfirmHostKeyPinsOnlyObservedFingerprint(t *testing.T) {
	record := core.NewRecord(core.NewBaseCollection("servers"))
	record.Set("pending_host_key_fingerprint", "SHA256:observed")
	if err := ConfirmHostKey(record, "SHA256:other"); err != ErrHostKeyMismatch {
		t.Fatalf("mismatched confirmation error = %v", err)
	}
	if record.GetString("host_key_fingerprint") != "" {
		t.Fatal("mismatched fingerprint was pinned")
	}
	if err := ConfirmHostKey(record, "SHA256:observed"); err != nil {
		t.Fatal(err)
	}
	if record.GetString("host_key_fingerprint") != "SHA256:observed" || record.GetString("pending_host_key_fingerprint") != "" || record.GetString("readiness_state") != "pending" {
		t.Fatal("confirmed fingerprint was not promoted to pinned state")
	}
}

func TestConfirmHostKeyRejectsMissingPendingFingerprint(t *testing.T) {
	record := core.NewRecord(core.NewBaseCollection("servers"))
	if err := ConfirmHostKey(record, "SHA256:any"); err == nil {
		t.Fatal("confirmation without a pending fingerprint was accepted")
	}
}
