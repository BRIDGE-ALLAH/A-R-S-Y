package persist

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"ganigeorgiev.com/pocketbase/dbx"
	"github.com/bridge-allah/a-r-s-y/arsy/engine/query"
	"github.com/bridge-allah/a-r-s-y/arsy/engine/registry"
	_ "modernc.org/sqlite"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	db, err := dbx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.NewQuery(`CREATE TABLE base_company (id TEXT PRIMARY KEY, name TEXT NOT NULL, code TEXT UNIQUE, active TEXT, updated TEXT)`).Execute(); err != nil {
		t.Fatal(err)
	}
	reg := registry.New()
	if err := reg.Register(registry.Descriptor{ID: "base.company", Label: "Company", TableName: "base_company", Version: "1.0.0", OwnerModule: "base", Source: registry.SourceMetadata}); err != nil {
		t.Fatal(err)
	}
	store, err := New(reg, db)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestStoreCRUDAndTypedErrors(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	id, err := store.Create(ctx, "base.company", map[string]any{"id": "c1", "name": "Acme", "code": "ACME", "active": "true"})
	if err != nil || id != "c1" {
		t.Fatalf("create: id=%q err=%v", id, err)
	}
	row, err := store.Get(ctx, "base.company", "c1")
	if err != nil || fmt.Sprint(row["name"]) != "Acme" {
		t.Fatalf("get: %#v err=%v", row, err)
	}
	if err := store.Update(ctx, "base.company", "c1", map[string]any{"name": "Acme Updated"}); err != nil {
		t.Fatal(err)
	}
	row, _ = store.Get(ctx, "base.company", "c1")
	if fmt.Sprint(row["name"]) != "Acme Updated" {
		t.Fatalf("updated row: %#v", row)
	}
	if _, err := store.Get(ctx, "base.company", "missing"); !errors.As(err, new(*ErrNotFound)) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := store.Update(ctx, "base.company", "missing", map[string]any{"name": "x"}); !errors.As(err, new(*ErrNotFound)) {
		t.Fatalf("expected ErrNotFound on update, got %v", err)
	}
	if err := store.Delete(ctx, "base.company", "missing"); !errors.As(err, new(*ErrNotFound)) {
		t.Fatalf("expected ErrNotFound on delete, got %v", err)
	}
	if err := store.Delete(ctx, "base.company", "c1"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(ctx, "base.company", "c1"); !errors.As(err, new(*ErrNotFound)) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
	if _, err := store.Create(ctx, "unknown.model", map[string]any{"id": "x"}); !errors.As(err, new(*ErrUnknownModel)) {
		t.Fatalf("expected ErrUnknownModel, got %v", err)
	}
	if _, err := store.Create(ctx, "base.company", nil); !errors.As(err, new(*ErrUnknownField)) {
		t.Fatalf("expected ErrUnknownField for nil values, got %v", err)
	}
}

func TestStoreConstraintAndQuery(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	if _, err := store.Create(ctx, "base.company", map[string]any{"id": "c1", "name": "Acme", "code": "ACME"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create(ctx, "base.company", map[string]any{"id": "c2", "name": "Acme 2", "code": "ACME"}); !errors.As(err, new(*ErrConstraintViolation)) {
		t.Fatalf("expected ErrConstraintViolation, got %v", err)
	}
	rows, err := store.Query(ctx, "base.company", &query.Query{
		Select: []string{"id", "name"},
		Filter: &query.Expr{Cmp: &query.Comparison{Field: "code", Op: query.OpEq, Value: json.RawMessage(`"ACME"`)}},
		Limit:  10,
	}, []string{"id", "name", "code"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || fmt.Sprint(rows[0]["id"]) != "c1" {
		ids := make([]string, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, fmt.Sprint(r["id"]))
		}
		t.Fatalf("unexpected query rows count=%d ids=%s", len(rows), strings.Join(ids, ","))
	}
	if _, err := store.Query(ctx, "base.company", &query.Query{Select: []string{"id"}}, nil); !errors.As(err, new(*ErrUnknownField)) {
		t.Fatalf("expected ErrUnknownField without explicit allowed fields, got %v", err)
	}
}
