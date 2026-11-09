package aeng

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNewAuditLogger verifies the audit logger is created correctly.
func TestNewAuditLogger(t *testing.T) {
	var buf bytes.Buffer
	l := NewAuditLogger(&buf)
	if l == nil {
		t.Fatal("NewAuditLogger() returned nil")
	}
}

// TestAuditLogEntry verifies a structured audit entry is logged correctly.
func TestAuditLogEntry(t *testing.T) {
	var buf bytes.Buffer
	l := NewAuditLogger(&buf)

	entry := AuditEntry{
		CorrelationID: "corr-123",
		RequestID:     "req-456",
		TenantID:      "tenant-abc",
		UserID:        "user-xyz",
		Operation:     "GET /api/companies",
		Resource:      "/api/companies",
		Action:        "GET",
		Status:        "ok",
		DurationMs:    42,
		Metadata: map[string]any{
			"remote_addr": "127.0.0.1",
			"user_agent":  "test-agent",
		},
	}

	if err := l.Log(entry); err != nil {
		t.Fatalf("Log() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "corr-123") {
		t.Fatal("expected correlation_id in log output")
	}
	if !strings.Contains(output, "tenant-abc") {
		t.Fatal("expected tenant_id in log output")
	}
	if !strings.Contains(output, "GET /api/companies") {
		t.Fatal("expected operation in log output")
	}
	if !strings.Contains(output, "\"ok\"") {
		t.Fatal("expected status ok in log output")
	}
}

// TestAuditRedaction verifies sensitive fields are redacted.
func TestAuditRedaction(t *testing.T) {
	var buf bytes.Buffer
	l := NewAuditLogger(&buf).WithRedaction("password", "***REDACTED***").WithRedaction("token", "***REDACTED***")

	entry := AuditEntry{
		CorrelationID: "corr-123",
		Operation:     "POST /api/login",
		Status:        "ok",
		Metadata: map[string]any{
			"password": "secret123",
			"token":    "abc-def-ghi",
			"username": "jdoe",
		},
	}

	if err := l.Log(entry); err != nil {
		t.Fatalf("Log() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "***REDACTED***") {
		t.Fatal("expected sensitive fields to be redacted")
	}
	if strings.Contains(output, "secret123") {
		t.Fatal("expected password to be redacted, found plaintext")
	}
	if strings.Contains(output, "abc-def-ghi") {
		t.Fatal("expected token to be redacted, found plaintext")
	}
	if !strings.Contains(output, "jdoe") {
		t.Fatal("expected non-sensitive field to remain visible")
	}
}

// TestAuditMiddlewareWrap verifies middleware logs requests and responses.
func TestAuditMiddlewareWrap(t *testing.T) {
	var buf bytes.Buffer
	logger := NewAuditLogger(&buf)
	middleware := NewAuditMiddleware(logger)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		w.Write([]byte("created"))
	})

	wrapped := middleware.Wrap(next)

	req := httptest.NewRequest("POST", "/api/companies", strings.NewReader(`{"name":"Acme"}`))
	req.Header.Set("X-Correlation-ID", "corr-999")
	req.Header.Set("X-Tenant-ID", "tenant-1")
	req.Header.Set("X-User-ID", "user-2")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != 201 {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	output := buf.String()
	if !strings.Contains(output, "corr-999") {
		t.Fatal("expected correlation ID in middleware log")
	}
	if !strings.Contains(output, "tenant-1") {
		t.Fatal("expected tenant ID in middleware log")
	}
	if !strings.Contains(output, "/api/companies") {
		t.Fatal("expected resource path in middleware log")
	}
	if !strings.Contains(output, "created") && !strings.Contains(output, "201") {
		t.Fatal("expected response status in middleware log")
	}
}

// TestAuditMiddlewareMissingHeaders verifies middleware generates IDs when missing.
func TestAuditMiddlewareMissingHeaders(t *testing.T) {
	var buf bytes.Buffer
	logger := NewAuditLogger(&buf)
	middleware := NewAuditMiddleware(logger)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	wrapped := middleware.Wrap(next)

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	output := buf.String()
	if !strings.Contains(output, "correlation_id") {
		t.Fatal("expected generated correlation_id when header missing")
	}
}

// TestAuditContextPropagation verifies context key functions work.
func TestAuditContextPropagation(t *testing.T) {
	ctx := context.Background()

	// Set correlation ID
	ctx = WithCorrelationID(ctx, "corr-test")
	id, ok := CorrelationIDFromContext(ctx)
	if !ok || id != "corr-test" {
		t.Fatal("expected correlation ID to be retrievable from context")
	}

	// Set tenant ID
	ctx = WithTenantID(ctx, "tenant-test")
	tid, ok := TenantIDFromContext(ctx)
	if !ok || tid != "tenant-test" {
		t.Fatal("expected tenant ID to be retrievable from context")
	}
}

// TestNewCorrelationID verifies unique correlation IDs are generated.
func TestNewCorrelationID(t *testing.T) {
	id1 := NewCorrelationID()
	id2 := NewCorrelationID()

	if id1 == "" {
		t.Fatal("expected non-empty correlation ID")
	}
	if id1 == id2 {
		t.Fatal("expected unique correlation IDs")
	}

	// Verify UUID format (36 chars with hyphens)
	if len(id1) != 36 {
		t.Fatalf("expected UUID v4 format (36 chars), got %d: %s", len(id1), id1)
	}
}

// TestAuditLoggerNilWriter verifies behavior with nil writer.
func TestAuditLoggerNilWriter(t *testing.T) {
	l := NewAuditLogger(nil)
	entry := AuditEntry{
		CorrelationID: "corr-1",
		Operation:     "GET /test",
		Status:        "ok",
	}

	// Should not panic with nil writer
	if err := l.Log(entry); err != nil {
		// Expected - nil writer will fail to encode
		t.Logf("expected error with nil writer: %v", err)
	}
}

// TestAuditJSONStructure verifies the log output is valid JSON.
func TestAuditJSONStructure(t *testing.T) {
	var buf bytes.Buffer
	l := NewAuditLogger(&buf)

	entry := AuditEntry{
		CorrelationID: "corr-json",
		RequestID:     "req-json",
		TenantID:      "tenant-json",
		UserID:        "user-json",
		Operation:     "GET /api/test",
		Resource:      "/api/test",
		Action:        "GET",
		Status:        "ok",
		DurationMs:    10,
	}

	if err := l.Log(entry); err != nil {
		t.Fatalf("Log() error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v", err)
	}

	if decoded["correlation_id"] != "corr-json" {
		t.Fatalf("expected correlation_id corr-json, got %v", decoded["correlation_id"])
	}
	if decoded["tenant_id"] != "tenant-json" {
		t.Fatalf("expected tenant_id tenant-json, got %v", decoded["tenant_id"])
	}
}

// TestAuditQueryString verifies LogQL query string generation.
func TestAuditQueryString(t *testing.T) {
	q := &AuditQuery{
		CorrelationID: "corr-123",
		TenantID:      "tenant-abc",
		Operation:     "GET /api",
		Limit:         100,
	}

	query := q.String()
	if !strings.Contains(query, `correlation_id="corr-123"`) {
		t.Fatal("expected correlation_id filter in query")
	}
	if !strings.Contains(query, `tenant_id="tenant-abc"`) {
		t.Fatal("expected tenant_id filter in query")
	}
	if !strings.Contains(query, "limit 100") {
		t.Fatal("expected limit in query")
	}
	if !strings.Contains(query, `{job="aeng-audit"}`) {
		t.Fatal("expected aeng-audit job selector in query")
	}
}

// TestAuditQueryEmpty verifies empty query string.
func TestAuditQueryEmpty(t *testing.T) {
	q := &AuditQuery{}
	query := q.String()
	if query != `{job="aeng-audit"}` {
		t.Fatalf("expected base query, got %s", query)
	}
}

// TestAuditLoggerFlush verifies flush writes buffered entries.
func TestAuditLoggerFlush(t *testing.T) {
	var buf bytes.Buffer
	l := NewAuditLogger(&buf)
	l.bufferMax = 2 // Force flush after 2 entries

	entry1 := AuditEntry{CorrelationID: "corr-1", Operation: "GET /a", Status: "ok"}

	if err := l.Log(entry1); err != nil {
		t.Fatalf("first Log() error = %v", err)
	}
	// Second log should trigger automatic flush
	entry2 := AuditEntry{CorrelationID: "corr-2", Operation: "GET /b", Status: "ok"}
	if err := l.Log(entry2); err != nil {
		t.Fatalf("second Log() error = %v", err)
	}

	if err := l.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "corr-1") {
		t.Fatal("expected corr-1 in flushed output")
	}
	if !strings.Contains(output, "corr-2") {
		t.Fatal("expected corr-2 in flushed output")
	}
}
