// Package aeng implements the Application Engine framework.
// AENG-045: Enterprise-Grade Audit Logging System.
package aeng

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AuditEntry represents a single structured audit log event.
type AuditEntry struct {
	Timestamp     string         `json:"timestamp"`
	CorrelationID string         `json:"correlation_id"`
	RequestID     string         `json:"request_id"`
	TenantID      string         `json:"tenant_id"`
	UserID        string         `json:"user_id"`
	Operation     string         `json:"operation"`
	Resource      string         `json:"resource"`
	Action        string         `json:"action"`
	Status        string         `json:"status"`
	DurationMs    int64          `json:"duration_ms"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

// RedactionRule defines a field redaction pattern for sensitive data.
type RedactionRule struct {
	Field   string
	Pattern string
}

// AuditLogger writes structured audit entries to an output writer.
type AuditLogger struct {
	mu        sync.Mutex
	writer    io.Writer
	encoder   *json.Encoder
	rules     []RedactionRule
	buffer    []AuditEntry
	bufferMax int
	flushSec  int
}

// NewAuditLogger creates an audit logger writing to the given writer.
// A nil writer disables logging (circuit breaker) but never panics.
func NewAuditLogger(w io.Writer) *AuditLogger {
	var enc *json.Encoder
	if w != nil {
		enc = json.NewEncoder(w)
	}
	return &AuditLogger{
		writer:    w,
		encoder:   enc,
		bufferMax: 1000,
		flushSec:  5,
	}
}

// WithRedaction adds a redaction rule for sensitive fields.
func (l *AuditLogger) WithRedaction(field, pattern string) *AuditLogger {
	l.rules = append(l.rules, RedactionRule{Field: field, Pattern: pattern})
	return l
}

// Log writes a structured audit entry.
// If the writer is nil (circuit breaker), the entry is silently dropped.
func (l *AuditLogger) Log(entry AuditEntry) error {
	if l.encoder == nil {
		return nil
	}

	entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	entry.Metadata = l.redact(entry.Metadata)

	l.mu.Lock()
	defer l.mu.Unlock()

	l.buffer = append(l.buffer, entry)
	if len(l.buffer) >= l.bufferMax {
		return l.flushLocked()
	}

	return l.encoder.Encode(entry)
}

// flushLocked writes all buffered entries to the writer.
func (l *AuditLogger) flushLocked() error {
	if l.encoder == nil {
		l.buffer = l.buffer[:0]
		return nil
	}
	for _, entry := range l.buffer {
		if err := l.encoder.Encode(entry); err != nil {
			return err
		}
	}
	l.buffer = l.buffer[:0]
	return nil
}

// Flush forces writing any buffered entries.
func (l *AuditLogger) Flush() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.flushLocked()
}

// redact applies redaction rules to metadata map.
func (l *AuditLogger) redact(metadata map[string]any) map[string]any {
	if len(l.rules) == 0 || len(metadata) == 0 {
		return metadata
	}

	for _, rule := range l.rules {
		if _, ok := metadata[rule.Field]; ok {
			metadata[rule.Field] = rule.Pattern
		}
	}
	return metadata
}

// AuditMiddleware wraps an HTTP handler to automatically log requests/responses.
type AuditMiddleware struct {
	logger *AuditLogger
}

// NewAuditMiddleware creates audit middleware backed by an audit logger.
func NewAuditMiddleware(logger *AuditLogger) *AuditMiddleware {
	return &AuditMiddleware{logger: logger}
}

// Wrap wraps an http.Handler with audit logging.
func (m *AuditMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = NewCorrelationID()
		}

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = NewCorrelationID()
		}

		tenantID := r.Header.Get("X-Tenant-ID")
		userID := r.Header.Get("X-User-ID")

		entry := AuditEntry{
			CorrelationID: correlationID,
			RequestID:     requestID,
			TenantID:      tenantID,
			UserID:        userID,
			Operation:     r.Method + " " + r.URL.Path,
			Resource:      r.URL.Path,
			Action:        r.Method,
			Status:        "pending",
			Metadata: map[string]any{
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
			},
		}

		// Log request start
		if err := m.logger.Log(entry); err != nil {
			// Audit log failures must not block request processing
			_ = err
		}

		// Wrap response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: 200}
		next.ServeHTTP(rw, r)

		durationMs := time.Since(start).Milliseconds()
		entry.Status = httpStatusText(rw.statusCode)
		entry.DurationMs = durationMs
		entry.Metadata["status_code"] = rw.statusCode

		// Log response
		if err := m.logger.Log(entry); err != nil {
			_ = err
		}
	})
}

// responseWriter captures the HTTP status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func httpStatusText(code int) string {
	switch code {
	case 200:
		return "ok"
	case 201:
		return "created"
	case 204:
		return "no_content"
	case 400:
		return "bad_request"
	case 401:
		return "unauthorized"
	case 403:
		return "forbidden"
	case 404:
		return "not_found"
	case 500:
		return "internal_error"
	default:
		return fmt.Sprintf("status_%d", code)
	}
}

// AuditContextKey is the context key type for audit context propagation.
type AuditContextKey string

const (
	// CorrelationIDKey is the context key for correlation ID.
	CorrelationIDKey AuditContextKey = "correlation_id"
	// TenantIDKey is the context key for tenant ID.
	TenantIDKey AuditContextKey = "tenant_id"
	// UserIDKey is the context key for user ID.
	UserIDKey AuditContextKey = "user_id"
)

// WithCorrelationID returns a context with the given correlation ID.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, id)
}

// CorrelationIDFromContext extracts the correlation ID from context.
func CorrelationIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(CorrelationIDKey).(string)
	return id, ok
}

// WithTenantID returns a context with the given tenant ID.
func WithTenantID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, TenantIDKey, id)
}

// TenantIDFromContext extracts the tenant ID from context.
func TenantIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(TenantIDKey).(string)
	return id, ok
}

// NewCorrelationID generates a new UUID v4 correlation ID.
func NewCorrelationID() string {
	return uuid.New().String()
}

// AuditQuery builds LogQL queries for searching audit entries.
type AuditQuery struct {
	CorrelationID string
	TenantID      string
	Operation     string
	StartTime     time.Time
	EndTime       time.Time
	Limit         int
}

// String renders the audit query as a LogQL query string.
func (q *AuditQuery) String() string {
	var filters []string
	if q.CorrelationID != "" {
		filters = append(filters, fmt.Sprintf(`correlation_id="%s"`, q.CorrelationID))
	}
	if q.TenantID != "" {
		filters = append(filters, fmt.Sprintf(`tenant_id="%s"`, q.TenantID))
	}
	if q.Operation != "" {
		filters = append(filters, fmt.Sprintf(`operation=~"%s"`, q.Operation))
	}
	if !q.StartTime.IsZero() {
		filters = append(filters, fmt.Sprintf(`timestamp >= "%s"`, q.StartTime.Format(time.RFC3339)))
	}
	if !q.EndTime.IsZero() {
		filters = append(filters, fmt.Sprintf(`timestamp <= "%s"`, q.EndTime.Format(time.RFC3339)))
	}

	query := `{job="aeng-audit"}`
	if len(filters) > 0 {
		query += " | " + strings.Join(filters, " ")
	}
	if q.Limit > 0 {
		query += fmt.Sprintf(" | limit %d", q.Limit)
	}
	return query
}
