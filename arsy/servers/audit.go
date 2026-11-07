package servers

import (
	"encoding/json"

	"github.com/bridge-allah/a-r-s-y/core"
)

// LogAuditEvent creates an immutable audit log entry for credential/key events.
// Soft-fails (logs error but does not block) if audit collection is unavailable.
func LogAuditEvent(app core.App, eventType string, serverID string, changedByID string, metadata map[string]any) error {
	collection, err := app.FindCollectionByNameOrId("arsy_audit_log")
	if err != nil {
		return err
	}

	record := core.NewRecord(collection)
	record.Set("event_type", eventType)
	record.Set("server", serverID)
	if changedByID != "" {
		record.Set("changed_by", changedByID)
	}

	if metadata != nil {
		metadataBytes, _ := json.Marshal(metadata)
		record.Set("metadata", string(metadataBytes))
	}

	return app.Save(record)
}

// GetServerAuditLog retrieves audit events for a server, with optional filtering.
func GetServerAuditLog(app core.App, serverID string, eventType string, limit int, offset int) ([]*core.Record, error) {
	filter := "server = {:serverID}"
	if eventType != "" {
		filter += " && event_type = {:eventType}"
	}

	params := map[string]any{"serverID": serverID}
	if eventType != "" {
		params["eventType"] = eventType
	}

	return app.FindRecordsByFilter(
		"arsy_audit_log",
		filter,
		"-created",
		limit, offset,
		params,
	)
}
