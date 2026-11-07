package validation

import (
	"errors"
	"testing"
)

func TestApplyDefaults(t *testing.T) {
	fields := []Field{
		{Name: "id", SystemManaged: true},
		{Name: "name", Type: "string", Required: true},
		{Name: "active", Type: "bool", Default: true},
	}

	t.Run("create applies defaults", func(t *testing.T) {
		out, err := ApplyDefaults(fields, map[string]any{"name": "Acme"}, OpCreate)
		if err != nil {
			t.Fatal(err)
		}
		if out["active"] != true {
			t.Errorf("expected default active=true, got %v", out["active"])
		}
		if _, ok := out["id"]; ok {
			t.Errorf("system-managed field should not be set by defaults pipeline")
		}
	})

	t.Run("update does not overwrite omitted with defaults", func(t *testing.T) {
		out, err := ApplyDefaults(fields, map[string]any{"name": "New"}, OpUpdate)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := out["active"]; ok {
			t.Errorf("omitted field should not get default on update")
		}
	})

	t.Run("strips client-provided protected field", func(t *testing.T) {
		out, _ := ApplyDefaults(fields, map[string]any{"id": "hacker", "name": "x"}, OpCreate)
		if out["id"] == "hacker" {
			t.Errorf("client must not override protected field")
		}
	})
}

func TestValidate(t *testing.T) {
	fields := []Field{
		{Name: "name", Type: "string", Required: true},
		{Name: "age", Type: "int", Required: false},
	}

	tests := []struct {
		name      string
		values    map[string]any
		op        Op
		wantField string
		wantRule  string
	}{
		{name: "required missing create", values: map[string]any{}, op: OpCreate, wantField: "name", wantRule: "required"},
		{name: "required omitted update ok", values: map[string]any{}, op: OpUpdate},
		{name: "type mismatch", values: map[string]any{"name": "ok", "age": "not-int"}, op: OpCreate, wantField: "age", wantRule: "type:int"},
		{name: "valid", values: map[string]any{"name": "ok", "age": 30}, op: OpCreate},
		{name: "empty string required", values: map[string]any{"name": "  "}, op: OpCreate, wantField: "name", wantRule: "required"},
		{name: "nil value skips type", values: map[string]any{"name": "ok", "age": nil}, op: OpCreate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := Validate(fields, tt.values, tt.op)
			if tt.wantField == "" {
				if len(errs) != 0 {
					t.Fatalf("expected no errors, got %v", errs)
				}
				return
			}
			if len(errs) == 0 {
				t.Fatal("expected error")
			}
			var fe *FieldError
			if !errors.As(errs[0], &fe) {
				t.Fatalf("expected FieldError, got %T", errs[0])
			}
			if fe.Field != tt.wantField || fe.Rule != tt.wantRule {
				t.Fatalf("got %s/%s, want %s/%s", fe.Field, fe.Rule, tt.wantField, tt.wantRule)
			}
		})
	}
}

func TestValidateInput(t *testing.T) {
	fields := []Field{{Name: "id", SystemManaged: true}, {Name: "name"}}
	errs := ValidateInput(fields, map[string]any{"id": "x", "name": "ok"})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	var fe *FieldError
	errors.As(errs[0], &fe)
	if fe.Field != "id" || !errors.Is(fe.Err, ErrProtected) {
		t.Fatalf("unexpected error: %v", fe)
	}
}
