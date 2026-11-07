package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bridge-allah/a-r-s-y/core"
)

func TestRequireSuperuserRejectsMissingAndRegularAuth(t *testing.T) {
	regularUser := core.NewRecord(&core.Collection{Name: "users"})
	for name, auth := range map[string]*core.Record{
		"missing": nil,
		"regular user": regularUser,
	} {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			event := &core.RequestEvent{Auth: auth}
			event.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			event.Response = rec

			called := false
			err := requireSuperuser(func(e *core.RequestEvent) error {
				called = true
				return e.JSON(http.StatusOK, map[string]bool{"ok": true})
			})(event)
			if err != nil {
				t.Fatal(err)
			}
			if called {
				t.Fatal("protected handler was called")
			}
			wantStatus := http.StatusUnauthorized
			if name == "regular user" {
				wantStatus = http.StatusForbidden
			}
			if rec.Code != wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, wantStatus)
			}
		})
	}
}

func TestRequireSuperuserAllowsSuperuser(t *testing.T) {
	superuser := core.NewRecord(&core.Collection{Name: core.CollectionNameSuperusers})
	rec := httptest.NewRecorder()
	event := &core.RequestEvent{Auth: superuser}
	event.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	event.Response = rec

	called := false
	err := requireSuperuser(func(e *core.RequestEvent) error {
		called = true
		return e.JSON(http.StatusOK, map[string]bool{"ok": true})
	})(event)
	if err != nil {
		t.Fatal(err)
	}
	if !called || rec.Code != http.StatusOK {
		t.Fatalf("handler called = %v, status = %d", called, rec.Code)
	}
}
