package aeng

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bridge-allah/a-r-s-y/core"
)

const (
	auditQueueSize     = 100
	auditEventTypeMax  = 64
	auditChangedByMax  = 80
	auditEventFallback = "audit_event"
)

var sensitiveAuditFields = map[string]struct{}{
	"api_key":       {},
	"authorization": {},
	"client_secret": {},
	"password":      {},
	"private_key":   {},
	"refresh_token": {},
	"secret":        {},
	"ssn":           {},
	"token":         {},
}

// AuditLogEntry represents a structured audit log entry
type AuditLogEntry struct {
	EventID        string                 `json:"event_id"`
	Timestamp      time.Time              `json:"timestamp"`
	ServerID       string                 `json:"server_id"`
	ChangedByID    string                 `json:"changed_by_id,omitempty"`
	CorrelationID  string                 `json:"correlation_id"`
	TenantID       string                 `json:"tenant_id"`
	UserID         string                 `json:"user_id"`
	Operation      string                 `json:"operation"`
	Resource       string                 `json:"resource"`
	Status         int                    `json:"status_code"`
	Duration       int64                  `json:"duration_ms"`
	Metadata       json.RawMessage        `json:"metadata,omitempty"`
	RedactedFields map[string]interface{} `json:"redacted_fields,omitempty"`
}

type AuditModel struct {
	Data        map[string]interface{}
	lastSavedPK string
	Id          string `db:"id" json:"id" form:"id" xml:"id"`
}

// DBExport implements core.DBExporter.
func (am *AuditModel) DBExport(app core.App) (map[string]interface{}, error) {
	data := make(map[string]interface{}, len(am.Data)+1)
	for key, value := range am.Data {
		data[key] = value
	}
	data["id"] = am.Id
	return data, nil
}

func (am *AuditModel) TableName() string {
	return "arsy_audit_log"
}

func (am *AuditModel) PK() any {
	return am.Id
}

func (am *AuditModel) LastSavedPK() any {
	return am.lastSavedPK
}

func (am *AuditModel) IsNew() bool {
	return am.lastSavedPK == ""
}

func (am *AuditModel) MarkAsNew() {
	am.lastSavedPK = ""
}

func (am *AuditModel) MarkAsNotNew() {
	am.lastSavedPK = am.Id
}

type AsyncAuditLogger struct {
	app        core.App
	collection *core.Collection
	saveFn     func(core.Model) error
	mu         sync.RWMutex
	queue      chan *AuditLogEntry
	isClosed   bool
	closeOnce  sync.Once
	wg         sync.WaitGroup
	done       chan struct{}
}

// NewAsyncAuditLogger creates a new async audit logger
func NewAsyncAuditLogger(app core.App) (*AsyncAuditLogger, error) {
	collection, err := app.FindCollectionByNameOrId("arsy_audit_log")
	if err != nil {
		return nil, err
	}

	logger := &AsyncAuditLogger{
		app:        app,
		collection: collection,
		saveFn:     app.Save,
		queue:      make(chan *AuditLogEntry, auditQueueSize),
		isClosed:   false,
		done:       make(chan struct{}),
	}

	// Start processing loop
	logger.wg.Add(1)
	go logger.processLoop()

	return logger, nil
}

// Log creates an audit log entry
func (l *AsyncAuditLogger) Log(entry *AuditLogEntry) {
	if entry == nil {
		return
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.isClosed {
		return
	}

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
	defer close(l.done)

	for entry := range l.queue {
		if err := l.saveEntry(entry); err != nil {
			fmt.Printf("Audit log save error: %v\n", err)
		}
	}
}

// saveEntry saves a single audit log entry
func (l *AsyncAuditLogger) saveEntry(entry *AuditLogEntry) error {
	// Keep caller-owned metadata immutable.
	metadata := entry.Metadata
	if entry.NeedsRedaction() {
		metadata = l.redactFields(metadata)
	}

	eventType := entry.Operation
	if eventType == "" {
		eventType = auditEventFallback
	}
	eventType = truncate(eventType, auditEventTypeMax)

	changedBy := entry.ChangedByID
	if changedBy == "" {
		changedBy = entry.UserID
	}

	eventID := entry.EventID
	if eventID == "" {
		eventID = core.GenerateDefaultRandomId()
	}
	timestamp := entry.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}

	payload := map[string]interface{}{
		"event_id":       eventID,
		"timestamp":      timestamp.UTC().Format(time.RFC3339Nano),
		"correlation_id": entry.CorrelationID,
		"tenant_id":      entry.TenantID,
		"user_id":        entry.UserID,
		"operation":      entry.Operation,
		"resource":       entry.Resource,
		"status_code":    entry.Status,
		"duration_ms":    entry.Duration,
	}
	if len(metadata) > 0 {
		var m map[string]interface{}
		if err := json.Unmarshal(metadata, &m); err != nil {
			return fmt.Errorf("decode audit metadata: %w", err)
		}
		for key, value := range m {
			payload[key] = value
		}
	}

	// Create a model that implements core.Model
	model := &AuditModel{
		Id:          core.GenerateDefaultRandomId(),
		lastSavedPK: "",
		Data: map[string]interface{}{
			"event_type": eventType,
			"changed_by": truncate(changedBy, auditChangedByMax),
			"metadata":   payload,
		},
	}
	if entry.ServerID != "" {
		model.Data["server"] = entry.ServerID
	}

	// Save the record
	if err := l.saveFn(model); err != nil {
		return err
	}

	return nil
}

// NeedsRedaction determines if the entry needs field redaction
func (e *AuditLogEntry) NeedsRedaction() bool {
	if e.Metadata == nil {
		return false
	}

	var raw interface{}
	if err := json.Unmarshal(e.Metadata, &raw); err != nil {
		return false
	}

	return containsSensitiveField(raw)
}

// redactFields performs field redaction on metadata
func (l *AsyncAuditLogger) redactFields(data json.RawMessage) json.RawMessage {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return data
	}

	redacted := redactSensitiveValue(raw)

	marshaled, err := json.Marshal(redacted)
	if err != nil {
		// In case of error, return the original data
		return data
	}

	return marshaled
}

func containsSensitiveField(value interface{}) bool {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, nested := range typed {
			if _, sensitive := sensitiveAuditFields[strings.ToLower(key)]; sensitive {
				return true
			}
			if containsSensitiveField(nested) {
				return true
			}
		}
	case []interface{}:
		for _, nested := range typed {
			if containsSensitiveField(nested) {
				return true
			}
		}
	}
	return false
}

func redactSensitiveValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		redacted := make(map[string]interface{}, len(typed))
		for key, nested := range typed {
			if _, sensitive := sensitiveAuditFields[strings.ToLower(key)]; sensitive {
				redacted[key] = "[REDACTED]"
				continue
			}
			redacted[key] = redactSensitiveValue(nested)
		}
		return redacted
	case []interface{}:
		redacted := make([]interface{}, len(typed))
		for index, nested := range typed {
			redacted[index] = redactSensitiveValue(nested)
		}
		return redacted
	default:
		return value
	}
}

// Close shuts down the async audit logger
func (l *AsyncAuditLogger) Close() {
	l.closeOnce.Do(func() {
		l.mu.Lock()
		l.isClosed = true
		close(l.queue)
		l.mu.Unlock()
	})
	<-l.done
}

func truncate(value string, maxRunes int) string {
	if len(value) <= maxRunes {
		return value
	}
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return string(runes[:maxRunes])
}
