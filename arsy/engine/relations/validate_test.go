package relations

import (
	"errors"
	"testing"

	"github.com/bridge-allah/a-r-s-y/arsy/engine/registry"
)

func testRegistry(t *testing.T) *registry.InMemoryRegistry {
	t.Helper()
	r := registry.New()
	for _, d := range []registry.Descriptor{
		{ID: "base.user", Label: "User", TableName: "users", Version: "1.0.0", OwnerModule: "base", Source: registry.SourceStaticGo},
		{ID: "base.company", Label: "Company", TableName: "companies", Version: "1.0.0", OwnerModule: "base", Source: registry.SourceStaticGo},
		{ID: "base.team", Label: "Team", TableName: "teams", Version: "1.0.0", OwnerModule: "base", Source: registry.SourceStaticGo},
	} {
		if err := r.Register(d); err != nil {
			t.Fatalf("Register(%s): %v", d.ID, err)
		}
	}
	return r
}

func TestValidateResolvesManyToOneAndInverseOneToMany(t *testing.T) {
	models := testRegistry(t)
	input := []Descriptor{
		{Source: "base.user", Name: "company", Cardinality: ManyToOne, Target: "base.company", InverseField: "users", OnDelete: DeleteRestrict},
		{Source: "base.company", Name: "users", Cardinality: OneToMany, Target: "base.user", InverseField: "company", OnDelete: DeleteRestrict},
	}

	got, err := Validate(input, models)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(got) != 2 || got[0].Source != "base.company" || got[1].Source != "base.user" {
		t.Fatalf("relations are not deterministically sorted: %#v", got)
	}
}

