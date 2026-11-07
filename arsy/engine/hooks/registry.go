package hooks

import (
	"context"
	"fmt"
)

// Op identifies a model lifecycle operation.
type Op string

const (
	OpCreate Op = "create"
	OpUpdate Op = "update"
	OpDelete Op = "delete"
)

// Phase identifies when a hook runs relative to persistence.
type Phase string

const (
	PhaseBefore      Phase = "before"
	PhaseAfter       Phase = "after"
	PhaseAfterCommit Phase = "after_commit"
)

// Handler executes one hook. Returning an error aborts the operation for
// before hooks and aborts the remaining queue for after phases.
type Handler func(ctx context.Context, ev *Event) error

// Event carries the operation and the affected model value. Model stays
// driver-neutral so static Go models and metadata records can share hooks.
type Event struct {
	Op    Op
	Model any
}

type registration struct {
	id string
	fn Handler
}

// Registry stores hooks per operation and phase. The zero value is ready
// for use and safe for concurrent registration and triggering.
type Registry struct {
	store map[Op]map[Phase][]registration
}

func (r *Registry) slot(op Op, phase Phase) []registration {
	return r.store[op][phase]
}

// Register adds a hook. Order of registration is preserved.
func (r *Registry) Register(op Op, phase Phase, id string, fn Handler) error {
	if r.store == nil {
		r.store = make(map[Op]map[Phase][]registration)
	}
	if phase != PhaseBefore && phase != PhaseAfter && phase != PhaseAfterCommit {
		return fmt.Errorf("%w: %q", ErrUnknownPhase, phase)
	}
	byPhase := r.store[op]
	if byPhase == nil {
		byPhase = make(map[Phase][]registration)
		r.store[op] = byPhase
	}
	for _, reg := range byPhase[phase] {
		if reg.id == id {
			return fmt.Errorf("%w: %s/%s/%s", ErrDuplicateHook, op, phase, id)
		}
	}
	byPhase[phase] = append(byPhase[phase], registration{id: id, fn: fn})
	return nil
}

func (r *Registry) runList(ctx context.Context, list []registration, ev *Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, reg := range list {
		if err := reg.fn(ctx, ev); err != nil {
			return fmt.Errorf("hooks: %s failed: %w", reg.id, err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}

// TriggerBefore runs before hooks for the operation in registration order.
// The first error stops execution.
func (r *Registry) TriggerBefore(ctx context.Context, op Op, ev *Event) error {
	return r.trigger(ctx, op, PhaseBefore, ev)
}

// TriggerAfter runs in-transaction after hooks in registration order.
func (r *Registry) TriggerAfter(ctx context.Context, op Op, ev *Event) error {
	return r.trigger(ctx, op, PhaseAfter, ev)
}

func (r *Registry) trigger(ctx context.Context, op Op, phase Phase, ev *Event) error {
	if ctx.Value(reentrantKey{}) != nil {
		return ErrReentrant
	}
	hookCtx := context.WithValue(ctx, reentrantKey{}, true)
	return r.runList(hookCtx, r.slot(op, phase), ev)
}

// Queue holds after-commit hooks collected during an operation.
type Queue struct {
	items []registration
	ev    *Event
}

// CollectAfterCommit snapshots the after-commit hooks for the operation.
// Call Queue.Flush only after the surrounding transaction commits.
func (r *Registry) CollectAfterCommit(op Op, ev *Event) *Queue {
	return &Queue{items: append([]registration{}, r.slot(op, PhaseAfterCommit)...), ev: ev}
}

// Flush runs the queued after-commit hooks in registration order.
func (q *Queue) Flush(ctx context.Context) error {
	if q == nil {
		return nil
	}
	// After-commit hooks are side-effect hooks; re-entrancy between them
	// is still guarded to keep execution deterministic.
	if ctx.Value(reentrantKey{}) != nil {
		return ErrReentrant
	}
	hookCtx := context.WithValue(ctx, reentrantKey{}, true)
	return (&Registry{}).runList(hookCtx, q.items, q.ev)
}

type reentrantKey struct{}
