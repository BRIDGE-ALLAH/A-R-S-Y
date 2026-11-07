package registry

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func validDescriptor(id ModelID, source ModelSource) Descriptor {
	return Descriptor{
		ID:          id,
		Label:       "Company",
		TableName:   "companies",
		Version:     "1.0.0",
		OwnerModule: "base",
		Source:      source,
	}
}

func TestRegisterAndLookupSupportsBothSources(t *testing.T) {
	for _, source := range []ModelSource{SourceStaticGo, SourceMetadata} {
		t.Run(string(source), func(t *testing.T) {
			r := New()
			want := validDescriptor("base.company", source)
			if err := r.Register(want); err != nil {
				t.Fatalf("Register() error = %v", err)
			}

			got, err := r.Lookup(want.ID)
			if err != nil {
				t.Fatalf("Lookup() error = %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("Lookup() = %#v, want %#v", got, want)
			}
		})
	}
}

func TestRegisterRejectsInvalidDescriptors(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Descriptor)
		want error
	}{
		{name: "invalid model ID", edit: func(d *Descriptor) { d.ID = "Company" }, want: ErrInvalidModelID},
		{name: "empty label", edit: func(d *Descriptor) { d.Label = "  " }, want: ErrInvalidLabel},
		{name: "invalid table", edit: func(d *Descriptor) { d.TableName = "Company; DROP TABLE users" }, want: ErrInvalidTableName},
		{name: "invalid version", edit: func(d *Descriptor) { d.Version = "01.0.0" }, want: ErrInvalidVersion},
		{name: "missing owner", edit: func(d *Descriptor) { d.OwnerModule = "" }, want: ErrMissingOwnerModule},
		{name: "invalid owner", edit: func(d *Descriptor) { d.OwnerModule = "../base" }, want: ErrInvalidModuleID},
		{name: "invalid source", edit: func(d *Descriptor) { d.Source = "generated" }, want: ErrInvalidSource},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := validDescriptor("base.company", SourceStaticGo)
			tt.edit(&d)
			err := New().Register(d)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Register() error = %v, want errors.Is(_, %v)", err, tt.want)
			}
			var typed *RegistrationError
			if !errors.As(err, &typed) {
				t.Fatalf("Register() error %T is not *RegistrationError", err)
			}
		})
	}
}

func TestRegisterDistinguishesDuplicateAndVersionConflict(t *testing.T) {
	r := New()
	d := validDescriptor("base.company", SourceStaticGo)
	if err := r.Register(d); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(d); !errors.Is(err, ErrDuplicateModel) {
		t.Fatalf("duplicate Register() error = %v", err)
	}

	changed := d
	changed.Version = "2.0.0"
	if err := r.Register(changed); !errors.Is(err, ErrModelVersionConflict) {
		t.Fatalf("conflicting version Register() error = %v", err)
	}
}

func TestLookupUnknownModelReturnsTypedError(t *testing.T) {
	_, err := New().Lookup("base.company")
	if !errors.Is(err, ErrUnknownModel) {
		t.Fatalf("Lookup() error = %v, want ErrUnknownModel", err)
	}
	var typed *RegistrationError
	if !errors.As(err, &typed) {
		t.Fatalf("Lookup() error %T is not *RegistrationError", err)
	}
}

func TestSnapshotIsDeterministicRegardlessOfRegistrationOrder(t *testing.T) {
	models := []Descriptor{
		validDescriptor("base.zebra", SourceStaticGo),
		validDescriptor("base.company", SourceMetadata),
		validDescriptor("base.account", SourceStaticGo),
	}
	first, second := New(), New()
	for _, d := range models {
		if err := first.Register(d); err != nil {
			t.Fatal(err)
		}
	}
	for i := len(models) - 1; i >= 0; i-- {
		if err := second.Register(models[i]); err != nil {
			t.Fatal(err)
		}
	}

	got, want := first.Snapshot(), second.Snapshot()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("snapshots differ by registration order:\n got %#v\nwant %#v", got, want)
	}
	if got[0].ID != "base.account" || got[1].ID != "base.company" || got[2].ID != "base.zebra" {
		t.Fatalf("snapshot is not sorted by model ID: %#v", got)
	}
}

func TestSnapshotDoesNotExposeRegistryStorage(t *testing.T) {
	r := New()
	d := validDescriptor("base.company", SourceStaticGo)
	if err := r.Register(d); err != nil {
		t.Fatal(err)
	}

	snapshot := r.Snapshot()
	snapshot[0].Label = "mutated"
	got, err := r.Lookup(d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Label != d.Label {
		t.Fatalf("snapshot mutation changed registry label: got %q, want %q", got.Label, d.Label)
	}
}

func TestConcurrentRegistrationAndLookup(t *testing.T) {
	const count = 64
	r := New()
	var wg sync.WaitGroup
	errs := make(chan error, count)

	for i := 0; i < count; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := ModelID(fmt.Sprintf("base.model_%02d", i))
			if err := r.Register(validDescriptor(id, SourceStaticGo)); err != nil {
				errs <- err
				return
			}
			if _, err := r.Lookup(id); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	if got := len(r.Snapshot()); got != count {
		t.Fatalf("Snapshot() contains %d models, want %d", got, count)
	}
}
