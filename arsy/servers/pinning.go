package servers

import (
	"errors"
	"strings"

	"github.com/bridge-allah/a-r-s-y/core"
)

// ConfirmHostKey pins only the exact fingerprint captured by the most recent
// unauthenticated probe. The caller must require authenticated administrator
// confirmation before invoking this helper.
func ConfirmHostKey(record *core.Record, confirmedFingerprint string) error {
	if record == nil {
		return errors.New("server record is required")
	}
	pending := strings.TrimSpace(record.GetString("pending_host_key_fingerprint"))
	if pending == "" {
		return errors.New("no pending SSH host-key fingerprint is available")
	}
	if strings.TrimSpace(confirmedFingerprint) != pending {
		return ErrHostKeyMismatch
	}
	record.Set("host_key_fingerprint", pending)
	record.Set("pending_host_key_fingerprint", "")
	record.Set("readiness_state", "pending")
	return nil
}
