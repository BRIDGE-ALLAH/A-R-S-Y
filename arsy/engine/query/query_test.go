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

func TestValidateWithOptionsComplexity(t *testing.T) {
	cmp := func() *Expr {
		return &Expr{Cmp: &Comparison{Field: "name", Op: OpEq, Value: json.RawMessage(`"x"`)}}
	}
	deep := cmp()
	for i := 0; i < 4; i++ {
		deep = &Expr{Not: deep}
	}
	many := &Expr{And: []*Expr{cmp(), cmp(), cmp(), cmp()}}

	cases := []struct {
		name    string
		q       *Query
		opts    ValidateOptions
		wantErr error
	}{
		{name: "limit too large", q: &Query{Limit: 5}, opts: ValidateOptions{MaxLimit: 2}, wantErr: ErrLimit},
		{name: "offset too large", q: &Query{Offset: 5}, opts: ValidateOptions{MaxOffset: 2}, wantErr: ErrOffset},
		{name: "depth too large", q: &Query{Filter: deep}, opts: ValidateOptions{MaxDepth: 2}, wantErr: ErrDepth},
		{name: "node count too large", q: &Query{Filter: many}, opts: ValidateOptions{MaxNodeCount: 2}, wantErr: ErrNodeCount},
		{name: "configured limits allow", q: &Query{Filter: cmp(), Limit: 5, Offset: 5}, opts: ValidateOptions{MaxLimit: 5, MaxOffset: 5, MaxDepth: 1, MaxNodeCount: 1}, wantErr: nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.q.ValidateWithOptions(schema, tc.opts)
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
