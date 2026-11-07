package manifest

import (
	"reflect"
	"strings"
	"testing"
)

func newM(name string, deps ...string) *Manifest {
	m := &Manifest{
		SchemaVersion: CurrentSchemaVersion,
		Name:          name,
		Version:       "1.0.0",
	}
	for _, d := range deps {
		m.Depends = append(m.Depends, Dependency{Name: d})
	}
	return m
}

func TestResolveOrder(t *testing.T) {
	tests := []struct {
		name    string
		modules []*Manifest
		want    []string
		wantErr string
	}{
		{
			name:    "empty graph",
			modules: nil,
			want:    []string{},
		},
		{
			name:    "linear chain",
			modules: []*Manifest{newM("a", "b"), newM("b", "c"), newM("c")},
			want:    []string{"c", "b", "a"},
		},
		{
			name:    "branching",
			modules: []*Manifest{newM("app", "core", "ui"), newM("core"), newM("ui", "core")},
			want:    []string{"core", "ui", "app"},
		},
		{
			name:    "missing dependency",
			modules: []*Manifest{newM("a", "missing")},
			wantErr: "unknown module",
		},
		{
			name:    "cycle",
			modules: []*Manifest{newM("a", "b"), newM("b", "a")},
			wantErr: "cycle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveOrder(tt.modules)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveOrderDeterministic(t *testing.T) {
	mk := func() []*Manifest {
		return []*Manifest{newM("a", "b"), newM("b", "c"), newM("c"), newM("d", "a")}
	}
	first, err := ResolveOrder(mk())
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	for i := 0; i < 5; i++ {
		got, err := ResolveOrder(mk())
		if err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
		if !reflect.DeepEqual(got, first) {
			t.Fatalf("non-deterministic output: run %d got %v, first %v", i, got, first)
		}
	}
}

func TestResolveOrderVersionIncompatibility(t *testing.T) {
	core := &Manifest{SchemaVersion: CurrentSchemaVersion, Name: "core", Version: "2.0.0"}
	app := newM("app", "core")
	app.Depends[0].Constraint = ">=1.0.0 <2.0.0"

	_, err := ResolveOrder([]*Manifest{app, core})
	if err == nil {
		t.Fatal("expected version incompatibility error")
	}
	if !strings.Contains(err.Error(), "requires") {
		t.Fatalf("expected actionable error, got %v", err)
	}
}
