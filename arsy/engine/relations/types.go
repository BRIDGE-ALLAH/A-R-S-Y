package relations

import "github.com/bridge-allah/a-r-s-y/arsy/engine/registry"

// Cardinality identifies the direction and multiplicity of a relation.
type Cardinality string

const (
	ManyToOne  Cardinality = "many_to_one"
	OneToMany  Cardinality = "one_to_many"
	ManyToMany Cardinality = "many_to_many"
)

// DeleteBehavior is the explicit action taken when a referenced record is deleted.
type DeleteBehavior string

const (
	DeleteRestrict DeleteBehavior = "restrict"
	DeleteSetNull  DeleteBehavior = "set_null"
	DeleteCascade  DeleteBehavior = "cascade"
)

// Descriptor defines one named relation field on a source model.
// InverseField names the reciprocal relation field on the target model.
type Descriptor struct {
	Source       registry.ModelID `json:"source"`
	Name         string           `json:"name"`
	Cardinality  Cardinality      `json:"cardinality"`
	Target       registry.ModelID `json:"target"`
	InverseField string           `json:"inverse_field,omitempty"`
	OnDelete     DeleteBehavior   `json:"on_delete"`
	Required     bool             `json:"required,omitempty"`
	JoinTable    string           `json:"join_table,omitempty"`
	// JoinTableOwner is derived: the lexicographically first relation endpoint owns the join table.
	JoinTableOwner bool `json:"join_table_owner,omitempty"`
}
