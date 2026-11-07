package compiler

import (
	"fmt"

	"github.com/bridge-allah/a-r-s-y/arsy/engine/query"
)

var (
	ErrUnsupportedNode = fmt.Errorf("compiler: unsupported node type")
	ErrUnknownField    = fmt.Errorf("compiler: unknown field")
	ErrLimitTooLarge   = fmt.Errorf("compiler: limit exceeds maximum allowed")
	ErrOffsetTooLarge  = fmt.Errorf("compiler: offset exceeds maximum allowed")
	ErrDepthExceeded   = fmt.Errorf("compiler: nesting depth exceeded")
	ErrTooManyNodes    = fmt.Errorf("compiler: too many nodes")
)

const (
	defaultMaxLimit     = 1000
	defaultMaxOffset    = 10000
	defaultMaxDepth     = 20
	defaultMaxNodeCount = 1000
)

type Schema interface {
	HasField(name string) bool
}

type Plan struct {
	SQL  string
	Args []interface{}
}

type Options struct {
	MaxLimit     int
	MaxOffset    int
	MaxDepth     int
	MaxNodeCount int
}

func (o Options) defaults() Options {
	if o.MaxLimit <= 0 {
		o.MaxLimit = defaultMaxLimit
	}
	if o.MaxOffset <= 0 {
		o.MaxOffset = defaultMaxOffset
	}
	if o.MaxDepth <= 0 {
		o.MaxDepth = defaultMaxDepth
	}
	if o.MaxNodeCount <= 0 {
		o.MaxNodeCount = defaultMaxNodeCount
	}
	return o
}

var allowedOpsMap = map[string]bool{
	query.OpEq: true, query.OpNeq: true, query.OpGt: true, query.OpGte: true,
	query.OpLt: true, query.OpLte: true, query.OpIn: true, query.OpNin: true,
	query.OpLike: true, query.OpNLike: true,
}

func isSafeIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

func quoteField(p *Plan, name string, s Schema) error {
	if name == "*" {
		p.SQL += name
		return nil
	}
	if s != nil && !s.HasField(name) {
		return fmt.Errorf("%w: %q", ErrUnknownField, name)
	}
	if !isSafeIdentifier(name) {
		return fmt.Errorf("%w: %q", ErrUnknownField, name)
	}
	p.SQL += `"` + name + `"`
	return nil
}

func Compile(q *query.Query, s Schema, opts Options) (*Plan, error) {
	o := opts.defaults()

	plan := &Plan{SQL: "SELECT ", Args: nil}

	fields := q.Select
	if len(fields) == 0 {
		fields = []string{"*"}
	}
	for i, f := range fields {
		if i > 0 {
			plan.SQL += ", "
		}
		if err := quoteField(plan, f, s); err != nil {
			return nil, err
		}
	}
	plan.SQL += " FROM target"

	if q.Filter != nil {
		if err := validateExpr(q.Filter, s); err != nil {
			return nil, err
		}
		plan.SQL += " WHERE "
		if err := compileExpr(q.Filter, plan, s, o, 0); err != nil {
			return nil, err
		}
	}

	if len(q.Sort) > 0 {
		plan.SQL += " ORDER BY "
		for i, srt := range q.Sort {
			if i > 0 {
				plan.SQL += ", "
			}
			if srt.Field == "" || (s != nil && !s.HasField(srt.Field)) {
				return nil, fmt.Errorf("%w: %q", ErrUnknownField, srt.Field)
			}
			if !isSafeIdentifier(srt.Field) {
				return nil, fmt.Errorf("%w: %q", ErrUnknownField, srt.Field)
			}
			plan.SQL += `"` + srt.Field + `"`
			if srt.Desc {
				plan.SQL += " DESC"
			} else {
				plan.SQL += " ASC"
			}
		}
	}

	if q.Limit > 0 {
		if q.Limit > o.MaxLimit {
			return nil, ErrLimitTooLarge
		}
		plan.SQL += " LIMIT $"
		plan.SQL += fmt.Sprintf("%d", len(plan.Args)+1)
		plan.Args = append(plan.Args, q.Limit)
	}
	if q.Offset > 0 {
		if q.Offset > o.MaxOffset {
			return nil, ErrOffsetTooLarge
		}
		plan.SQL += " OFFSET $"
		plan.SQL += fmt.Sprintf("%d", len(plan.Args)+1)
		plan.Args = append(plan.Args, q.Offset)
	}

	return plan, nil
}

