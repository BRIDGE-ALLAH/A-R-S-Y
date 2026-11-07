package relations

import (
	"errors"
	"fmt"

	"github.com/bridge-allah/a-r-s-y/arsy/engine/registry"
)

var (
	ErrInvalidSourceModel  = errors.New("invalid relation source model")
	ErrInvalidTargetModel  = errors.New("invalid relation target model")
	ErrInvalidRelationName = errors.New("invalid relation name")
	ErrInvalidCardinality  = errors.New("invalid relation cardinality")
	ErrDuplicateRelation   = errors.New("duplicate relation field")
	ErrInvalidInverse      = errors.New("invalid inverse relation")
	ErrInvalidDeletePolicy = errors.New("invalid relation delete behavior")
	ErrInvalidJoinTable    = errors.New("invalid relation join table")
	ErrInvalidRelationSet  = errors.New("invalid relation set")
)

// ValidationError identifies the relation and rule that failed validation.
type ValidationError struct {
	Source registry.ModelID
	Name   string
	Rule   error
	Detail string
}

func (e *ValidationError) Error() string {
	if e.Name == "" {
		return fmt.Sprintf("%v: %s", e.Rule, e.Detail)
	}
	return fmt.Sprintf("%v: relation %q.%s: %s", e.Rule, e.Source, e.Name, e.Detail)
}

func (e *ValidationError) Unwrap() error { return e.Rule }
