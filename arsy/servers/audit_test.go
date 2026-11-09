package servers

import (
	"strings"
	"testing"
	"time"

	"github.com/bridge-allah/a-r-s-y/core"
	"github.com/bridge-allah/a-r-s-y/tests"
)

func TestLogAuditEventCreatesRecord(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	collection, err := app.FindCollectionByNameOrId("arsy_servers")
	if err != nil {
		t.Fatal(err)
	}
	server := core.NewRecord(collection)
	server.Set("name", "audit-server")
	server.Set("host", "audit.example.test")
	server.Set("ssh_port", 22)
	server.Set("ssh_username", "deploy")
	if err := app.Save(server); err != nil {
		t.Fatal(err)
	}

	if err := LogAuditEvent(app, "server.credentials.rotate", server.Id, "user-1", map[string]any{
		"method": "ssh_private_key",
	}); err != nil {
		t.Fatalf("LogAuditEvent() error = %v", err)
	}

	records, err := app.FindRecordsByFilter(
		"arsy_audit_log",
		"server = {:serverID}",
		"-created",
		100, 0,
		map[string]any{"serverID": server.Id},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) == 0 {
		t.Fatal("expected at least one audit record")
	}
	if records[0].GetString("event_type") != "server.credentials.rotate" {
		t.Fatalf("event_type = %q, want %q", records[0].GetString("event_type"), "server.credentials.rotate")
	}
	if records[0].GetString("changed_by") != "user-1" {
		t.Fatalf("changed_by = %q, want %q", records[0].GetString("changed_by"), "user-1")
	}
	if records[0].GetString("server") != server.Id {
		t.Fatalf("server = %q, want %q", records[0].GetString("server"), server.Id)
	}
}

func TestLogAuditEventMissingCollectionGraceful(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	// Temporarily drop the audit collection so LogAuditEvent soft-fails.
	collection, err := app.FindCollectionByNameOrId("arsy_audit_log")
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Delete(collection); err != nil {
		t.Fatal(err)
	}

	err = LogAuditEvent(app, "server.credentials.rotate", "server-1", "user-1", nil)
	if err == nil {
		t.Fatal("expected error when audit collection is missing")
	}
}

func TestGetServerAuditLogFiltersByEventType(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	collection, err := app.FindCollectionByNameOrId("arsy_servers")
	if err != nil {
		t.Fatal(err)
	}
	server := core.NewRecord(collection)
	server.Set("name", "audit-server-2")
	server.Set("host", "audit2.example.test")
	server.Set("ssh_port", 22)
	server.Set("ssh_username", "deploy")
	if err := app.Save(server); err != nil {
		t.Fatal(err)
	}

	_ = LogAuditEvent(app, "server.credentials.rotate", server.Id, "user-1", map[string]any{"method": "ssh_private_key"})
	_ = LogAuditEvent(app, "host_key_revoked", server.Id, "user-1", map[string]any{"revoked_at": "now"})

	records, err := GetServerAuditLog(app, server.Id, "host_key_revoked", 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 host_key_revoked record, got %d", len(records))
	}
	if records[0].GetString("event_type") != "host_key_revoked" {
		t.Fatalf("event_type = %q, want %q", records[0].GetString("event_type"), "host_key_revoked")
	}

	all, err := GetServerAuditLog(app, server.Id, "", 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 total audit records, got %d", len(all))
	}
}

func TestRevokeServerHostKeyMarksRevokedAndLogs(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	collection, err := app.FindCollectionByNameOrId("arsy_servers")
	if err != nil {
		t.Fatal(err)
	}
	server := core.NewRecord(collection)
	server.Set("name", "revoke-server")
	server.Set("host", "revoke.example.test")
	server.Set("ssh_port", 22)
	server.Set("ssh_username", "deploy")
	if err := app.Save(server); err != nil {
		t.Fatal(err)
	}

	if err := RevokeServerHostKey(app, server.Id, "user-1"); err != nil {
		t.Fatalf("RevokeServerHostKey() error = %v", err)
	}

	updated, err := app.FindRecordById("arsy_servers", server.Id)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.GetBool("host_key_revoked") {
		t.Fatal("host_key_revoked should be true after revocation")
	}

	records, err := GetServerAuditLog(app, server.Id, "host_key_revoked", 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 revocation audit record, got %d", len(records))
	}

	// Idempotent: revoking again is a no-op.
	if err := RevokeServerHostKey(app, server.Id, "user-1"); err != nil {
		t.Fatalf("second RevokeServerHostKey() error = %v", err)
	}
	records, err = GetServerAuditLog(app, server.Id, "host_key_revoked", 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("expected revocation audit to be idempotent (1 record), got %d", len(records))
	}
}

