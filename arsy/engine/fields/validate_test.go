package fields

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func ptr[T any](v T) *T { return &v }

func TestValidateAcceptsEverySupportedType(t *testing.T) {
	tests := []struct {
		name  string
		field Descriptor
	}{
		{"string", Descriptor{Name: "name", Type: TypeString, Label: "Name"}},
		{"text", Descriptor{Name: "description", Type: TypeText, Label: "Description"}},
		{"email", Descriptor{Name: "email", Type: TypeEmail, Label: "Email", Default: "user@example.com"}},
		{"url", Descriptor{Name: "website", Type: TypeURL, Label: "Website", Default: "https://example.com"}},
		{"boolean", Descriptor{Name: "active", Type: TypeBoolean, Label: "Active", Default: true}},
		{"integer", Descriptor{Name: "count", Type: TypeInteger, Label: "Count", Default: 3}},
		{"number", Descriptor{Name: "rating", Type: TypeNumber, Label: "Rating", Default: 4.5}},
		{"decimal", Descriptor{Name: "price", Type: TypeDecimal, Label: "Price", Default: "12.34", Options: &Options{Precision: ptr(8), Scale: ptr(2)}}},
		{"date", Descriptor{Name: "start_date", Type: TypeDate, Label: "Start date", Default: "2026-01-31"}},
		{"datetime", Descriptor{Name: "starts_at", Type: TypeDateTime, Label: "Starts at", Default: "2026-01-31T10:00:00Z"}},
		{"json", Descriptor{Name: "settings", Type: TypeJSON, Label: "Settings", Default: map[string]any{"theme": "dark"}}},
		{"select", Descriptor{Name: "status", Type: TypeSelect, Label: "Status", Default: "active", Options: &Options{Choices: []string{"active", "inactive"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Validate([]Descriptor{tt.field}, SourceMetadata)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if len(got) != 1 || got[0].Name != tt.field.Name {
				t.Fatalf("unexpected result: %#v", got)
			}
		})
	}
}

func TestValidateRejectsInvalidFields(t *testing.T) {
	tests := []struct {
		name   string
		fields []Descriptor
		source Source
		want   error
	}{
		{"unknown type", []Descriptor{{Name: "name", Type: "money", Label: "Name"}}, SourceMetadata, ErrInvalidFieldType},
		{"invalid identifier", []Descriptor{{Name: "Bad Name", Type: TypeString, Label: "Name"}}, SourceMetadata, ErrInvalidFieldName},
		{"duplicate identifier", []Descriptor{{Name: "name", Type: TypeString, Label: "Name"}, {Name: "name", Type: TypeText, Label: "Name"}}, SourceMetadata, ErrDuplicateField},
		{"empty label", []Descriptor{{Name: "name", Type: TypeString, Label: " "}}, SourceMetadata, ErrInvalidFieldLabel},
		{"invalid regex", []Descriptor{{Name: "code", Type: TypeString, Label: "Code", Options: &Options{Pattern: "["}}}, SourceMetadata, ErrInvalidFieldOption},
		{"incompatible options", []Descriptor{{Name: "active", Type: TypeBoolean, Label: "Active", Options: &Options{Min: ptr(1.0)}}}, SourceMetadata, ErrIncompatibleOption},
		{"invalid decimal precision", []Descriptor{{Name: "price", Type: TypeDecimal, Label: "Price", Options: &Options{Precision: ptr(2), Scale: ptr(3)}}}, SourceMetadata, ErrInvalidFieldOption},
		{"empty select", []Descriptor{{Name: "status", Type: TypeSelect, Label: "Status"}}, SourceMetadata, ErrInvalidFieldOption},
		{"invalid source", []Descriptor{{Name: "name", Type: TypeString, Label: "Name"}}, Source("remote"), ErrInvalidSource},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Validate(tt.fields, tt.source)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestUntrustedMetadataCannotDeclareSystemFieldsWritable(t *testing.T) {
	_, err := Validate([]Descriptor{{Name: "id", Type: TypeString, Label: "ID"}}, SourceMetadata)
	if !errors.Is(err, ErrWritableSystemField) {
		t.Fatalf("Validate() error = %v, want ErrWritableSystemField", err)
	}
	got, err := Validate([]Descriptor{{Name: "id", Type: TypeString, Label: "ID", ReadOnly: true}}, SourceMetadata)
	if err != nil {
		t.Fatalf("read-only system field rejected: %v", err)
	}
	if !got[0].ReadOnly || !got[0].SystemManaged {
		t.Fatalf("system field was not protected: %#v", got[0])
	}
}

func TestUniqueImpliesIndex(t *testing.T) {
	got, err := Validate([]Descriptor{{Name: "email", Type: TypeEmail, Label: "Email", Unique: true}}, SourceMetadata)
	if err != nil {
		t.Fatal(err)
	}
	if !got[0].Index {
		t.Fatal("unique field must normalize to indexed")
	}
}

func TestValidateNormalizesAndDetachesOptions(t *testing.T) {
	min, max := 2, 20
	input := []Descriptor{
		{Name: "zeta", Type: TypeString, Label: " Zeta ", Options: &Options{MinLength: &min, MaxLength: &max}},
		{Name: "alpha", Type: TypeString, Label: "Alpha"},
	}
	got, err := Validate(input, SourceStaticGo)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Name != "alpha" || got[1].Name != "zeta" {
		t.Fatalf("fields are not sorted: %#v", got)
	}
	*got[1].Options.MinLength = 99
	if *input[0].Options.MinLength != 2 {
		t.Fatal("normalized options alias caller memory")
	}
	if got[1].Label != "Zeta" {
		t.Fatalf("label not normalized: %q", got[1].Label)
	}
}

func TestCanonicalSerializationIsStable(t *testing.T) {
	a := []Descriptor{
		{Name: "zeta", Type: TypeJSON, Label: "Zeta", Default: map[string]any{"b": 2, "a": 1}},
		{Name: "alpha", Type: TypeBoolean, Label: "Alpha", Default: true},
	}
	b := []Descriptor{a[1], a[0]}
	first, err := MarshalCanonical(a, SourceMetadata)
	if err != nil {
		t.Fatal(err)
	}
	second, err := MarshalCanonical(b, SourceMetadata)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("serialization differs:\n%s\n%s", first, second)
	}
	var decoded []Descriptor
	if err := json.Unmarshal(first, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded[0].Name, "alpha") {
		t.Fatalf("unexpected canonical order: %s", first)
	}
}

func TestInvalidDefaultsAreRejected(t *testing.T) {
	tests := []Descriptor{
		{Name: "email", Type: TypeEmail, Label: "Email", Default: "not-an-email"},
		{Name: "url", Type: TypeURL, Label: "URL", Default: "/relative/path"},
		{Name: "count", Type: TypeInteger, Label: "Count", Default: 1.5},
		{Name: "active", Type: TypeBoolean, Label: "Active", Default: "true"},
		{Name: "status", Type: TypeSelect, Label: "Status", Default: "unknown", Options: &Options{Choices: []string{"active"}}},
		{Name: "day", Type: TypeDate, Label: "Day", Default: "31-01-2026"},
		{Name: "price", Type: TypeDecimal, Label: "Price", Default: "123.456", Options: &Options{Precision: ptr(5), Scale: ptr(2)}},
	}
	for _, field := range tests {
		t.Run(field.Name, func(t *testing.T) {
			_, err := Validate([]Descriptor{field}, SourceMetadata)
			if !errors.Is(err, ErrInvalidDefault) {
				t.Fatalf("Validate() error = %v, want ErrInvalidDefault", err)
			}
		})
	}
}

func TestStaticSystemFieldsAreAlwaysReadOnly(t *testing.T) {
	got, err := Validate([]Descriptor{{Name: "created", Type: TypeDateTime, Label: "Created"}}, SourceStaticGo)
	if err != nil {
		t.Fatal(err)
	}
	if !got[0].ReadOnly || !got[0].SystemManaged {
		t.Fatalf("system field is not protected: %#v", got[0])
	}
}