func validateExpr(e *query.Expr, s Schema) error {
	if e == nil {
		return nil
	}
	if err := validateShape(e, s); err != nil {
		return err
	}
	if e.Cmp != nil {
		return validateCmp(e.Cmp, s)
	}
	return validateGroup(e, s)
}

func validateShape(e *query.Expr, s Schema) error {
	count := 0
	if len(e.And) > 0 {
		count++
		for _, c := range e.And {
			if c == nil {
				return fmt.Errorf("%w: nil AND child", query.ErrMalformed)
			}
		}
	}
	if len(e.Or) > 0 {
		count++
		for _, c := range e.Or {
			if c == nil {
				return fmt.Errorf("%w: nil OR child", query.ErrMalformed)
			}
		}
	}
	if e.Not != nil {
		count++
	}
	if e.Cmp != nil {
		count++
	}
	if count != 1 {
		return fmt.Errorf("%w: expression must have exactly one of and/or/not/cmp", query.ErrMalformed)
	}
	return nil
}

func validateGroup(e *query.Expr, s Schema) error {
	for _, c := range e.And {
		if err := validateExpr(c, s); err != nil {
			return err
		}
	}
	for _, c := range e.Or {
		if err := validateExpr(c, s); err != nil {
			return err
		}
	}
	if e.Not != nil {
		if err := validateExpr(e.Not, s); err != nil {
			return err
		}
	}
	return nil
}

func validateCmp(c *query.Comparison, s Schema) error {
	if s != nil && !s.HasField(c.Field) {
		return fmt.Errorf("%w: %q", ErrUnknownField, c.Field)
	}
	if !allowedOpsMap[c.Op] {
		return fmt.Errorf("%w: %q", query.ErrOperator, c.Op)
	}
	if len(c.Value) == 0 {
		return fmt.Errorf("%w: comparison value required", query.ErrMalformed)
	}
	return nil
}

func compileExpr(e *query.Expr, p *Plan, s Schema, o Options, depth int) error {
	if depth > o.MaxDepth {
		return ErrDepthExceeded
	}
	o.MaxNodeCount--
	if o.MaxNodeCount <= 0 {
		return ErrTooManyNodes
	}

	if e.Cmp != nil {
		return compileCmp(e.Cmp, p, s)
	}
	if len(e.And) > 0 {
		p.SQL += "("
		for i, c := range e.And {
			if i > 0 {
				p.SQL += " AND "
			}
			if err := compileExpr(c, p, s, o, depth+1); err != nil {
				return err
			}
		}
		p.SQL += ")"
		return nil
	}
	if len(e.Or) > 0 {
		p.SQL += "("
		for i, c := range e.Or {
			if i > 0 {
				p.SQL += " OR "
			}
			if err := compileExpr(c, p, s, o, depth+1); err != nil {
				return err
			}
		}
		p.SQL += ")"
		return nil
	}
	if e.Not != nil {
		p.SQL += "( NOT "
		if err := compileExpr(e.Not, p, s, o, depth+1); err != nil {
			return err
		}
		p.SQL += ")"
		return nil
	}
	return ErrUnsupportedNode
}

func compileCmp(c *query.Comparison, p *Plan, s Schema) error {
	if s != nil && !s.HasField(c.Field) {
		return fmt.Errorf("%w: %q", ErrUnknownField, c.Field)
	}
	if !allowedOpsMap[c.Op] {
		return fmt.Errorf("%w: %q", query.ErrOperator, c.Op)
	}
	if !isSafeIdentifier(c.Field) {
		return fmt.Errorf("%w: %q", ErrUnknownField, c.Field)
	}
	p.SQL += `"` + c.Field + `" ` + c.Op + " $"
	p.SQL += fmt.Sprintf("%d", len(p.Args)+1)
	p.Args = append(p.Args, c.Value)
	return nil
}