func TestRevokeServerHostKeyRejectsUnknownServer(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	err = RevokeServerHostKey(app, "nonexistent-server-id", "user-1")
	if err == nil {
		t.Fatal("expected error for unknown server")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected 'not found' error, got: %v", err)
	}
}

func TestCreateAndGetSession(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	collection, err := app.FindCollectionByNameOrId("arsy_servers")
	if err != nil {
		t.Fatal(err)
	}
	server := core.NewRecord(collection)
	server.Set("name", "session-server")
	server.Set("host", "session.example.test")
	server.Set("ssh_port", 22)
	server.Set("ssh_username", "deploy")
	if err := app.Save(server); err != nil {
		t.Fatal(err)
	}

	if err := CreateSession(app, "session-abc", server.Id, "key", time.Time{}); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	session, err := GetSession(app, "session-abc")
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if session.GetString("session_id") != "session-abc" {
		t.Fatalf("session_id = %q, want %q", session.GetString("session_id"), "session-abc")
	}
	if session.GetString("server") != server.Id {
		t.Fatalf("server = %q, want %q", session.GetString("server"), server.Id)
	}
	if session.GetString("auth_method") != "key" {
		t.Fatalf("auth_method = %q, want %q", session.GetString("auth_method"), "key")
	}

	if err := CreateSession(app, "session-def", server.Id, "password", time.Time{}); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	if _, err := GetSession(app, "session-missing"); err == nil {
		t.Fatal("expected error for missing session")
	}
}

func TestInvalidateSessionsBlocksReuse(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	collection, err := app.FindCollectionByNameOrId("arsy_servers")
	if err != nil {
		t.Fatal(err)
	}
	server := core.NewRecord(collection)
	server.Set("name", "invalidate-server")
	server.Set("host", "invalidate.example.test")
	server.Set("ssh_port", 22)
	server.Set("ssh_username", "deploy")
	if err := app.Save(server); err != nil {
		t.Fatal(err)
	}

	if err := CreateSession(app, "session-1", server.Id, "key", time.Time{}); err != nil {
		t.Fatal(err)
	}
	if err := CreateSession(app, "session-2", server.Id, "key", time.Time{}); err != nil {
		t.Fatal(err)
	}

	if err := InvalidateSessions(app, server.Id); err != nil {
		t.Fatalf("InvalidateSessions() error = %v", err)
	}

	for _, sid := range []string{"session-1", "session-2"} {
		_, err := GetSession(app, sid)
		if err == nil {
			t.Fatalf("expected GetSession(%q) to fail after invalidation", sid)
		}
		if !strings.Contains(err.Error(), "invalidated") {
			t.Fatalf("expected invalidated error, got: %v", err)
		}
	}

	// Sessions for another server must remain valid.
	collection2, err := app.FindCollectionByNameOrId("arsy_servers")
	if err != nil {
		t.Fatal(err)
	}
	other := core.NewRecord(collection2)
	other.Set("name", "other-server")
	other.Set("host", "other.example.test")
	other.Set("ssh_port", 23)
	other.Set("ssh_username", "deploy")
	if err := app.Save(other); err != nil {
		t.Fatal(err)
	}
	if err := CreateSession(app, "session-other", other.Id, "key", time.Time{}); err != nil {
		t.Fatal(err)
	}
	if _, err := GetSession(app, "session-other"); err != nil {
		t.Fatalf("expected other server session to remain valid: %v", err)
	}
}
