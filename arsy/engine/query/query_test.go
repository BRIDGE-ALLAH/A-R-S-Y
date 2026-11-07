package query

import (
	"encoding/json"
	"errors"
	"testing"
)

type testSchema struct{ fields map[string]bool }

func (s testSchema) HasField(n string) bool { return s.fields[n] }

var schema = testSchema{fields: map[string]bool{"name": true, "age": true, "status": true}}

func TestValidate(t *testing.T) {
	cases := []struct {
		name    string
		q       *Query
		wantErr error
	}{
		{name: "nil query", q: nil, wantErr: ErrMalformed},
		{name: "valid comparison", q: &Query{Filter: &Expr{Cmp: &Comparison{Field: "name", Op: OpEq, Value: json.RawMessage(`"x"`)}}}},
		{name: "unknown field", q: &Query{Filter: &Expr{Cmp: &Comparison{Field: "nope", Op: OpEq, Value: json.RawMessage(`1`)}}}, wantErr: ErrUnknownField},
		{name: "bad operator", q: &Query{Filter: &Expr{Cmp: &Comparison{Field: "name", Op: "drop", Value: json.RawMessage(`1`)}}}, wantErr: ErrOperator},
		{name: "malformed empty expr", q: &Query{Filter: &Expr{}}, wantErr: ErrMalformed},
		{name: "negative limit", q: &Query{Limit: -1}, wantErr: ErrMalformed},
		{name: "valid and group", q: &Query{Filter: &Expr{And: []*Expr{{Cmp: &Comparison{Field: "age", Op: OpGte, Value: json.RawMessage(`18`)}}}}}},
		{name: "nil and child", q: &Query{Filter: &Expr{And: []*Expr{nil}}}, wantErr: ErrMalformed},
		{name: "unknown sort field", q: &Query{Sort: []Sort{{Field: "nope"}}}, wantErr: ErrUnknownField},
		{name: "unknown select field", q: &Query{Select: []string{"nope"}}, wantErr: ErrUnknownField},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.q.Validate(schema)
			if tc.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestDeterministicMarshal(t *testing.T) {
	q := &Query{Select: []string{"status", "age", "name"}}
	b1, _ := json.Marshal(q)
	b2, _ := json.Marshal(q)
	if string(b1) != string(b2) {
		t.Fatalf("non-deterministic: %s vs %s", b1, b2)
	}
	if string(b1) != `{"select":["age","name","status"]}` {
		t.Fatalf("unexpected json: %s", b1)
	}
}
