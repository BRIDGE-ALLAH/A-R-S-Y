package hooks

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRegisterDuplicateRejected(t *testing.T) {
	var r Registry
	if err := r.Register(OpCreate, PhaseBefore, "validate", func(context.Context, *Event) error { return nil }); err != nil {
		t.Fatalf("first register: %v", err)
	}
	err := r.Register(OpCreate, PhaseBefore, "validate", func(context.Context, *Event) error { return nil })
	if !errors.Is(err, ErrDuplicateHook) {
		t.Fatalf("expected ErrDuplicateHook, got %v", err)
	}
	// Same ID on another op/phase is fine.
	if err := r.Register(OpUpdate, PhaseBefore, "validate", func(context.Context, *Event) error { return nil }); err != nil {
		t.Fatalf("other op register: %v", err)
	}
}

func TestRegisterUnknownPhase(t *testing.T) {
	var r Registry
	err := r.Register(OpCreate, Phase("nope"), "x", func(context.Context, *Event) error { return nil })
	if !errors.Is(err, ErrUnknownPhase) {
		t.Fatalf("expected ErrUnknownPhase, got %v", err)
	}
}

func TestBeforeHookOrderDeterministicAndVeto(t *testing.T) {
	var r Registry
	var order []string
	mk := func(id string, fail bool) Handler {
		return func(context.Context, *Event) error {
			order = append(order, id)
			if fail {
				return errors.New("veto")
			}
			return nil
		}
	}
	_ = r.Register(OpCreate, PhaseBefore, "a", mk("a", false))
	_ = r.Register(OpCreate, PhaseBefore, "b", mk("b", true))
	_ = r.Register(OpCreate, PhaseBefore, "c", mk("c", false))

	err := r.TriggerBefore(context.Background(), OpCreate, &Event{Op: OpCreate})
	if err == nil || !strings.Contains(err.Error(), "b failed") {
		t.Fatalf("expected veto error from b, got %v", err)
	}
	if strings.Join(order, ",") != "a,b" {
		t.Fatalf("expected order a,b (c must not run), got %v", order)
	}

	// Repeat: identical order.
	order = nil
	_ = r.Register(OpUpdate, PhaseBefore, "a", mk("a", false))
	_ = r.Register(OpUpdate, PhaseBefore, "b", mk("b", false))
	_ = r.TriggerBefore(context.Background(), OpUpdate, &Event{Op: OpUpdate})
	if strings.Join(order, ",") != "a,b" {
		t.Fatalf("non-deterministic order: %v", order)
	}
}

func TestAfterCommitNeverBeforeFlush(t *testing.T) {
	var r Registry
	ran := false
	_ = r.Register(OpCreate, PhaseAfterCommit, "notify", func(context.Context, *Event) error {
		ran = true
		return nil
	})

	q := r.CollectAfterCommit(OpCreate, &Event{Op: OpCreate})
	if ran {
		t.Fatal("after-commit hook ran before Flush")
	}
	if err := q.Flush(context.Background()); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if !ran {
		t.Fatal("after-commit hook did not run on Flush")
	}
}

func TestAfterCommitEmptyQueue(t *testing.T) {
	var r Registry
	q := r.CollectAfterCommit(OpDelete, &Event{Op: OpDelete})
	if err := q.Flush(context.Background()); err != nil {
		t.Fatalf("empty flush: %v", err)
	}
	// Collect on zero registry must also work.
	var zero Registry
	q2 := zero.CollectAfterCommit(OpDelete, &Event{Op: OpDelete})
	if err := q2.Flush(context.Background()); err != nil {
		t.Fatalf("zero-value flush: %v", err)
	}
}

func TestReentrancyGuard(t *testing.T) {
	var r Registry
	var gotErr error
	_ = r.Register(OpCreate, PhaseBefore, "outer", func(ctx context.Context, ev *Event) error {
		gotErr = r.TriggerBefore(ctx, ev.Op, ev)
		return nil
	})
	err := r.TriggerBefore(context.Background(), OpCreate, &Event{Op: OpCreate})
	if err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if !errors.Is(gotErr, ErrReentrant) {
		t.Fatalf("expected ErrReentrant from nested trigger, got %v", gotErr)
	}
}

func TestContextCancellationPropagates(t *testing.T) {
	var r Registry
	ran := false
	_ = r.Register(OpCreate, PhaseBefore, "a", func(context.Context, *Event) error {
		ran = true
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := r.TriggerBefore(ctx, OpCreate, &Event{Op: OpCreate})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if ran {
		t.Fatal("hook must not run on canceled context")
	}
}

func TestFlushReentrancy(t *testing.T) {
	var r Registry
	var inner error
	_ = r.Register(OpCreate, PhaseAfter, "b", func(ctx context.Context, ev *Event) error {
		inner = r.CollectAfterCommit(ev.Op, ev).Flush(ctx)
		return nil
	})
	err := r.TriggerAfter(context.Background(), OpCreate, &Event{Op: OpCreate})
	if err != nil {
		t.Fatalf("trigger after: %v", err)
	}
	if !errors.Is(inner, ErrReentrant) {
		t.Fatalf("expected ErrReentrant flushing inside a hook, got %v", inner)
	}
}
