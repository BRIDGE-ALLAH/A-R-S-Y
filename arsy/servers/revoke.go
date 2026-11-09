package servers

import (
	"fmt"
	"time"

	"github.com/bridge-allah/a-r-s-y/core"
)

// RevokeServerHostKey marks a server's host key as revoked, logs the event, and invalidates all sessions.
func RevokeServerHostKey(app core.App, serverID string, changedByID string) error {
	record, err := app.FindRecordById("arsy_servers", serverID)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Check if already revoked (idempotent).
	if record.GetBool("host_key_revoked") {
		return nil
	}

	// Mark as revoked.
	record.Set("host_key_revoked", true)
	if err := app.Save(record); err != nil {
		return err
	}

	// Log revocation event.
	_ = LogAuditEvent(app, "host_key_revoked", serverID, changedByID, map[string]any{
		"revoked_at": time.Now().UTC().Format(time.RFC3339),
	})

	// Invalidate all active sessions for this server.
	return InvalidateSessions(app, serverID)
}
