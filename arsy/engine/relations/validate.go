package relations

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"

	"github.com/bridge-allah/a-r-s-y/arsy/engine/registry"
)

var relationNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func relationError(d Descriptor, rule error, detail string) error {
	return &ValidationError{
		Source: d.Source,
		Name:   d.Name,
		Rule:   rule,
		Detail: detail,
	}
}

type relationKey struct {
	model registry.ModelID
	name  string
}

// Validate validates a complete relation graph against registered models.
// Inverse references are resolved only within the supplied graph; no database
// joins, foreign keys, or schema mutations are performed.
func Validate(input []Descriptor, models registry.ModelRegistry) ([]Descriptor, error) {
	if models == nil {
		return nil, &ValidationError{Rule: ErrInvalidRelationSet, Detail: "model registry is required"}
	}

	byKey := make(map[relationKey]Descriptor, len(input))
	out := make([]Descriptor, len(input))
	copy(out, input)

	for _, d := range input {
		if !relationNamePattern.MatchString(d.Name) {
			return nil, relationError(d, ErrInvalidRelationName, "use lowercase letters, digits, and underscores; the first character must be a letter")
		}
		if _, err := models.Lookup(d.Source); err != nil {
			return nil, relationError(d, ErrInvalidSourceModel, "source model is not registered")
		}
		if _, err := models.Lookup(d.Target); err != nil {
			return nil, relationError(d, ErrInvalidTargetModel, "target model is not registered")
		}
		key := relationKey{model: d.Source, name: d.Name}
		if _, exists := byKey[key]; exists {
			return nil, relationError(d, ErrDuplicateRelation, "relation names must be unique within a source model")
		}
		byKey[key] = d
		if err := validateLocal(d); err != nil {
			return nil, err
		}
	}

	for i, d := range out {
		if d.InverseField == "" {
			if d.Cardinality == OneToMany || d.Cardinality == ManyToMany {
				return nil, relationError(d, ErrInvalidInverse, "this cardinality requires an inverse_field")
			}
			continue
		}

		inverse, ok := byKey[relationKey{model: d.Target, name: d.InverseField}]
		if !ok {
			return nil, relationError(d, ErrInvalidInverse, "inverse field does not exist on the target model")
		}
		if inverse.Target != d.Source || inverse.InverseField != d.Name {
			return nil, relationError(d, ErrInvalidInverse, "inverse relation must point back to the source model and field")
		}
		if !compatibleCardinality(d.Cardinality, inverse.Cardinality) {
			return nil, relationError(d, ErrInvalidCardinality, "inverse relation has an incompatible cardinality")
		}
		if inverse.OnDelete != d.OnDelete {
			return nil, relationError(d, ErrInvalidDeletePolicy, "both sides of a relation must declare the same delete behavior")
		}
		if d.Cardinality == ManyToMany {
			join := joinTableName(d.Source, d.Name, d.Target, d.InverseField)
			if d.JoinTable != "" && d.JoinTable != join {
				return nil, relationError(d, ErrInvalidJoinTable, fmt.Sprintf("join_table must be the deterministic name %q", join))
			}
			out[i].JoinTable = join
		}
	}

	// Normalize the inverse many-to-many side to the same deterministic table.
	for i := range out {
		d := out[i]
		out[i].JoinTableOwner = false
		if d.Cardinality != ManyToMany {
			continue
		}
		out[i].JoinTable = joinTableName(d.Source, d.Name, d.Target, d.InverseField)
		out[i].JoinTableOwner = ownsJoinTable(d.Source, d.Name, d.Target, d.InverseField)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Source == out[j].Source {
			return out[i].Name < out[j].Name
		}
		return out[i].Source < out[j].Source
	})
	return out, nil
}

func validateLocal(d Descriptor) error {
	switch d.Cardinality {
	case ManyToOne, OneToMany, ManyToMany:
	default:
		return relationError(d, ErrInvalidCardinality, "cardinality must be many_to_one, one_to_many, or many_to_many")
	}

	switch d.OnDelete {
	case DeleteRestrict, DeleteSetNull, DeleteCascade:
	default:
		return relationError(d, ErrInvalidDeletePolicy, "on_delete must explicitly be restrict, set_null, or cascade")
	}

	if d.Cardinality == ManyToOne && d.OnDelete == DeleteSetNull && d.Required {
		return relationError(d, ErrInvalidDeletePolicy, "set_null is incompatible with a required many-to-one relation")
	}
	if d.Cardinality == ManyToMany && d.OnDelete == DeleteSetNull {
		return relationError(d, ErrInvalidDeletePolicy, "set_null is not valid for many-to-many relations")
	}
	if d.Cardinality != ManyToMany && d.JoinTable != "" {
		return relationError(d, ErrInvalidJoinTable, "join_table is only valid for many-to-many relations")
	}
	return nil
}

func compatibleCardinality(a, b Cardinality) bool {
	return (a == ManyToOne && b == OneToMany) ||
		(a == OneToMany && b == ManyToOne) ||
		(a == ManyToMany && b == ManyToMany)
}

func joinTableName(source registry.ModelID, sourceField string, target registry.ModelID, targetField string) string {
	left := string(source) + "." + sourceField
	right := string(target) + "." + targetField
	if right < left {
		left, right = right, left
	}
	sum := sha256.Sum256([]byte(left + "|" + right))
	// A short fixed-width digest keeps the identifier portable across SQL
	// engines and independent of model or field name length.
	return "rel_" + hex.EncodeToString(sum[:6])
}

// JoinTableName exposes the canonical join-table name for tooling and tests.
func JoinTableName(source registry.ModelID, sourceField string, target registry.ModelID, targetField string) string {
	return joinTableName(source, sourceField, target, targetField)
}

// ownsJoinTable assigns ownership to the lexicographically first endpoint.
// Both inverse descriptors compute the same owner independently.
func ownsJoinTable(source registry.ModelID, sourceField string, target registry.ModelID, targetField string) bool {
	left := string(source) + "." + sourceField
	right := string(target) + "." + targetField
	return left < right
}
