package aeng

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/bridge-allah/a-r-s-y/core"
)

func TestAuditLogEntryNeedsRedaction(t *testing.T) {
	tests := []struct {
		name     string
		metadata json.RawMessage
		want     bool
	}{
		{name: "nil metadata", metadata: nil, want: false},
		{name: "invalid metadata", metadata: json.RawMessage(`{`), want: false},
		{name: "safe metadata", metadata: json.RawMessage(`{"username":"jdoe"}`), want: false},
		{name: "password", metadata: json.RawMessage(`{"password":"secret"}`), want: true},
		{name: "api key", metadata: json.RawMessage(`{"api_key":"key"}`), want: true},
		{name: "nested sensitive key", metadata: json.RawMessage(`{"auth":{"token":"value"}}`), want: true},
		{name: "array sensitive key", metadata: json.RawMessage(`[{"token":"value"}]`), want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := AuditLogEntry{Metadata: test.metadata}
			if got := entry.NeedsRedaction(); got != test.want {
				t.Fatalf("NeedsRedaction() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestAsyncAuditLoggerRedactsWithoutMutatingEntry(t *testing.T) {
	var (
		mu    sync.Mutex
		saved *AuditModel
	)
	logger := newTestAsyncAuditLogger(func(model core.Model) error {
		mu.Lock()
		defer mu.Unlock()
		saved = model.(*AuditModel)
		return nil
	})

	entry := &AuditLogEntry{
		EventID:       "event-1",
		Timestamp:     time.Date(2026, 11, 9, 9, 0, 0, 0, time.UTC),
		ServerID:      "server-1",
		ChangedByID:   "user-1",
		CorrelationID: "correlation-1",
		TenantID:      "tenant-1",
		UserID:        "user-1",
		Operation:     "server.credentials.rotate",
		Resource:      "server/server-1",
		Status:        200,
		Duration:      12,
		Metadata:      json.RawMessage(`{"password":"secret","auth":{"token":"bearer"},"region":"jakarta"}`),
	}

	logger.Log(entry)
	logger.Close()

	mu.Lock()
	defer mu.Unlock()
	if saved == nil {
		t.Fatal("expected audit entry to be saved")
	}

	metadata, ok := saved.Data["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("metadata has type %T, want map[string]interface{}", saved.Data["metadata"])
	}
	auth, ok := metadata["auth"].(map[string]interface{})
	if !ok {
		t.Fatalf("nested metadata has type %T, want map[string]interface{}", metadata["auth"])
	}
	if metadata["password"] != "[REDACTED]" || auth["token"] != "[REDACTED]" {
		t.Fatalf("sensitive metadata was not redacted: %#v", metadata)
	}
	if metadata["region"] != "jakarta" {
		t.Fatalf("safe metadata was dropped: %#v", metadata)
	}
	if metadata["event_id"] != "event-1" || metadata["correlation_id"] != "correlation-1" {
		t.Fatalf("audit envelope is incomplete: %#v", metadata)
	}
	if string(entry.Metadata) != `{"password":"secret","auth":{"token":"bearer"},"region":"jakarta"}` {
		t.Fatalf("caller metadata was mutated: %s", entry.Metadata)
	}
}

func TestAsyncAuditLoggerCloseIsIdempotentAndDrainsQueue(t *testing.T) {
	var (
		mu    sync.Mutex
		saved int
	)

	logger := newTestAsyncAuditLogger(func(core.Model) error {
		mu.Lock()
		saved++
		mu.Unlock()
		return nil
	})

	for index := 0; index < auditQueueSize; index++ {
		logger.Log(&AuditLogEntry{EventID: "queued"})
	}

	done := make(chan struct{})
	go func() {
		logger.Close()
		logger.Close()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Close did not return")
	}

	mu.Lock()
	defer mu.Unlock()
	if saved != auditQueueSize {
		t.Fatalf("saved %d entries, want %d", saved, auditQueueSize)
	}
}

func TestAuditModelDBExport(t *testing.T) {
	model := &AuditModel{
		Id:   "audit-id",
		Data: map[string]interface{}{"event_type": "server.created"},
	}

	exported, err := model.DBExport(nil)
	if err != nil {
		t.Fatalf("DBExport() error = %v", err)
	}
	if exported["id"] != "audit-id" {
		t.Fatalf("exported id = %v, want audit-id", exported["id"])
	}
	if exported["event_type"] != "server.created" {
		t.Fatalf("exported event_type = %v, want server.created", exported["event_type"])
	}
}

func newTestAsyncAuditLogger(save func(core.Model) error) *AsyncAuditLogger {
	logger := &AsyncAuditLogger{
		saveFn: save,
		queue:  make(chan *AuditLogEntry, auditQueueSize),
		done:   make(chan struct{}),
	}
	logger.wg.Add(1)
	go logger.processLoop()
	return logger
}
