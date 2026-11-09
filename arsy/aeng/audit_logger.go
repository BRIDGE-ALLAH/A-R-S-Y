package aeng

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/bridge-allah/a-r-s-y/core"
)

// AuditLogEntry represents a structured audit log entry
type AuditLogEntry struct {
	EventID        string    `json:"event_id"`
	Timestamp      time.Time   `json:"timestamp"`
	ServerID       string    `json:"server_id"`
	ChangedByID    string    `json:"changed_by_id,omitempty"`
	CorrelationID  string    `json:"correlation_id"`
	TenantID       string    `json:"tenant_id"`
	UserID         string    `json:"user_id"`
	Operation      string    `json:"operation"`
	Resource       string    `json:"resource"`
	Status         int         `json:"status_code"`
	Duration       int64       `json:"duration_ms"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
	RedactedFields map[string]interface{} `json:"redacted_fields,omitempty"`
}

type AsyncAuditLogger struct {
	app           core.App
	collection    *core.Collection
	mu            sync.RWMutex
	queue         chan *AuditLogEntry
	isClosed      bool
	wg            sync.WaitGroup
}

// NewAsyncAuditLogger creates a new async audit logger
func NewAsyncAuditLogger(app core.App) (*AsyncAuditLogger, error) {
	collection, err := app.FindCollectionByNameOrId("arsy_audit_log")
	if err != nil {
		return nil, err
	}

	logger := &AsyncAuditLogger{
		app:           app,
		collection:    collection,
		queue:         make(chan *AuditLogEntry, 100),
		isClosed:      false,
	}

	// Start processing loop
	logger.wg.Add(1)
	go logger.processLoop()

	return logger, nil
}

// Log creates an audit log entry
func (l *AsyncAuditLogger) Log(entry *AuditLogEntry) {
	if l.isClosed {
		return
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	select {
	case l.queue <- entry:
		// Successfully queued
	default:
		// Queue full, drop the entry
		return
	}
}

// processLoop processes audit log entries asynchronously
func (l *AsyncAuditLogger) processLoop() {
	defer l.wg.Done()

	for entry := range l.queue {
		if err := l.saveEntry(entry); err != nil {
			fmt.Printf("Audit log save error: %v\n", err)
		}
	}

	// Wait for any pending entries to be processed
	l.wg.Wait()
}

// saveEntry saves a single audit log entry
func (l *AsyncAuditLogger) saveEntry(entry *AuditLogEntry) error {
	// Apply redaction if needed
	if entry.NeedsRedaction() {
		entry.Metadata = l.redactFields(entry.Metadata)
	}

	// For now, just log to stdout; actual persistence can be implemented later
	fmt.Printf("Audit entry: %+v\n", entry)
	return nil
}

// NeedsRedaction determines if the entry needs field redaction
func (e *AuditLogEntry) NeedsRedaction() bool {
	if e.Metadata == nil {
		return false
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(e.Metadata, &raw); err != nil {
		return false
	}

	for key := range raw {
		if key == "password" || key == "token" || key == "ssn" || key == "secret" || key == "private_key" || key == "api_key" {
			return true
		}
	}
	return false
}

// redactFields performs field redaction on metadata
func (l *AsyncAuditLogger) redactFields(data json.RawMessage) json.RawMessage {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return data
	}

	redacted := make(map[string]interface{})
	for k, v := range raw {
		switch k {
		case "password", "token", "ssn", "secret", "private_key", "api_key":
			redacted[k] = "[REDACTED]"
		default:
			redacted[k] = v
		}
	}

	return json.Marshal(redacted)
}

// Close shuts down the async audit logger
func (l *AsyncAuditLogger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.isClosed {
		return
	}

	l.isClosed = true
	close(l.queue)
	l.wg.Wait()
}