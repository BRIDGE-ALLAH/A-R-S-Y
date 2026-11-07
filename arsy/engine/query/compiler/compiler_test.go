package compiler

import (
	"errors"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/engine/query"
)

type testSchema struct{ fields map[string]bool }

func (s testSchema) HasField(name string) bool { return s.fields[name] }

func TestCompile(t *testing.T) {
	s := testSchema{fields: map[string]bool{"name": true, "age": true, "active": true}}

	tc := []struct {
		name     string
		q        *query.Query
		wantSQL  string
		wantArgs int
		wantErr  error
	}{
		{
			name: "simple equality",
			q: &query.Query{
				Filter: &query.Expr{Cmp: &query.Comparison{Field: "name", Op: query.OpEq, Value: []byte(`"foo"`)}},
			},
			wantSQL:  `SELECT * FROM target WHERE "name" = $1`,
			wantArgs: 1,
		},
		{
			name: "and group",
			q: &query.Query{
				Filter: &query.Expr{
					And: []*query.Expr{
						{Cmp: &query.Comparison{Field: "name", Op: query.OpEq, Value: []byte(`"foo"`)}},
						{Cmp: &query.Comparison{Field: "age", Op: query.OpGt, Value: []byte(`30`)}},
					},
				},
			},
			wantSQL:  `SELECT * FROM target WHERE ("name" = $1 AND "age" > $2)`,
			wantArgs: 2,
		},
		{
			name: "limit and offset",
			q: &query.Query{
				Filter: &query.Expr{Cmp: &query.Comparison{Field: "active", Op: query.OpEq, Value: []byte(`true`)}},
				Limit:  10,
				Offset: 5,
			},
			wantSQL:  `SELECT * FROM target WHERE "active" = $1 LIMIT $2 OFFSET $3`,
			wantArgs: 3,
		},
		{
			name:    "unknown field",
			q:       &query.Query{Filter: &query.Expr{Cmp: &query.Comparison{Field: "nope", Op: query.OpEq, Value: []byte(`1`)}}},
			wantErr: ErrUnknownField,
		},
		{
			name:    "bad operator",
			q:       &query.Query{Filter: &query.Expr{Cmp: &query.Comparison{Field: "name", Op: "drop", Value: []byte(`1`)}}},
			wantErr: query.ErrOperator,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Compile(tt.q, s, Options{})
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.SQL != tt.wantSQL {
				t.Fatalf("got %q, want %q", p.SQL, tt.wantSQL)
			}
			if len(p.Args) != tt.wantArgs {
				t.Fatalf("got %d args, want %d", len(p.Args), tt.wantArgs)
			}
		})
	}
}

func TestCompileLimitTooLarge(t *testing.T) {
	q := &query.Query{Limit: 9999}
	_, err := Compile(q, nil, Options{MaxLimit: 100})
	if !errors.Is(err, ErrLimitTooLarge) {
		t.Fatalf("got %v, want %v", err, ErrLimitTooLarge)
	}
}

func TestCompileUnknownField(t *testing.T) {
	q := &query.Query{Select: []string{"secret"}}
	_, err := Compile(q, testSchema{fields: map[string]bool{"name": true}}, Options{})
	if !errors.Is(err, ErrUnknownField) {
		t.Fatalf("got %v, want %v", err, ErrUnknownField)
	}
}
