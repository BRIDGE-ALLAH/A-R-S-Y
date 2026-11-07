package servers

import (
	"time"

	"github.com/bridge-allah/a-r-s-y/core"
)

// CreateSession creates a new SSH session record.
func CreateSession(app core.App, sessionID string, serverID string, authMethod string, expiresAt time.Time) error {
	collection, err := app.FindCollectionByNameOrId("arsy_ssh_sessions")
	if err != nil {
		return err
	}

	record := core.NewRecord(collection)
	record.Set("session_id", sessionID)
	record.Set("server", serverID)
	record.Set("auth_method", authMethod)
	if !expiresAt.IsZero() {
		record.Set("expires_at", expiresAt)
	}

	return app.Save(record)
}

// GetSession retrieves a session by ID and checks validity (not invalidated, not expired).
func GetSession(app core.App, sessionID string) (*core.Record, error) {
	record, err := app.FindFirstRecordByData("arsy_ssh_sessions", "session_id", sessionID)
	if err != nil {
		return nil, err
	}

	// Check if invalidated.
	invalidatedAtStr := record.GetDateTime("invalidated_at").String()
	if invalidatedAtStr != "" {
		return nil, ErrSessionInvalidated
	}

	// Check if expired (parse and compare as time.Time).
	expiresAtStr := record.GetDateTime("expires_at").String()
	if expiresAtStr != "" {
		expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
		if err == nil && time.Now().After(expiresAt) {
			return nil, ErrSessionExpired
		}
	}

	return record, nil
}

// InvalidateSessions marks all sessions for a server as invalidated (revoked).
func InvalidateSessions(app core.App, serverID string) error {
	records, err := app.FindRecordsByFilter(
		"arsy_ssh_sessions",
		"server = {:serverID} && invalidated_at = null",
		"",
		1000, 0,
		map[string]any{"serverID": serverID},
	)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, record := range records {
		record.Set("invalidated_at", now)
		if err := app.Save(record); err != nil {
			return err
		}
	}

	return nil
}

// CleanupExpiredSessions deletes sessions invalidated more than 7 days ago.
func CleanupExpiredSessions(app core.App) error {
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	records, err := app.FindRecordsByFilter(
		"arsy_ssh_sessions",
		"invalidated_at != null && invalidated_at < {:cutoff}",
		"",
		10000, 0,
		map[string]any{"cutoff": sevenDaysAgo},
	)
	if err != nil {
		return err
	}

	for _, record := range records {
		if err := app.Delete(record); err != nil {
			return err
		}
	}

	return nil
}
