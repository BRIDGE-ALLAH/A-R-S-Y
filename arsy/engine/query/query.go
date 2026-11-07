package query

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

// Package query defines an engine-neutral, validated query filter AST (issue #107).
// No SQL generation and no raw user SQL fragments are exposed by this package.

var (
	ErrOperator     = errors.New("query: operator not allowlisted")
	ErrMalformed    = errors.New("query: malformed expression")
	ErrUnknownField = errors.New("query: unknown field")
	ErrLimit        = errors.New("query: limit exceeds configured maximum")
	ErrOffset       = errors.New("query: offset exceeds configured maximum")
	ErrDepth        = errors.New("query: nesting depth exceeds configured maximum")
	ErrNodeCount    = errors.New("query: node count exceeds configured maximum")
)

// Operator allowlist.
const (
	OpEq    = "="
	OpNeq   = "!="
	OpGt    = ">"
	OpGte   = ">="
	OpLt    = "<"
	OpLte   = "<="
	OpIn    = "in"
	OpNin   = "nin"
	OpLike  = "like"
	OpNLike = "nlike"
)

var allowedOps = map[string]bool{
	OpEq: true, OpNeq: true, OpGt: true, OpGte: true, OpLt: true, OpLte: true,
	OpIn: true, OpNin: true, OpLike: true, OpNLike: true,
}

// Schema describes which fields are valid for validation and selection.
type Schema interface {
	HasField(name string) bool
}

// Expr is a filter expression.
type Expr struct {
	And []*Expr     `json:"and,omitempty"`
	Or  []*Expr     `json:"or,omitempty"`
	Not *Expr       `json:"not,omitempty"`
	Cmp *Comparison `json:"cmp,omitempty"`
}

// Comparison is a leaf predicate. Value is always a bound parameter.
type Comparison struct {
	Field string          `json:"field"`
	Op    string          `json:"op"`
	Value json.RawMessage `json:"value"`
}

// Query is a complete, validated request.
type Query struct {
	Filter *Expr    `json:"filter,omitempty"`
	Sort   []Sort   `json:"sort,omitempty"`
	Limit  int      `json:"limit,omitempty"`
	Offset int      `json:"offset,omitempty"`
	Select []string `json:"select,omitempty"`
}

type Sort struct {
	Field string `json:"field"`
	Desc  bool   `json:"desc,omitempty"`
}

// ValidateOptions configures query complexity constraints (#127).
// Zero values use package defaults.
type ValidateOptions struct {
	MaxLimit     int
	MaxOffset    int
	MaxDepth     int
	MaxNodeCount int
}

const (
	DefaultMaxLimit     = 1000
	DefaultMaxOffset    = 10000
	DefaultMaxDepth     = 20
	DefaultMaxNodeCount = 1000
)

func (o ValidateOptions) withDefaults() ValidateOptions {
	if o.MaxLimit <= 0 {
		o.MaxLimit = DefaultMaxLimit
	}
	if o.MaxOffset <= 0 {
		o.MaxOffset = DefaultMaxOffset
	}
	if o.MaxDepth <= 0 {
		o.MaxDepth = DefaultMaxDepth
	}
	if o.MaxNodeCount <= 0 {
		o.MaxNodeCount = DefaultMaxNodeCount
	}
	return o
}

// Validate checks operator allowlist, expression shape, schema fields, and default complexity constraints.
func (q *Query) Validate(s Schema) error {
	return q.ValidateWithOptions(s, ValidateOptions{})
}

// ValidateWithOptions checks operator allowlist, expression shape, schema fields, and configurable complexity constraints.
func (q *Query) ValidateWithOptions(s Schema, opts ValidateOptions) error {
	o := opts.withDefaults()
	if q == nil {
		return fmt.Errorf("%w: nil query", ErrMalformed)
	}
	if q.Limit < 0 || q.Offset < 0 {
		return fmt.Errorf("%w: negative limit/offset", ErrMalformed)
	}
	if q.Limit > o.MaxLimit {
		return fmt.Errorf("%w: %d > %d", ErrLimit, q.Limit, o.MaxLimit)
	}
	if q.Offset > o.MaxOffset {
		return fmt.Errorf("%w: %d > %d", ErrOffset, q.Offset, o.MaxOffset)
	}
	nodes := 0
	if err := q.Filter.validate(s, o, 0, &nodes); err != nil {
		return err
	}
	for _, so := range q.Sort {
		if so.Field == "" || (s != nil && !s.HasField(so.Field)) {
			return fmt.Errorf("%w: %q", ErrUnknownField, so.Field)
		}
	}
	for _, f := range q.Select {
		if f == "" || (s != nil && !s.HasField(f)) {
			return fmt.Errorf("%w: %q", ErrUnknownField, f)
		}
	}
	return nil
}

// Validate checks operator allowlist, expression shape, schema fields, and default complexity constraints.
func (e *Expr) Validate(s Schema) error {
	nodes := 0
	return e.validate(s, ValidateOptions{}.withDefaults(), 0, &nodes)
}

func (e *Expr) validate(s Schema, o ValidateOptions, depth int, nodes *int) error {
	if e == nil {
		return nil
	}
	if depth > o.MaxDepth {
		return fmt.Errorf("%w: %d > %d", ErrDepth, depth, o.MaxDepth)
	}
	*nodes++
	if *nodes > o.MaxNodeCount {
		return fmt.Errorf("%w: %d > %d", ErrNodeCount, *nodes, o.MaxNodeCount)
	}
	count := 0
	if len(e.And) > 0 {
		count++
		for _, c := range e.And {
			if c == nil {
				return fmt.Errorf("%w: nil AND child", ErrMalformed)
			}
			if err := c.validate(s, o, depth+1, nodes); err != nil {
				return err
			}
		}
	}
	if len(e.Or) > 0 {
		count++
		for _, c := range e.Or {
			if c == nil {
				return fmt.Errorf("%w: nil OR child", ErrMalformed)
			}
			if err := c.validate(s, o, depth+1, nodes); err != nil {
				return err
			}
		}
	}
	if e.Not != nil {
		count++
		if err := e.Not.validate(s, o, depth+1, nodes); err != nil {
			return err
		}
	}
	if e.Cmp != nil {
		count++
		if !allowedOps[e.Cmp.Op] {
			return fmt.Errorf("%w: %q", ErrOperator, e.Cmp.Op)
		}
		if e.Cmp.Field == "" || (s != nil && !s.HasField(e.Cmp.Field)) {
			return fmt.Errorf("%w: %q", ErrUnknownField, e.Cmp.Field)
		}
		if len(e.Cmp.Value) == 0 || (string(e.Cmp.Value) == "null" && (e.Cmp.Op == OpIn || e.Cmp.Op == OpNin)) {
			return fmt.Errorf("%w: comparison value required", ErrMalformed)
		}
	}
	if count != 1 {
		return fmt.Errorf("%w: expression must have exactly one of and/or/not/cmp", ErrMalformed)
	}
	return nil
}

// MarshalJSON produces deterministic output (sorted select keys by field order preserved,
// stable struct field order via encoding/json on fixed structs).
func (q *Query) MarshalJSON() ([]byte, error) {
	if q == nil {
		return []byte("null"), nil
	}
	cp := *q
	if len(cp.Select) > 1 {
		sel := append([]string(nil), cp.Select...)
		sort.Strings(sel)
		cp.Select = sel
	}
	type alias Query
	return json.Marshal((*alias)(&cp))
}