func TestValidateRejectsMissingTargetAndInverse(t *testing.T) {
	models := testRegistry(t)
	tests := []struct {
		name string
		in   []Descriptor
		want error
	}{
		{
			name: "missing source",
			in:   []Descriptor{{Source: "base.missing", Name: "company", Cardinality: ManyToOne, Target: "base.company", OnDelete: DeleteRestrict}},
			want: ErrInvalidSourceModel,
		},
		{
			name: "missing target",
			in:   []Descriptor{{Source: "base.user", Name: "company", Cardinality: ManyToOne, Target: "base.missing", OnDelete: DeleteRestrict}},
			want: ErrInvalidTargetModel,
		},
		{
			name: "invalid relation name",
			in:   []Descriptor{{Source: "base.user", Name: "Bad Name", Cardinality: ManyToOne, Target: "base.company", OnDelete: DeleteRestrict}},
			want: ErrInvalidRelationName,
		},
		{
			name: "missing inverse",
			in:   []Descriptor{{Source: "base.company", Name: "users", Cardinality: OneToMany, Target: "base.user", InverseField: "missing", OnDelete: DeleteRestrict}},
			want: ErrInvalidInverse,
		},
		{
			name: "inverse points to wrong model",
			in: []Descriptor{
				{Source: "base.user", Name: "company", Cardinality: ManyToOne, Target: "base.company", InverseField: "users", OnDelete: DeleteRestrict},
				{Source: "base.company", Name: "users", Cardinality: OneToMany, Target: "base.team", InverseField: "company", OnDelete: DeleteRestrict},
			},
			want: ErrInvalidInverse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Validate(tt.in, models)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestValidateRejectsDuplicateRelationAndCardinalityMismatch(t *testing.T) {
	models := testRegistry(t)
	tests := []struct {
		name string
		in   []Descriptor
		want error
	}{
		{
			name: "duplicate field",
			in: []Descriptor{
				{Source: "base.user", Name: "company", Cardinality: ManyToOne, Target: "base.company", OnDelete: DeleteRestrict},
				{Source: "base.user", Name: "company", Cardinality: ManyToOne, Target: "base.team", OnDelete: DeleteRestrict},
			},
			want: ErrDuplicateRelation,
		},
		{
			name: "incompatible inverse cardinality",
			in: []Descriptor{
				{Source: "base.user", Name: "company", Cardinality: ManyToOne, Target: "base.company", InverseField: "users", OnDelete: DeleteRestrict},
				{Source: "base.company", Name: "users", Cardinality: ManyToMany, Target: "base.user", InverseField: "company", OnDelete: DeleteRestrict},
			},
			want: ErrInvalidCardinality,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Validate(tt.in, models)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestDeleteBehaviorMustBeExplicitAndSafe(t *testing.T) {
	models := testRegistry(t)
	tests := []struct {
		name string
		in   Descriptor
		want error
	}{
		{
			name: "missing behavior",
			in:   Descriptor{Source: "base.user", Name: "company", Cardinality: ManyToOne, Target: "base.company"},
			want: ErrInvalidDeletePolicy,
		},
		{
			name: "required relation cannot set null",
			in:   Descriptor{Source: "base.user", Name: "company", Cardinality: ManyToOne, Target: "base.company", OnDelete: DeleteSetNull, Required: true},
			want: ErrInvalidDeletePolicy,
		},
		{
			name: "many to many cannot set null",
			in:   Descriptor{Source: "base.user", Name: "teams", Cardinality: ManyToMany, Target: "base.team", OnDelete: DeleteSetNull},
			want: ErrInvalidDeletePolicy,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Validate([]Descriptor{tt.in}, models)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestManyToManyUsesStableSharedJoinTable(t *testing.T) {
	models := testRegistry(t)
	input := []Descriptor{
		{Source: "base.user", Name: "teams", Cardinality: ManyToMany, Target: "base.team", InverseField: "members", OnDelete: DeleteCascade},
		{Source: "base.team", Name: "members", Cardinality: ManyToMany, Target: "base.user", InverseField: "teams", OnDelete: DeleteCascade},
	}
	first, err := Validate(input, models)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	second, err := Validate([]Descriptor{input[1], input[0]}, models)
	if err != nil {
		t.Fatalf("Validate() with reversed input error = %v", err)
	}
	if first[0].JoinTable == "" || first[0].JoinTable != first[1].JoinTable {
		t.Fatalf("inverse relations do not share a join table: %#v", first)
	}
	if first[0].JoinTableOwner == first[1].JoinTableOwner || !first[0].JoinTableOwner {
		t.Fatalf("join table must have exactly one deterministic owner: %#v", first)
	}
	if first[0].JoinTable != second[0].JoinTable || first[1].JoinTable != second[1].JoinTable {
		t.Fatalf("join table depends on input order: %#v vs %#v", first, second)
	}
	if got := JoinTableName("base.user", "teams", "base.team", "members"); got != first[0].JoinTable {
		t.Fatalf("JoinTableName() = %q, want %q", got, first[0].JoinTable)
	}
}

func TestManyToManyRejectsCallerDefinedJoinTable(t *testing.T) {
	models := testRegistry(t)
	input := []Descriptor{
		{Source: "base.user", Name: "teams", Cardinality: ManyToMany, Target: "base.team", InverseField: "members", OnDelete: DeleteRestrict, JoinTable: "custom_join"},
		{Source: "base.team", Name: "members", Cardinality: ManyToMany, Target: "base.user", InverseField: "teams", OnDelete: DeleteRestrict},
	}
	_, err := Validate(input, models)
	if !errors.Is(err, ErrInvalidJoinTable) {
		t.Fatalf("Validate() error = %v, want ErrInvalidJoinTable", err)
	}
}

func TestSelfRelationGraphIsValidAndDeterministic(t *testing.T) {
	models := testRegistry(t)
	input := []Descriptor{
		{Source: "base.user", Name: "manager", Cardinality: ManyToOne, Target: "base.user", InverseField: "reports", OnDelete: DeleteSetNull},
		{Source: "base.user", Name: "reports", Cardinality: OneToMany, Target: "base.user", InverseField: "manager", OnDelete: DeleteSetNull},
	}
	got, err := Validate(input, models)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if got[0].Name != "manager" || got[1].Name != "reports" {
		t.Fatalf("self relation ordering is unstable: %#v", got)
	}
}
